package persistence_test

import (
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"go.artefactual.dev/tools/mockutil"
	"go.opentelemetry.io/otel/trace/noop"
	"go.uber.org/mock/gomock"
	"gotest.tools/v3/assert"

	"github.com/artefactual-sdps/enduro/internal/datatypes"
	"github.com/artefactual-sdps/enduro/internal/persistence"
	"github.com/artefactual-sdps/enduro/internal/persistence/fake"
)

func TestCreateUser(t *testing.T) {
	t.Parallel()

	userID := uuid.New()
	createdAt := time.Now().Truncate(time.Second)

	type test struct {
		name    string
		mock    func(*fake.MockService)
		params  *datatypes.User
		wantErr string
	}
	for _, tt := range []test{
		{
			name: "Creates a new user",
			mock: func(svc *fake.MockService) {
				svc.EXPECT().
					CreateUser(
						mockutil.Context(),
						&datatypes.User{
							UUID:      userID,
							CreatedAt: createdAt,
							Name:      "Test User",
						},
					).Return(nil)
			},
			params: &datatypes.User{
				UUID:      userID,
				CreatedAt: createdAt,
				Name:      "Test User",
			},
		},
		{
			name: "Errors when creating a user",
			mock: func(svc *fake.MockService) {
				svc.EXPECT().
					CreateUser(
						mockutil.Context(),
						&datatypes.User{
							CreatedAt: createdAt,
							Name:      "Test User",
						},
					).
					Return(errors.New("invalid data error: field \"UUID\" is required"))
			},
			params: &datatypes.User{
				CreatedAt: createdAt,
				Name:      "Test User",
			},
			wantErr: "CreateUser: invalid data error: field \"UUID\" is required",
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			svc := fake.NewMockService(gomock.NewController(t))
			if tt.mock != nil {
				tt.mock(svc)
			}

			tracer := noop.NewTracerProvider().Tracer("test")
			w := persistence.WithTelemetry(svc, tracer)

			err := w.CreateUser(t.Context(), tt.params)
			if tt.wantErr != "" {
				assert.Error(t, err, tt.wantErr)
				return
			}

			assert.NilError(t, err)
		})
	}
}

func TestReadUser(t *testing.T) {
	t.Parallel()

	userID := uuid.New()
	createdAt := time.Now().Truncate(time.Second)

	type test struct {
		name    string
		mock    func(*fake.MockService)
		params  uuid.UUID
		want    *datatypes.User
		wantErr string
	}
	for _, tt := range []test{
		{
			name: "Reads a user",
			mock: func(svc *fake.MockService) {
				svc.EXPECT().
					ReadUser(mockutil.Context(), userID).
					Return(
						&datatypes.User{
							UUID:      userID,
							CreatedAt: createdAt,
							Name:      "Test User",
						},
						nil,
					)
			},
			params: userID,
			want: &datatypes.User{
				UUID:      userID,
				CreatedAt: createdAt,
				Name:      "Test User",
			},
		},
		{
			name: "Errors when reading a user",
			mock: func(svc *fake.MockService) {
				svc.EXPECT().
					ReadUser(mockutil.Context(), userID).
					Return(nil, errors.New("not found error: db: user not found"))
			},
			params:  userID,
			wantErr: "ReadUser: not found error: db: user not found",
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			svc := fake.NewMockService(gomock.NewController(t))
			if tt.mock != nil {
				tt.mock(svc)
			}

			tracer := noop.NewTracerProvider().Tracer("test")
			w := persistence.WithTelemetry(svc, tracer)

			got, err := w.ReadUser(t.Context(), tt.params)
			if tt.wantErr != "" {
				assert.Error(t, err, tt.wantErr)
				return
			}

			assert.NilError(t, err)
			assert.DeepEqual(t, got, tt.want)
		})
	}
}

