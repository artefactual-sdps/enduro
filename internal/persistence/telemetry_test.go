package persistence_test

import (
	"errors"
	"testing"
	"time"

	"github.com/artefactual-sdps/enduro/internal/datatypes"
	"github.com/artefactual-sdps/enduro/internal/persistence"
	"github.com/artefactual-sdps/enduro/internal/persistence/fake"
	"github.com/google/uuid"
	"go.artefactual.dev/tools/mockutil"
	"go.artefactual.dev/tools/ref"
	"go.opentelemetry.io/otel/trace/noop"
	"go.uber.org/mock/gomock"
	"gotest.tools/v3/assert"
)

func TestCreateUser(t *testing.T) {
	t.Parallel()

	userID := uuid.New()
	createdAt := ref.New(time.Now().Truncate(time.Second))

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
							Name:      ref.New("Test User"),
						},
					).Return(nil)
			},
			params: &datatypes.User{
				UUID:      userID,
				CreatedAt: createdAt,
				Name:      ref.New("Test User"),
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
							Name:      ref.New("Test User"),
						},
					).
					Return(errors.New("invalid data error: field \"UUID\" is required"))
			},
			params: &datatypes.User{
				CreatedAt: ref.New(time.Now().Truncate(time.Second)),
				Name:      ref.New("Test User"),
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
	createdAt := ref.New(time.Now().Truncate(time.Second))

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
							Name:      ref.New("Test User"),
						},
						nil,
					)
			},
			params: userID,
			want: &datatypes.User{
				UUID:      userID,
				CreatedAt: createdAt,
				Name:      ref.New("Test User"),
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
