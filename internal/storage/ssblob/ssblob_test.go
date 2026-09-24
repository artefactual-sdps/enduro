package ssblob_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/go-cmp/cmp/cmpopts"
	"go.artefactual.dev/ssclient"
	"gocloud.dev/blob"
	"gocloud.dev/blob/driver"
	"gocloud.dev/gcerrors"
	"gotest.tools/v3/assert"

	"github.com/artefactual-sdps/enduro/internal/storage/ssblob"
)

// setUpTest creates a ssblob bucket configured against a fake SS API.
func setUpTest(t *testing.T, h http.HandlerFunc, opts *ssblob.Options) *blob.Bucket {
	t.Helper()

	if opts == nil {
		opts = &ssblob.Options{}
	}

	srv := httptest.NewServer(h)
	t.Cleanup(func() { srv.Close() })
	opts.URL = srv.URL

	b, err := ssblob.OpenBucket(opts)
	assert.NilError(t, err)
	t.Cleanup(func() { b.Close() })

	return b
}

func TestAPIErrorError(t *testing.T) {
	t.Parallel()

	type testCase struct {
		name string
		err  *ssblob.APIError
		want string
	}
	for _, tc := range []testCase{
		{
			name: "Error returns <nil>",
			want: "<nil>",
		},
		{
			name: "Error returns status",
			err:  &ssblob.APIError{Status: "Not Found", Code: http.StatusNotFound},
			want: "Not Found",
		},
		{
			name: "Error returns underlying cause error",
			err:  &ssblob.APIError{Cause: errors.New("underlying cause")},
			want: "underlying cause",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tc.err.Error(), tc.want)
		})
	}
}

func TestAPIErrorUnwrap(t *testing.T) {
	t.Parallel()

	wrappedErr := errors.New("underlying cause")

	type testCase struct {
		name string
		err  *ssblob.APIError
		want error
	}
	for _, tc := range []testCase{
		{
			name: "Unwrap nil error",
			want: nil,
		},
		{
			name: "Unwrap underlying error",
			err:  &ssblob.APIError{Cause: wrappedErr},
			want: wrappedErr,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, errors.Unwrap(tc.err), tc.want)
		})
	}
}

func TestOpenBucket(t *testing.T) {
	t.Parallel()

	t.Run("OpenBucket error when the URL is invalid", func(t *testing.T) {
		t.Parallel()

		b, err := ssblob.OpenBucket(&ssblob.Options{
			URL: string([]byte{0x7f}), // DEL character is rejected.
		})
		assert.ErrorContains(t, err, "net/url: invalid control character in URL")
		assert.Assert(t, b == nil)
	})
}

func TestBucketAs(t *testing.T) {
	t.Parallel()

	t.Run("Exposes the underlying Storage Service client", func(t *testing.T) {
		t.Parallel()

		b := setUpTest(t, nil, nil)

		// Supported type.
		var client *ssclient.Client
		assert.Equal(t, b.As(&client), true)
		assert.Assert(t, client != nil)

		// Unsupported type.
		var s string
		assert.Equal(t, b.As(&s), false)
	})
}

func TestBucketErrorAs(t *testing.T) {
	t.Parallel()

	type testCase struct {
		name    string
		wrapped error
		toType  any
		want    bool
	}
	for _, tc := range []testCase{
		{
			name: "ErrorAs returns true when casting wrapped to APIError",
			wrapped: fmt.Errorf("wrapped: %w", &ssblob.APIError{
				Status: http.StatusText(http.StatusNotFound),
				Code:   http.StatusNotFound,
			}),
			toType: &ssblob.APIError{},
			want:   true,
		},
		{
			name:    "ErrorAs returns false if no APIError is in the error chain",
			wrapped: fmt.Errorf("wrapped: %w", errors.New("some other error")),
			toType:  &ssblob.APIError{},
		},
		{
			name: "ErrorAs returns false if casting to a type other than APIError",
			wrapped: fmt.Errorf("wrapped: %w", &ssblob.APIError{
				Status: http.StatusText(http.StatusNotFound),
				Code:   http.StatusNotFound,
			}),
			toType: "",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			b := setUpTest(t, nil, nil)
			switch i := tc.toType.(type) {
			case *ssblob.APIError:
				assert.Equal(t, b.ErrorAs(tc.wrapped, &i), tc.want)
			case string:
				assert.Equal(t, b.ErrorAs(tc.wrapped, &i), false)
			default:
				t.Fatalf("unsupported ErrorAs type: %T", tc.toType)
			}
		})
	}
}