func TestReadOIDCUser(t *testing.T) {
	t.Parallel()

	userID := uuid.New()
	createdAt := time.Now().Truncate(time.Second)

	type params struct {
		iss string
		sub string
	}

	type test struct {
		name    string
		mock    func(*fake.MockService)
		args    params
		want    *datatypes.User
		wantErr string
	}
	for _, tt := range []test{
		{
			name: "Reads a user using OIDC data",
			mock: func(svc *fake.MockService) {
				svc.EXPECT().
					ReadOIDCUser(
						mockutil.Context(),
						"https://oidc.example.com",
						"1234567890",
					).
					Return(
						&datatypes.User{
							UUID:      userID,
							CreatedAt: createdAt,
							Name:      "Test User",
							OIDCIss:   "https://oidc.example.com",
							OIDCSub:   "1234567890",
						},
						nil,
					)
			},
			args: params{iss: "https://oidc.example.com", sub: "1234567890"},
			want: &datatypes.User{
				UUID:      userID,
				CreatedAt: createdAt,
				Name:      "Test User",
				OIDCIss:   "https://oidc.example.com",
				OIDCSub:   "1234567890",
			},
		},
		{
			name: "Errors when reading a user using OIDC data",
			mock: func(svc *fake.MockService) {
				svc.EXPECT().
					ReadOIDCUser(mockutil.Context(), "", "").
					Return(nil, errors.New("not found error: db: user not found"))
			},
			args:    params{iss: "", sub: ""},
			wantErr: "ReadOIDCUser: not found error: db: user not found",
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			svc := fake.NewMockService(gomock.NewController(t))
			if tt.mock != nil {
				tt.mock(svc)
			}

			tracer := noop.NewTracerProvider().Tracer("test")
			w := persistence.WithTelemetry(svc, tracer)

			got, err := w.ReadOIDCUser(t.Context(), tt.args.iss, tt.args.sub)
			if tt.wantErr != "" {
				assert.Error(t, err, tt.wantErr)
				return
			}

			assert.NilError(t, err)
			assert.DeepEqual(t, got, tt.want)
		})
	}
}

func TestListUsers(t *testing.T) {
	t.Parallel()

	userID := uuid.New()
	createdAt := time.Now().Truncate(time.Second)

	type test struct {
		name    string
		mock    func(*fake.MockService)
		filter  *persistence.UserFilter
		want    []*datatypes.User
		wantPg  *persistence.Page
		wantErr string
	}
	for _, tt := range []test{
		{
			name: "Lists all users",
			mock: func(svc *fake.MockService) {
				svc.EXPECT().
					ListUsers(mockutil.Context(), nil).
					Return(
						[]*datatypes.User{
							{
								UUID:      userID,
								CreatedAt: createdAt,
								Email:     "nobody@example.com",
								Name:      "Nobody Example",
								OIDCIss:   "https://oidc.example.com",
								OIDCSub:   "1234567890",
							},
						},
						&persistence.Page{
							Limit:  20,
							Offset: 0,
							Total:  1,
						},
						nil,
					)
			},
			want: []*datatypes.User{
				{
					UUID:      userID,
					CreatedAt: createdAt,
					Email:     "nobody@example.com",
					Name:      "Nobody Example",
					OIDCIss:   "https://oidc.example.com",
					OIDCSub:   "1234567890",
				},
			},
			wantPg: &persistence.Page{
				Limit:  20,
				Offset: 0,
				Total:  1,
			},
		},
		{
			name: "Errors when listing users",
			mock: func(svc *fake.MockService) {
				svc.EXPECT().
					ListUsers(mockutil.Context(), nil).
					Return(nil, nil, persistence.ErrNotValid)
			},
			wantErr: fmt.Sprintf("ListUsers: %v", persistence.ErrNotValid),
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			svc := fake.NewMockService(gomock.NewController(t))
			if tt.mock != nil {
				tt.mock(svc)
			}

			tracer := noop.NewTracerProvider().Tracer("test")
			w := persistence.WithTelemetry(svc, tracer)

			got, pg, err := w.ListUsers(t.Context(), tt.filter)
			if tt.wantErr != "" {
				assert.Error(t, err, tt.wantErr)
				return
			}

			assert.NilError(t, err)
			assert.DeepEqual(t, got, tt.want)
			assert.DeepEqual(t, pg, tt.wantPg)
		})
	}
}

