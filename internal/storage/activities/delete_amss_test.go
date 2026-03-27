package activities_test

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/hashicorp/go-cleanhttp"
	temporalsdk_activity "go.temporal.io/sdk/activity"
	temporalsdk_testsuite "go.temporal.io/sdk/testsuite"
	"gotest.tools/v3/assert"

	"github.com/artefactual-sdps/enduro/internal/storage"
	"github.com/artefactual-sdps/enduro/internal/storage/activities"
	"github.com/artefactual-sdps/enduro/internal/storage/types"
)

func setupServer(t *testing.T, h http.HandlerFunc) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(h)
	t.Cleanup(func() { srv.Close() })
	return srv
}

func TestDeleteFromAMSSLocationActivity(t *testing.T) {
	t.Parallel()

	aipUUID := "2db707f3-3cd2-44b7-9012-9b68eb10d207"
	pipelineUUID := "a1b2c3d4-e5f6-4321-9876-abcdef123456"

	// Common happy path handlers.
	handleAIPInfo := func(w http.ResponseWriter, r *http.Request, s string) {
		// Validate request.
		assert.Equal(t, r.Header.Get("Authorization"), "ApiKey test:test")
		assert.Equal(t, r.URL.Path, "/api/v2/file/"+aipUUID+"/")

		// Fake response.
		w.Header().Set("Content-Type", "application/json")
		resp := map[string]any{
			"origin_pipeline": "/api/v2/pipeline/" + pipelineUUID + "/",
			"status":          s,
		}
		err := json.NewEncoder(w).Encode(resp)
		assert.NilError(t, err)
	}
	handleRequest := func(w http.ResponseWriter, r *http.Request) {
		// Validate request.
		assert.Equal(t, r.Header.Get("Authorization"), "ApiKey test:test")
		assert.Equal(t, r.URL.Path, "/api/v2/file/"+aipUUID+"/delete_aip/")
		type deletionReq struct {
			EventReason string `json:"event_reason"`
			Pipeline    string `json:"pipeline"`
			UserID      int    `json:"user_id"`
			UserEmail   string `json:"user_email"`
		}
		body, err := io.ReadAll(r.Body)
		assert.NilError(t, err)
		var got deletionReq
		err = json.Unmarshal(body, &got)
		assert.NilError(t, err)
		expected := deletionReq{
			EventReason: "Deletion from Enduro",
			Pipeline:    pipelineUUID,
			UserID:      123,
			UserEmail:   "enduro@example.com",
		}
		assert.DeepEqual(t, got, expected)

		// Fake response.
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusAccepted)
		err = json.NewEncoder(w).Encode(map[string]any{"id": int64(42)})
		assert.NilError(t, err)
	}
	handleExistingRequest := func(w http.ResponseWriter, r *http.Request) {
		// Validate request.
		assert.Equal(t, r.Header.Get("Authorization"), "ApiKey test:test")
		assert.Equal(t, r.URL.Path, "/api/v2/file/"+aipUUID+"/delete_aip/")

		// Fake response.
		w.Header().Set("Content-Type", "application/json")
		err := json.NewEncoder(w).Encode(map[string]any{
			"error_message": "A deletion request already exists for this AIP.",
		})
		assert.NilError(t, err)
	}
	handleReview := func(w http.ResponseWriter, r *http.Request) {
		// Validate request.
		assert.Equal(t, r.Header.Get("Authorization"), "ApiKey test:test")
		assert.Equal(t, r.URL.Path, "/api/v2/file/"+aipUUID+"/review_aip_deletion/")
		type approvalReq struct {
			EventID  int64  `json:"event_id"`
			Decision string `json:"decision"`
			Reason   string `json:"reason"`
		}
		body, err := io.ReadAll(r.Body)
		assert.NilError(t, err)
		var got approvalReq
		err = json.Unmarshal(body, &got)
		assert.NilError(t, err)
		expected := approvalReq{
			EventID:  42,
			Decision: "approve",
			Reason:   "Approval from Enduro",
		}
		assert.DeepEqual(t, got, expected)

		// Fake response.
		w.Header().Set("Content-Type", "application/json")
		err = json.NewEncoder(w).Encode(map[string]any{"message": "Deleted"})
		assert.NilError(t, err)
	}

	for _, tt := range []struct {
		name    string
		approve bool
		url     string
		poll    time.Duration
		handler http.HandlerFunc
		want    activities.DeleteFromAMSSLocationActivityResult
		wantErr string
	}{
		{
			name:    "Deletes AIP with approval",
			approve: true,
			handler: func(w http.ResponseWriter, r *http.Request) {
				switch r.URL.Path {
				case "/api/v2/file/" + aipUUID + "/":
					handleAIPInfo(w, r, "UPLOADED")
				case "/api/v2/file/" + aipUUID + "/delete_aip/":
					handleRequest(w, r)
				case "/api/v2/file/" + aipUUID + "/review_aip_deletion/":
					handleReview(w, r)
				default:
					t.Fatalf("unexpected request to %s", r.URL.Path)
				}
			},
			want: activities.DeleteFromAMSSLocationActivityResult{Deleted: true},
		},
		{
			name:    "Fails getting pipeline UUID (HTTP error)",
			approve: true,
			handler: func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path == "/api/v2/file/"+aipUUID+"/" {
					http.Error(w, "", http.StatusInternalServerError)
					return
				}
				t.Fatalf("unexpected request to %s", r.URL.Path)
			},
			wantErr: "get pipeline UUID: storage service request failed with status 500",
		},
		{
			name:    "Fails getting pipeline UUID (decode error)",
			approve: true,
			handler: func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path == "/api/v2/file/"+aipUUID+"/" {
					w.Header().Set("Content-Type", "application/json")
					_, err := w.Write([]byte("{"))
					assert.NilError(t, err)
					return
				}
				t.Fatalf("unexpected request to %s", r.URL.Path)
			},
			wantErr: "get pipeline UUID:",
		},
		{
			name:    "Fails getting pipeline UUID (missing origin pipeline)",
			approve: true,
			handler: func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path == "/api/v2/file/"+aipUUID+"/" {
					w.Header().Set("Content-Type", "application/json")
					err := json.NewEncoder(w).Encode(map[string]any{
						"status": "UPLOADED",
					})
					assert.NilError(t, err)
					return
				}
				t.Fatalf("unexpected request to %s", r.URL.Path)
			},
			wantErr: "get pipeline UUID: missing origin pipeline",
		},
		{
			name:    "Fails requesting deletion (HTTP error)",
			approve: true,
			handler: func(w http.ResponseWriter, r *http.Request) {
				switch r.URL.Path {
				case "/api/v2/file/" + aipUUID + "/":
					handleAIPInfo(w, r, "UPLOADED")
				case "/api/v2/file/" + aipUUID + "/delete_aip/":
					http.Error(w, "", http.StatusInternalServerError)
				default:
					t.Fatalf("unexpected request to %s", r.URL.Path)
				}
			},
			wantErr: "request deletion: storage service request failed with status 500",
		},
		{
			name:    "Fails requesting deletion (decode error)",
			approve: true,
			handler: func(w http.ResponseWriter, r *http.Request) {
				switch r.URL.Path {
				case "/api/v2/file/" + aipUUID + "/":
					handleAIPInfo(w, r, "UPLOADED")
				case "/api/v2/file/" + aipUUID + "/delete_aip/":
					w.Header().Set("Content-Type", "application/json")
					_, err := w.Write([]byte("{"))
					assert.NilError(t, err)
				default:
					t.Fatalf("unexpected request to %s", r.URL.Path)
				}
			},
			wantErr: "request deletion:",
		},
		{
			name:    "Fails approving existing deletion request",
			approve: true,
			handler: func(w http.ResponseWriter, r *http.Request) {
				switch r.URL.Path {
				case "/api/v2/file/" + aipUUID + "/":
					handleAIPInfo(w, r, "UPLOADED")
				case "/api/v2/file/" + aipUUID + "/delete_aip/":
					handleExistingRequest(w, r)
				default:
					t.Fatalf("unexpected request to %s", r.URL.Path)
				}
			},
			wantErr: "approve deletion: deletion request already exists and cannot be approved without an event ID",
		},
		{
			name:    "Fails approving deletion (HTTP error)",
			approve: true,
			handler: func(w http.ResponseWriter, r *http.Request) {
				switch r.URL.Path {
				case "/api/v2/file/" + aipUUID + "/":
					handleAIPInfo(w, r, "UPLOADED")
				case "/api/v2/file/" + aipUUID + "/delete_aip/":
					handleRequest(w, r)
				case "/api/v2/file/" + aipUUID + "/review_aip_deletion/":
					http.Error(w, "", http.StatusInternalServerError)
				default:
					t.Fatalf("unexpected request to %s", r.URL.Path)
				}
			},
			wantErr: "approve deletion: storage service request failed with status 500",
		},
		{
			name:    "Fails approving deletion (decode error)",
			approve: true,
			handler: func(w http.ResponseWriter, r *http.Request) {
				switch r.URL.Path {
				case "/api/v2/file/" + aipUUID + "/":
					handleAIPInfo(w, r, "UPLOADED")
				case "/api/v2/file/" + aipUUID + "/delete_aip/":
					handleRequest(w, r)
				case "/api/v2/file/" + aipUUID + "/review_aip_deletion/":
					w.Header().Set("Content-Type", "application/json")
					_, err := w.Write([]byte("{"))
					assert.NilError(t, err)
				default:
					t.Fatalf("unexpected request to %s", r.URL.Path)
				}
			},
			wantErr: "approve deletion:",
		},
		{
			name:    "Fails approving deletion (response error_message)",
			approve: true,
			handler: func(w http.ResponseWriter, r *http.Request) {
				switch r.URL.Path {
				case "/api/v2/file/" + aipUUID + "/":
					handleAIPInfo(w, r, "UPLOADED")
				case "/api/v2/file/" + aipUUID + "/delete_aip/":
					handleRequest(w, r)
				case "/api/v2/file/" + aipUUID + "/review_aip_deletion/":
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusOK)
					err := json.NewEncoder(w).Encode(map[string]any{"error_message": "error message"})
					assert.NilError(t, err)
				default:
					t.Fatalf("unexpected request to %s", r.URL.Path)
				}
			},
			wantErr: "approve deletion: error message",
		},
		{
			name: "Keeps polling when a new deletion request still reports uploaded",
			handler: func() http.HandlerFunc {
				var requested bool
				polls := 0
				return func(w http.ResponseWriter, r *http.Request) {
					switch r.URL.Path {
					case "/api/v2/file/" + aipUUID + "/":
						if !requested {
							handleAIPInfo(w, r, "UPLOADED")
							return
						}
						polls++
						switch polls {
						case 1:
							handleAIPInfo(w, r, "UPLOADED")
						case 2:
							handleAIPInfo(w, r, "DEL_REQ")
						default:
							handleAIPInfo(w, r, "DELETED")
						}
					case "/api/v2/file/" + aipUUID + "/delete_aip/":
						handleRequest(w, r)
						requested = true
					default:
						t.Fatalf("unexpected request to %s", r.URL.Path)
					}
				}
			}(),
			want: activities.DeleteFromAMSSLocationActivityResult{Deleted: true},
		},
		{
			name: "Deletes AIP after polling for status",
			handler: func() http.HandlerFunc {
				var requested, polled bool
				return func(w http.ResponseWriter, r *http.Request) {
					switch r.URL.Path {
					case "/api/v2/file/" + aipUUID + "/":
						if !requested {
							handleAIPInfo(w, r, "UPLOADED")
							return
						}
						if !polled {
							handleAIPInfo(w, r, "DEL_REQ")
							polled = true
							return
						}
						handleAIPInfo(w, r, "DELETED")
					case "/api/v2/file/" + aipUUID + "/delete_aip/":
						handleRequest(w, r)
						requested = true
					default:
						t.Fatalf("unexpected request to %s", r.URL.Path)
					}
				}
			}(),
			want: activities.DeleteFromAMSSLocationActivityResult{Deleted: true},
		},
		{
			name: "Doesn't delete AIP after polling for status",
			handler: func() http.HandlerFunc {
				var requested, polled bool
				return func(w http.ResponseWriter, r *http.Request) {
					switch r.URL.Path {
					case "/api/v2/file/" + aipUUID + "/":
						if !requested {
							handleAIPInfo(w, r, "UPLOADED")
							return
						}
						if !polled {
							handleAIPInfo(w, r, "DEL_REQ")
							polled = true
							return
						}
						handleAIPInfo(w, r, "UPLOADED")
					case "/api/v2/file/" + aipUUID + "/delete_aip/":
						handleRequest(w, r)
						requested = true
					default:
						t.Fatalf("unexpected request to %s", r.URL.Path)
					}
				}
			}(),
			want: activities.DeleteFromAMSSLocationActivityResult{Deleted: false},
		},
		{
			name: "Polls existing deletion request until AIP is deleted",
			handler: func() http.HandlerFunc {
				var requested, polled bool
				return func(w http.ResponseWriter, r *http.Request) {
					switch r.URL.Path {
					case "/api/v2/file/" + aipUUID + "/":
						if !requested {
							handleAIPInfo(w, r, "UPLOADED")
							return
						}
						if !polled {
							handleAIPInfo(w, r, "DEL_REQ")
							polled = true
							return
						}
						handleAIPInfo(w, r, "DELETED")
					case "/api/v2/file/" + aipUUID + "/delete_aip/":
						handleExistingRequest(w, r)
						requested = true
					default:
						t.Fatalf("unexpected request to %s", r.URL.Path)
					}
				}
			}(),
			want: activities.DeleteFromAMSSLocationActivityResult{Deleted: true},
		},
		{
			name: "Stops polling when an existing deletion request still reports uploaded",
			handler: func() http.HandlerFunc {
				var requested bool
				return func(w http.ResponseWriter, r *http.Request) {
					switch r.URL.Path {
					case "/api/v2/file/" + aipUUID + "/":
						if !requested {
							handleAIPInfo(w, r, "UPLOADED")
							return
						}
						handleAIPInfo(w, r, "UPLOADED")
					case "/api/v2/file/" + aipUUID + "/delete_aip/":
						handleExistingRequest(w, r)
						requested = true
					default:
						t.Fatalf("unexpected request to %s", r.URL.Path)
					}
				}
			}(),
			want: activities.DeleteFromAMSSLocationActivityResult{Deleted: false},
		},
		{
			name: "Polls existing deletion request until AIP is restored to uploaded",
			handler: func() http.HandlerFunc {
				var requested, polled bool
				return func(w http.ResponseWriter, r *http.Request) {
					switch r.URL.Path {
					case "/api/v2/file/" + aipUUID + "/":
						if !requested {
							handleAIPInfo(w, r, "UPLOADED")
							return
						}
						if !polled {
							handleAIPInfo(w, r, "DEL_REQ")
							polled = true
							return
						}
						handleAIPInfo(w, r, "UPLOADED")
					case "/api/v2/file/" + aipUUID + "/delete_aip/":
						handleExistingRequest(w, r)
						requested = true
					default:
						t.Fatalf("unexpected request to %s", r.URL.Path)
					}
				}
			}(),
			want: activities.DeleteFromAMSSLocationActivityResult{Deleted: false},
		},
		{
			name: "Fails polling status (missing package status)",
			handler: func() http.HandlerFunc {
				var requested bool
				return func(w http.ResponseWriter, r *http.Request) {
					switch r.URL.Path {
					case "/api/v2/file/" + aipUUID + "/":
						w.Header().Set("Content-Type", "application/json")
						var payload map[string]any
						if requested {
							payload = map[string]any{
								"origin_pipeline": "/api/v2/pipeline/" + pipelineUUID + "/",
							}
						} else {
							payload = map[string]any{
								"origin_pipeline": "/api/v2/pipeline/" + pipelineUUID + "/",
								"status":          "UPLOADED",
							}
						}
						err := json.NewEncoder(w).Encode(payload)
						assert.NilError(t, err)
					case "/api/v2/file/" + aipUUID + "/delete_aip/":
						handleRequest(w, r)
						requested = true
					default:
						t.Fatalf("unexpected request to %s", r.URL.Path)
					}
				}
			}(),
			wantErr: "poll status: missing package status",
		},
		{
			name:    "Fails with invalid URL",
			approve: true,
			url:     string([]byte{0x7f}),
			poll:    -1,
			wantErr: "invalid control character in URL",
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ts := &temporalsdk_testsuite.WorkflowTestSuite{}
			env := ts.NewTestActivityEnvironment()
			env.SetTestTimeout(5 * time.Second)
			httpClient := cleanhttp.DefaultPooledClient()
			url := tt.url
			if tt.handler != nil {
				srv := setupServer(t, tt.handler)
				url = srv.URL
			}
			pollInterval := tt.poll
			if pollInterval == 0 {
				pollInterval = time.Microsecond
			}

			env.RegisterActivityWithOptions(
				activities.NewDeleteFromAMSSLocationActivity(httpClient, tt.approve, pollInterval).Execute,
				temporalsdk_activity.RegisterOptions{Name: storage.DeleteFromAMSSLocationActivityName},
			)

			fut, err := env.ExecuteActivity(
				storage.DeleteFromAMSSLocationActivityName,
				activities.DeleteFromAMSSLocationActivityParams{
					Config: types.AMSSConfig{
						URL:      url,
						Username: "test",
						APIKey:   "test",
					},
					AIPUUID: uuid.MustParse(aipUUID),
				},
			)
			if tt.wantErr != "" {
				assert.ErrorContains(t, err, tt.wantErr)
				return
			}
			assert.NilError(t, err)

			var res activities.DeleteFromAMSSLocationActivityResult
			err = fut.Get(&res)
			assert.NilError(t, err)
			assert.DeepEqual(t, res, tt.want)
		})
	}
}