func TestBucketAttributes(t *testing.T) {
	t.Parallel()

	aipID := "2db707f3-3cd2-44b7-9012-9b68eb10d207"

	type testCase struct {
		name        string
		httpHandler http.HandlerFunc
		want        *blob.Attributes
		wantErr     *ssblob.APIError
	}
	for _, tc := range []testCase{
		{
			name: "Attributes returns AIP attributes",
			httpHandler: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.Header().Set("Content-Disposition", "inline")
				err := json.NewEncoder(w).Encode(map[string]any{
					"package_type": "AIP",
					"size":         12345,
					"status":       "UPLOADED",
					"uuid":         aipID,
				})

				assert.NilError(t, err)
				assert.Equal(t, r.URL.Path, fmt.Sprintf("/api/v2/file/%s/", aipID))
			},
			want: &blob.Attributes{
				ContentDisposition: "attachment",
				ContentType:        "application/octet-stream",
				Size:               12345,
			},
		},
		{
			name: "Attributes returns error when blob is not found",
			httpHandler: func(w http.ResponseWriter, r *http.Request) {
				http.Error(w, "AIP not found.", http.StatusNotFound)
			},
			want: nil,
			wantErr: &ssblob.APIError{
				Code:   http.StatusNotFound,
				Status: http.StatusText(http.StatusNotFound),
			},
		},
		{
			name: "Attributes returns error when AIP is deleted",
			httpHandler: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.Header().Set("Content-Disposition", "inline")
				w.WriteHeader(http.StatusOK)
				err := json.NewEncoder(w).Encode(map[string]any{
					"package_type": "AIP",
					"size":         12345,
					"status":       "DELETED",
					"uuid":         aipID,
				})
				assert.NilError(t, err)
			},
			want: nil,
			wantErr: &ssblob.APIError{
				Code:   http.StatusNotFound,
				Status: http.StatusText(http.StatusNotFound),
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			b := setUpTest(t, tc.httpHandler, nil)
			attrs, err := b.Attributes(context.Background(), aipID)

			assert.DeepEqual(t, attrs, tc.want, cmpopts.IgnoreUnexported(blob.Attributes{}))
			if tc.wantErr != nil {
				apiErr, ok := errors.AsType[*ssblob.APIError](err)
				assert.Assert(t, ok)
				assert.Equal(t, apiErr.Code, tc.wantErr.Code)
				assert.Equal(t, apiErr.Status, tc.wantErr.Status)
			} else {
				assert.NilError(t, err)
			}
		})
	}
}

func TestBucketNewRangeReader(t *testing.T) {
	t.Parallel()

	type testCase struct {
		name    string
		offset  int64
		length  int64
		wantErr *ssblob.APIError
	}
	for _, tc := range []testCase{
		{
			name:    "Rejects offset ranged reads",
			offset:  1,
			length:  10,
			wantErr: &ssblob.APIError{},
		},
		{
			name:    "Rejects zero-length ranged reads",
			offset:  0,
			length:  0,
			wantErr: &ssblob.APIError{},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			b := setUpTest(t, func(w http.ResponseWriter, r *http.Request) {
				t.Fatal("unexpected request")
			}, nil)

			r, err := b.NewRangeReader(
				context.Background(),
				"2db707f3-3cd2-44b7-9012-9b68eb10d207",
				tc.offset,
				tc.length,
				nil,
			)
			assert.Equal(t, gcerrors.Code(err), gcerrors.Unimplemented)
			assert.Assert(t, r == nil)
		})
	}
}