func TestCreateBatch(t *testing.T) {
	t.Parallel()

	batchID := uuid.New()
	createdAt := time.Now().Truncate(time.Second)

	type test struct {
		name    string
		mock    func(*fake.MockService)
		params  *datatypes.Batch
		wantErr string
	}
	for _, tt := range []test{
		{
			name: "Creates a new batch",
			mock: func(svc *fake.MockService) {
				svc.EXPECT().
					CreateBatch(
						mockutil.Context(),
						&datatypes.Batch{
							UUID:       batchID,
							CreatedAt:  createdAt,
							Identifier: "batch-001",
						},
					).Return(nil)
			},
			params: &datatypes.Batch{
				UUID:       batchID,
				CreatedAt:  createdAt,
				Identifier: "batch-001",
			},
		},
		{
			name: "Errors when creating a batch",
			mock: func(svc *fake.MockService) {
				svc.EXPECT().
					CreateBatch(
						mockutil.Context(),
						&datatypes.Batch{
							CreatedAt:  createdAt,
							Identifier: "batch-001",
						},
					).
					Return(errors.New("invalid data error: field \"UUID\" is required"))
			},
			params: &datatypes.Batch{
				CreatedAt:  createdAt,
				Identifier: "batch-001",
			},
			wantErr: "CreateBatch: invalid data error: field \"UUID\" is required",
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			svc := fake.NewMockService(gomock.NewController(t))
			if tt.mock != nil {
				tt.mock(svc)
			}

			tracer := noop.NewTracerProvider().Tracer("test")
			w := persistence.WithTelemetry(svc, tracer)

			err := w.CreateBatch(t.Context(), tt.params)
			if tt.wantErr != "" {
				assert.Error(t, err, tt.wantErr)
				return
			}

			assert.NilError(t, err)
		})
	}
}

func TestUpdateBatch(t *testing.T) {
	t.Parallel()

	batchID := uuid.New()
	createdAt := time.Now().Truncate(time.Second)

	type test struct {
		name    string
		mock    func(*fake.MockService)
		id      uuid.UUID
		updater persistence.BatchUpdater
		want    *datatypes.Batch
		wantErr string
	}
	for _, tt := range []test{
		{
			name: "Updates a batch",
			mock: func(svc *fake.MockService) {
				svc.EXPECT().
					UpdateBatch(mockutil.Context(), batchID, nil).
					Return(
						&datatypes.Batch{
							UUID:       batchID,
							CreatedAt:  createdAt,
							Identifier: "batch-001-updated",
						},
						nil,
					)
			},
			id: batchID,
			want: &datatypes.Batch{
				UUID:       batchID,
				CreatedAt:  createdAt,
				Identifier: "batch-001-updated",
			},
		},
		{
			name: "Errors when updating a batch",
			mock: func(svc *fake.MockService) {
				svc.EXPECT().
					UpdateBatch(mockutil.Context(), batchID, nil).
					Return(nil, errors.New("not found error: db: batch not found"))
			},
			id:      batchID,
			wantErr: "UpdateBatch: not found error: db: batch not found",
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			svc := fake.NewMockService(gomock.NewController(t))
			if tt.mock != nil {
				tt.mock(svc)
			}

			tracer := noop.NewTracerProvider().Tracer("test")
			w := persistence.WithTelemetry(svc, tracer)

			got, err := w.UpdateBatch(t.Context(), tt.id, tt.updater)
			if tt.wantErr != "" {
				assert.Error(t, err, tt.wantErr)
				return
			}

			assert.NilError(t, err)
			assert.DeepEqual(t, got, tt.want)
		})
	}
}

func TestDeleteBatch(t *testing.T) {
	t.Parallel()

	batchUUID := uuid.New()

	type test struct {
		name    string
		mock    func(*fake.MockService)
		id      uuid.UUID
		wantErr string
	}
	for _, tt := range []test{
		{
			name: "Deletes a batch",
			mock: func(svc *fake.MockService) {
				svc.EXPECT().
					DeleteBatch(mockutil.Context(), batchUUID).
					Return(nil)
			},
			id: batchUUID,
		},
		{
			name: "Errors when deleting a batch",
			mock: func(svc *fake.MockService) {
				svc.EXPECT().
					DeleteBatch(mockutil.Context(), batchUUID).
					Return(errors.New("not found error: db: batch not found"))
			},
			id:      batchUUID,
			wantErr: "DeleteBatch: not found error: db: batch not found",
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			svc := fake.NewMockService(gomock.NewController(t))
			if tt.mock != nil {
				tt.mock(svc)
			}

			tracer := noop.NewTracerProvider().Tracer("test")
			w := persistence.WithTelemetry(svc, tracer)

			err := w.DeleteBatch(t.Context(), tt.id)
			if tt.wantErr != "" {
				assert.Error(t, err, tt.wantErr)
				return
			}

			assert.NilError(t, err)
		})
	}
}

func TestReadBatch(t *testing.T) {
	t.Parallel()

	batchID := uuid.New()
	createdAt := time.Now().Truncate(time.Second)

	type test struct {
		name    string
		mock    func(*fake.MockService)
		params  uuid.UUID
		want    *datatypes.Batch
		wantErr string
	}
	for _, tt := range []test{
		{
			name: "Reads a batch",
			mock: func(svc *fake.MockService) {
				svc.EXPECT().
					ReadBatch(mockutil.Context(), batchID).
					Return(
						&datatypes.Batch{
							UUID:       batchID,
							CreatedAt:  createdAt,
							Identifier: "batch-001",
						},
						nil,
					)
			},
			params: batchID,
			want: &datatypes.Batch{
				UUID:       batchID,
				CreatedAt:  createdAt,
				Identifier: "batch-001",
			},
		},
		{
			name: "Errors when reading a batch",
			mock: func(svc *fake.MockService) {
				svc.EXPECT().
					ReadBatch(mockutil.Context(), batchID).
					Return(nil, errors.New("not found error: db: batch not found"))
			},
			params:  batchID,
			wantErr: "ReadBatch: not found error: db: batch not found",
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			svc := fake.NewMockService(gomock.NewController(t))
			if tt.mock != nil {
				tt.mock(svc)
			}

			tracer := noop.NewTracerProvider().Tracer("test")
			w := persistence.WithTelemetry(svc, tracer)

			got, err := w.ReadBatch(t.Context(), tt.params)
			if tt.wantErr != "" {
				assert.Error(t, err, tt.wantErr)
				return
			}

			assert.NilError(t, err)
			assert.DeepEqual(t, got, tt.want)
		})
	}
}

func TestListBatches(t *testing.T) {
	t.Parallel()

	batchID := uuid.New()
	createdAt := time.Now().Truncate(time.Second)

	type test struct {
		name    string
		mock    func(*fake.MockService)
		filter  *persistence.BatchFilter
		want    []*datatypes.Batch
		wantPg  *persistence.Page
		wantErr string
	}
	for _, tt := range []test{
		{
			name: "Lists all batches",
			mock: func(svc *fake.MockService) {
				svc.EXPECT().
					ListBatches(mockutil.Context(), nil).
					Return(
						[]*datatypes.Batch{
							{
								UUID:       batchID,
								CreatedAt:  createdAt,
								Identifier: "batch-001",
							},
						},
						&persistence.Page{
							Limit:  20,
							Offset: 0,
							Total:  1,
						},
						nil,
					)
			},
			want: []*datatypes.Batch{
				{
					UUID:       batchID,
					CreatedAt:  createdAt,
					Identifier: "batch-001",
				},
			},
			wantPg: &persistence.Page{
				Limit:  20,
				Offset: 0,
				Total:  1,
			},
		},
		{
			name: "Errors when listing batches",
			mock: func(svc *fake.MockService) {
				svc.EXPECT().
					ListBatches(mockutil.Context(), nil).
					Return(nil, nil, persistence.ErrNotValid)
			},
			wantErr: fmt.Sprintf("ListBatches: %v", persistence.ErrNotValid),
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			svc := fake.NewMockService(gomock.NewController(t))
			if tt.mock != nil {
				tt.mock(svc)
			}

			tracer := noop.NewTracerProvider().Tracer("test")
			w := persistence.WithTelemetry(svc, tracer)

			got, pg, err := w.ListBatches(t.Context(), tt.filter)
			if tt.wantErr != "" {
				assert.Error(t, err, tt.wantErr)
				return
			}

			assert.NilError(t, err)
			assert.DeepEqual(t, got, tt.want)
			assert.DeepEqual(t, pg, tt.wantPg)
		})
	}
}