func TestBucketNewReader(t *testing.T) {
	t.Parallel()

	type testCase struct {
		name        string
		httpHandler http.HandlerFunc
		want        []byte
		wantAttrs   driver.ReaderAttributes
		wantAPIErr  *ssblob.APIError
	}
	for _, tc := range []testCase{
		{
			name: "Returns file reader",
			httpHandler: func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, r.Header.Get("Authorization"), "ApiKey test:test")
				assert.Equal(t, r.URL.Path, "/api/v2/file/2db707f3-3cd2-44b7-9012-9b68eb10d207/download/")

				w.Header().Set("Content-Type", "text/plain")
				w.Header().Set("Content-Disposition", "attachment; filename=\"hello.txt\"")
				_, err := w.Write([]byte("Hello World!"))
				assert.NilError(t, err)
			},
			wantAttrs: driver.ReaderAttributes{
				ContentType: "text/plain",
				Size:        int64(len("Hello World!")),
			},
			want: []byte("Hello World!"),
		},
		{
			name: "Returns an error if the package is unavailable",
			httpHandler: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusAccepted)
				err := fmt.Errorf(`{"message":"package is not available"}`)
				_, writeErr := w.Write([]byte(err.Error()))
				assert.NilError(t, writeErr)
			},
			wantAPIErr: &ssblob.APIError{
				Code:   http.StatusAccepted,
				Status: http.StatusText(http.StatusAccepted),
			},
		},
		{
			name: "401 Unauthorized error",
			httpHandler: func(w http.ResponseWriter, r *http.Request) {
				http.Error(w, "User is unauthorized.", http.StatusUnauthorized)
			},
			wantAPIErr: &ssblob.APIError{
				Code:   http.StatusUnauthorized,
				Status: http.StatusText(http.StatusUnauthorized),
			},
		},
		{
			name: "403 Forbidden error",
			httpHandler: func(w http.ResponseWriter, r *http.Request) {
				http.Error(w, "Forbidden", http.StatusForbidden)
			},
			wantAPIErr: &ssblob.APIError{
				Code:   http.StatusForbidden,
				Status: http.StatusText(http.StatusForbidden),
			},
		},
		{
			name: "404 Not found error",
			httpHandler: func(w http.ResponseWriter, r *http.Request) {
				http.Error(w, "AIP not found.", http.StatusNotFound)
			},
			wantAPIErr: &ssblob.APIError{
				Code:   http.StatusNotFound,
				Status: http.StatusText(http.StatusNotFound),
			},
		},
		{
			name: "500 Internal server error",
			httpHandler: func(w http.ResponseWriter, r *http.Request) {
				http.Error(w, "Internal server error.", http.StatusInternalServerError)
			},
			wantAPIErr: &ssblob.APIError{
				Code:   http.StatusInternalServerError,
				Status: http.StatusText(http.StatusInternalServerError),
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			b := setUpTest(t,
				tc.httpHandler,
				&ssblob.Options{
					Username: "test",
					Key:      "test",
				},
			)
			r, err := b.NewReader(context.Background(), "2db707f3-3cd2-44b7-9012-9b68eb10d207", nil)
			if tc.wantAPIErr != nil {
				gotErr := &ssblob.APIError{}
				assert.Equal(t, b.ErrorAs(err, &gotErr), true)
				assert.Equal(t, gotErr.Code, tc.wantAPIErr.Code)
				assert.Equal(t, gotErr.Status, tc.wantAPIErr.Status)
				assert.Assert(t, r == nil)
				return
			}
			defer r.Close()

			assert.Equal(t, r.Size(), tc.wantAttrs.Size)
			assert.Equal(t, r.ContentType(), tc.wantAttrs.ContentType)

			got, readErr := io.ReadAll(r)
			assert.NilError(t, readErr)
			assert.DeepEqual(t, got, tc.want)
		})
	}
}

func TestBucketUnsupportedMethods(t *testing.T) {
	t.Parallel()

	t.Run("Bucket rejects unsupported methods", func(t *testing.T) {
		t.Parallel()

		b := setUpTest(t, func(w http.ResponseWriter, r *http.Request) {
			t.Fatal("unexpected request")
		}, nil)

		ctx := t.Context()

		objs, nextPageToken, err := b.ListPage(ctx, nil, 10, nil)
		assert.Assert(t, objs == nil)
		assert.Assert(t, nextPageToken == nil)
		assert.Assert(t, err != nil)

		iter := b.List(nil)
		obj, err := iter.Next(ctx)
		assert.Assert(t, obj == nil)
		assert.Assert(t, err != nil)

		err = b.WriteAll(ctx, "2db707f3-3cd2-44b7-9012-9b68eb10d207", []byte("hello"), nil)
		assert.Equal(t, gcerrors.Code(err), gcerrors.Unimplemented)

		err = b.Copy(ctx, "dst", "src", nil)
		assert.Equal(t, gcerrors.Code(err), gcerrors.Unimplemented)

		err = b.Delete(ctx, "2db707f3-3cd2-44b7-9012-9b68eb10d207")
		assert.Equal(t, gcerrors.Code(err), gcerrors.Unimplemented)

		signedURL, err := b.SignedURL(ctx, "2db707f3-3cd2-44b7-9012-9b68eb10d207", nil)
		assert.Equal(t, signedURL, "")
		assert.Equal(t, gcerrors.Code(err), gcerrors.Unimplemented)
	})
}

func TestReaderAs(t *testing.T) {
	t.Parallel()

	t.Run("Exposes the underlying ssclient file stream", func(t *testing.T) {
		t.Parallel()

		b := setUpTest(t,
			func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "text/plain")
				_, err := w.Write([]byte("Hello World!"))
				assert.NilError(t, err)
			},
			nil,
		)

		r, err := b.NewReader(context.Background(), "2db707f3-3cd2-44b7-9012-9b68eb10d207", nil)
		assert.NilError(t, err)
		defer r.Close()

		var stream *ssclient.FileStream
		assert.Equal(t, r.As(&stream), true)
		assert.Equal(t, stream.ContentType, "text/plain")
	})

	t.Run("Returns false using As with an unsupported type", func(t *testing.T) {
		t.Parallel()

		b := setUpTest(t,
			func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "text/plain")
				_, err := w.Write([]byte("Hello World!"))
				assert.NilError(t, err)
			},
			nil,
		)

		r, err := b.NewReader(context.Background(), "2db707f3-3cd2-44b7-9012-9b68eb10d207", nil)
		assert.NilError(t, err)
		defer r.Close()

		var s string
		assert.Equal(t, r.As(&s), false)
	})
}
