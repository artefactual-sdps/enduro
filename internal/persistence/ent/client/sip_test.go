package entclient_test

import (
	"database/sql"
	"fmt"
	"testing"
	"time"

	"github.com/go-logr/logr"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/google/uuid"
	"go.artefactual.dev/tools/ref"
	"gotest.tools/v3/assert"

	"github.com/artefactual-sdps/enduro/internal/datatypes"
	"github.com/artefactual-sdps/enduro/internal/entfilter"
	"github.com/artefactual-sdps/enduro/internal/enums"
	"github.com/artefactual-sdps/enduro/internal/persistence"
	"github.com/artefactual-sdps/enduro/internal/persistence/ent/db"
	"github.com/artefactual-sdps/enduro/internal/timerange"
)

func TestCreateSIP(t *testing.T) {
	t.Parallel()

	sipUUID := uuid.New()
	aipID := uuid.NullUUID{UUID: uuid.New(), Valid: true}
	started := sql.NullTime{Time: time.Now(), Valid: true}
	completed := sql.NullTime{Time: started.Time.Add(time.Second), Valid: true}
	uploaderID := uuid.New()

	tests := []struct {
		name    string
		sip     *datatypes.SIP
		want    *datatypes.SIP
		wantErr string
	}{
		{
			name: "Creates a new SIP in the DB",
			sip: &datatypes.SIP{
				UUID:        sipUUID,
				Name:        "Test SIP 1",
				AIPID:       aipID,
				Status:      enums.SIPStatusProcessing,
				StartedAt:   started,
				CompletedAt: completed,
			},
			want: &datatypes.SIP{
				ID:          1,
				UUID:        sipUUID,
				Name:        "Test SIP 1",
				AIPID:       aipID,
				Status:      enums.SIPStatusProcessing,
				CreatedAt:   time.Now(),
				StartedAt:   started,
				CompletedAt: completed,
			},
		},
		{
			name: "Creates a SIP with missing optional fields",
			sip: &datatypes.SIP{
				UUID:   sipUUID,
				Name:   "Test SIP 2",
				Status: enums.SIPStatusProcessing,
			},
			want: &datatypes.SIP{
				ID:        1,
				UUID:      sipUUID,
				Name:      "Test SIP 2",
				Status:    enums.SIPStatusProcessing,
				CreatedAt: time.Now(),
			},
		},
		{
			name: "Creates a SIP with an uploader",
			sip: &datatypes.SIP{
				UUID:     sipUUID,
				Name:     "Test SIP 3",
				Status:   enums.SIPStatusProcessing,
				Uploader: &datatypes.Uploader{UUID: uploaderID},
			},
			want: &datatypes.SIP{
				ID:        1,
				UUID:      sipUUID,
				Name:      "Test SIP 3",
				Status:    enums.SIPStatusProcessing,
				CreatedAt: time.Now(),
				Uploader: &datatypes.Uploader{
					UUID:  uploaderID,
					Email: "nobody@example.com",
					Name:  "Test User",
				},
			},
		},
		{
			name:    "Required field error for missing UUID",
			sip:     &datatypes.SIP{},
			wantErr: "invalid data error: field \"UUID\" is required",
		},
		{
			name:    "Required field error for missing Name",
			sip:     &datatypes.SIP{UUID: sipUUID},
			wantErr: "invalid data error: field \"Name\" is required",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			c, svc := setUpClient(t, logr.Discard())
			ctx := t.Context()
			sip := *tt.sip // Make a local copy of sip.

			_, err := createUser(t, c, uploaderID)
			assert.NilError(t, err)

			err = svc.CreateSIP(ctx, &sip)
			if tt.wantErr != "" {
				assert.Error(t, err, tt.wantErr)
				return
			}
			assert.NilError(t, err)

			assert.DeepEqual(t, &sip, tt.want,
				cmpopts.EquateApproxTime(time.Millisecond*100),
				cmpopts.IgnoreUnexported(db.SIP{}, db.SIPEdges{}),
			)
		})
	}
}

func TestUpdateSIP(t *testing.T) {
	t.Parallel()

	sipUUID := uuid.New()
	sipUUID2 := uuid.New()
	aipID := uuid.NullUUID{
		UUID:  uuid.MustParse("e2ace0da-8697-453d-9ea1-4c9b62309e54"),
		Valid: true,
	}
	aipID2 := uuid.NullUUID{
		UUID:  uuid.MustParse("7d085541-af56-4444-9ce2-d6401ff4c97b"),
		Valid: true,
	}
	uploaderID := uuid.New()

	started := sql.NullTime{Time: time.Now(), Valid: true}
	started2 := sql.NullTime{
		Time: func() time.Time {
			t, _ := time.Parse(time.RFC3339, "1980-01-01T09:30:00Z")
			return t
		}(),
		Valid: true,
	}

	completed := sql.NullTime{Time: started.Time.Add(time.Second), Valid: true}
	completed2 := sql.NullTime{Time: started2.Time.Add(time.Second), Valid: true}

	type params struct {
		sipUUID uuid.UUID
		sip     *datatypes.SIP
		updater persistence.SIPUpdater
	}
	tests := []struct {
		name    string
		args    params
		want    *datatypes.SIP
		wantErr string
	}{
		{
			name: "Updates all SIP columns",
			args: params{
				sipUUID: sipUUID,
				sip: &datatypes.SIP{
					UUID:        sipUUID,
					Name:        "Test SIP",
					AIPID:       aipID,
					Status:      enums.SIPStatusProcessing,
					StartedAt:   started,
					CompletedAt: completed,
					Uploader: &datatypes.Uploader{
						UUID:  uploaderID,
						Email: "nobody@example.com",
						Name:  "Test User",
					},
				},
				updater: func(s *datatypes.SIP) (*datatypes.SIP, error) {
					s.ID = 100        // No-op, can't update ID.
					s.UUID = sipUUID2 // No-op, can't update UUID.
					s.Name = "Updated SIP"
					s.AIPID = aipID2
					s.Status = enums.SIPStatusIngested
					s.CreatedAt = started2.Time // No-op, can't update CreatedAt.
					s.StartedAt = started2
					s.CompletedAt = completed2
					s.FailedAs = enums.SIPFailedAsSIP
					s.FailedKey = "failed-key"
					s.Uploader = &datatypes.Uploader{UUID: uuid.New()} // No-op, can't update Uploader.
					return s, nil
				},
			},
			want: &datatypes.SIP{
				ID:          1,
				UUID:        sipUUID,
				Name:        "Updated SIP",
				AIPID:       aipID2,
				Status:      enums.SIPStatusIngested,
				CreatedAt:   time.Now(),
				StartedAt:   started2,
				CompletedAt: completed2,
				FailedAs:    enums.SIPFailedAsSIP,
				FailedKey:   "failed-key",
				Uploader: &datatypes.Uploader{
					UUID:  uploaderID,
					Email: "nobody@example.com",
					Name:  "Test User",
				},
			},
		},
		{
			name: "Only updates selected columns",
			args: params{
				sipUUID: sipUUID,
				sip: &datatypes.SIP{
					UUID:      sipUUID,
					Name:      "Test SIP",
					AIPID:     aipID,
					Status:    enums.SIPStatusProcessing,
					StartedAt: started,
				},
				updater: func(s *datatypes.SIP) (*datatypes.SIP, error) {
					// Immutable.
					s.ID = 1234

					// Mutable.
					s.Status = enums.SIPStatusIngested
					s.CompletedAt = completed

					return s, nil
				},
			},
			want: &datatypes.SIP{
				ID:          1,
				UUID:        sipUUID,
				Name:        "Test SIP",
				AIPID:       aipID,
				Status:      enums.SIPStatusIngested,
				CreatedAt:   time.Now(),
				StartedAt:   started,
				CompletedAt: completed,
			},
		},
		{
			name: "Ignores invalid fields",
			args: params{
				sipUUID: sipUUID,
				sip: &datatypes.SIP{
					UUID:   sipUUID,
					Name:   "Test SIP",
					AIPID:  aipID,
					Status: enums.SIPStatusQueued,
				},
				updater: func(s *datatypes.SIP) (*datatypes.SIP, error) {
					// Invalid.
					s.Status = ""
					s.FailedAs = ""

					// Valid.
					s.StartedAt = started

					return s, nil
				},
			},
			want: &datatypes.SIP{
				ID:        1,
				UUID:      sipUUID,
				Name:      "Test SIP",
				AIPID:     aipID,
				Status:    enums.SIPStatusQueued,
				CreatedAt: time.Now(),
				StartedAt: started,
			},
		},
		{
			name:    "Errors when SIP to update is not found",
			args:    params{sipUUID: sipUUID},
			wantErr: "not found error: db: sip not found",
		},
		{
			name: "Errors when the updater errors",
			args: params{
				sipUUID: sipUUID,
				sip: &datatypes.SIP{
					UUID:   sipUUID,
					Name:   "Test SIP",
					AIPID:  aipID,
					Status: enums.SIPStatusProcessing,
				},
				updater: func(s *datatypes.SIP) (*datatypes.SIP, error) {
					return nil, fmt.Errorf("Bad input")
				},
			},
			wantErr: "invalid data error: updater error: Bad input",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			c, svc := setUpClient(t, logr.Discard())
			ctx := t.Context()

			if tt.args.sip != nil {
				sip := *tt.args.sip // Make a local copy of sip.

				_, err := createUser(t, c, uploaderID)
				assert.NilError(t, err)

				err = svc.CreateSIP(ctx, &sip)
				assert.NilError(t, err)
			}

			sip, err := svc.UpdateSIP(ctx, tt.args.sipUUID, tt.args.updater)
			if tt.wantErr != "" {
				assert.Error(t, err, tt.wantErr)
				return
			}
			assert.NilError(t, err)

			assert.DeepEqual(t, sip, tt.want,
				cmpopts.EquateApproxTime(time.Millisecond*100),
				cmpopts.IgnoreUnexported(db.SIP{}, db.SIPEdges{}),
			)
		})
	}
}

func TestDeleteSIP(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name    string
		id      int
		wantErr string
	}{
		{
			name: "Deletes a SIP",
		},
		{
			name:    "Fails to delete a missing SIP",
			id:      12345,
			wantErr: "not found error: db: sip not found: delete SIP",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			client, svc := setUpClient(t, logr.Discard())
			ctx := t.Context()

			sip := &datatypes.SIP{
				UUID:   sipUUID,
				Name:   "Test SIP",
				Status: enums.SIPStatusQueued,
			}

			err := svc.CreateSIP(ctx, sip)
			assert.NilError(t, err)

			if tc.id == 0 {
				tc.id = sip.ID
			}

			err = svc.DeleteSIP(ctx, tc.id)
			if tc.wantErr != "" {
				assert.Error(t, err, tc.wantErr)
				return
			}
			assert.NilError(t, err)

			_, err = client.SIP.Get(ctx, tc.id)
			assert.Error(t, err, "db: sip not found")
		})
	}
}

func TestReadSIP(t *testing.T) {
	t.Parallel()

	uploaderID := uuid.New()

	for _, tt := range []struct {
		name    string
		sipUUID uuid.UUID
		want    *datatypes.SIP
		wantErr string
	}{
		{
			name:    "Reads a SIP",
			sipUUID: sipUUID,
			want: &datatypes.SIP{
				ID:          1,
				UUID:        sipUUID,
				Name:        "Test SIP",
				Status:      enums.SIPStatusError,
				AIPID:       uuid.NullUUID{UUID: uuid.New(), Valid: true},
				CreatedAt:   time.Now(),
				StartedAt:   sql.NullTime{Time: time.Now().Add(time.Second), Valid: true},
				CompletedAt: sql.NullTime{Time: time.Now().Add(time.Minute), Valid: true},
				FailedAs:    enums.SIPFailedAsPIP,
				FailedKey:   "failed-key",
				Uploader: &datatypes.Uploader{
					UUID:  uploaderID,
					Email: "nobody@example.com",
					Name:  "Test User",
				},
			},
		},
		{
			name:    "Fails to read a missing SIP",
			sipUUID: sipUUID,
			wantErr: "not found error: db: sip not found",
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			entc, svc := setUpClient(t, logr.Discard())
			ctx := t.Context()

			if tt.want != nil {
				user, err := createUser(t, entc, uploaderID)
				assert.NilError(t, err)

				_, err = entc.SIP.Create().
					SetUUID(tt.want.UUID).
					SetName(tt.want.Name).
					SetStatus(tt.want.Status).
					SetAipID(tt.want.AIPID.UUID).
					SetCreatedAt(tt.want.CreatedAt).
					SetStartedAt(tt.want.StartedAt.Time).
					SetCompletedAt(tt.want.CompletedAt.Time).
					SetFailedAs(tt.want.FailedAs).
					SetFailedKey(tt.want.FailedKey).
					SetUser(user).
					Save(ctx)
				assert.NilError(t, err)
			}

			s, err := svc.ReadSIP(ctx, tt.sipUUID)
			if tt.wantErr != "" {
				assert.Error(t, err, tt.wantErr)
				return
			}
			assert.NilError(t, err)
			assert.DeepEqual(t, s, tt.want)
		})
	}
}

func TestListSIPs(t *testing.T) {
	t.Parallel()

	sipUUID := uuid.New()
	sipUUID2 := uuid.New()
	sipUUID3 := uuid.New()
	aipID := uuid.NullUUID{
		UUID:  uuid.MustParse("e2ace0da-8697-453d-9ea1-4c9b62309e54"),
		Valid: true,
	}
	aipID2 := uuid.NullUUID{
		UUID:  uuid.MustParse("7d085541-af56-4444-9ce2-d6401ff4c97b"),
		Valid: true,
	}
	uploaderID := uuid.New()

	started := sql.NullTime{
		Time: func() time.Time {
			t, _ := time.Parse(time.RFC3339, "2024-09-25T09:31:11Z")
			return t
		}(),
		Valid: true,
	}
	started2 := sql.NullTime{
		Time: func() time.Time {
			t, _ := time.Parse(time.RFC3339, "2024-09-25T10:03:42Z")
			return t
		}(),
		Valid: true,
	}

	completed := sql.NullTime{Time: started.Time.Add(time.Second), Valid: true}
	completed2 := sql.NullTime{Time: started2.Time.Add(time.Second), Valid: true}

	type results struct {
		data []*datatypes.SIP
		page *persistence.Page
	}
	tests := []struct {
		name      string
		data      []*datatypes.SIP
		sipFilter *persistence.SIPFilter
		want      results
		wantErr   string
	}{
		{
			name: "Returns all SIPs",
			data: []*datatypes.SIP{
				{
					UUID:        sipUUID,
					Name:        "Test SIP 1",
					AIPID:       aipID,
					Status:      enums.SIPStatusIngested,
					StartedAt:   started,
					CompletedAt: completed,
				},
				{
					UUID:        sipUUID2,
					Name:        "Test SIP 2",
					AIPID:       aipID2,
					Status:      enums.SIPStatusProcessing,
					StartedAt:   started2,
					CompletedAt: completed2,
					Uploader:    &datatypes.Uploader{UUID: uploaderID},
				},
				{
					UUID:        sipUUID3,
					Name:        "Test SIP 3",
					Status:      enums.SIPStatusError,
					StartedAt:   started2,
					CompletedAt: completed2,
					FailedAs:    enums.SIPFailedAsPIP,
					FailedKey:   "failed-key",
				},
			},
			want: results{
				data: []*datatypes.SIP{
					{
						ID:          1,
						UUID:        sipUUID,
						Name:        "Test SIP 1",
						AIPID:       aipID,
						Status:      enums.SIPStatusIngested,
						CreatedAt:   time.Now(),
						StartedAt:   started,
						CompletedAt: completed,
					},
					{
						ID:          2,
						UUID:        sipUUID2,
						Name:        "Test SIP 2",
						AIPID:       aipID2,
						Status:      enums.SIPStatusProcessing,
						CreatedAt:   time.Now(),
						StartedAt:   started2,
						CompletedAt: completed2,
						Uploader: &datatypes.Uploader{
							UUID:  uploaderID,
							Email: "nobody@example.com",
							Name:  "Test User",
						},
					},
					{
						ID:          3,
						UUID:        sipUUID3,
						Name:        "Test SIP 3",
						Status:      enums.SIPStatusError,
						CreatedAt:   time.Now(),
						StartedAt:   started2,
						CompletedAt: completed2,
						FailedAs:    enums.SIPFailedAsPIP,
						FailedKey:   "failed-key",
					},
				},
				page: &persistence.Page{
					Limit: entfilter.DefaultPageSize,
					Total: 3,
				},
			},
		},
		{
			name: "Returns first page of SIPs",
			data: []*datatypes.SIP{
				{
					UUID:        sipUUID,
					Name:        "Test SIP 1",
					AIPID:       aipID,
					Status:      enums.SIPStatusIngested,
					StartedAt:   started,
					CompletedAt: completed,
				},
				{
					UUID:        sipUUID2,
					Name:        "Test SIP 2",
					AIPID:       aipID2,
					Status:      enums.SIPStatusProcessing,
					StartedAt:   started2,
					CompletedAt: completed2,
				},
			},
			sipFilter: &persistence.SIPFilter{
				Page: persistence.Page{Limit: 1},
			},
			want: results{
				data: []*datatypes.SIP{
					{
						ID:          1,
						UUID:        sipUUID,
						Name:        "Test SIP 1",
						AIPID:       aipID,
						Status:      enums.SIPStatusIngested,
						CreatedAt:   time.Now(),
						StartedAt:   started,
						CompletedAt: completed,
					},
				},
				page: &persistence.Page{
					Limit: 1,
					Total: 2,
				},
			},
		},
		{
			name: "Returns second page of SIPs",
			data: []*datatypes.SIP{
				{
					UUID:        sipUUID,
					Name:        "Test SIP 1",
					AIPID:       aipID,
					Status:      enums.SIPStatusIngested,
					StartedAt:   started,
					CompletedAt: completed,
				},
				{
					UUID:        sipUUID2,
					Name:        "Test SIP 2",
					AIPID:       aipID2,
					Status:      enums.SIPStatusProcessing,
					StartedAt:   started2,
					CompletedAt: completed2,
					Uploader:    &datatypes.Uploader{UUID: uploaderID},
				},
			},
			sipFilter: &persistence.SIPFilter{
				Page: persistence.Page{Limit: 1, Offset: 1},
			},
			want: results{
				data: []*datatypes.SIP{
					{
						ID:          2,
						UUID:        sipUUID2,
						Name:        "Test SIP 2",
						AIPID:       aipID2,
						Status:      enums.SIPStatusProcessing,
						CreatedAt:   time.Now(),
						StartedAt:   started2,
						CompletedAt: completed2,
						Uploader: &datatypes.Uploader{
							UUID:  uploaderID,
							Email: "nobody@example.com",
							Name:  "Test User",
						},
					},
				},
				page: &persistence.Page{
					Limit:  1,
					Offset: 1,
					Total:  2,
				},
			},
		},
		{
			name: "Returns SIPs whose names contain a string",
			data: []*datatypes.SIP{
				{
					UUID:        sipUUID,
					Name:        "Test SIP",
					AIPID:       aipID,
					Status:      enums.SIPStatusIngested,
					StartedAt:   started,
					CompletedAt: completed,
				},
				{
					UUID:        sipUUID2,
					Name:        "small.zip",
					AIPID:       aipID2,
					Status:      enums.SIPStatusProcessing,
					StartedAt:   started2,
					CompletedAt: completed2,
				},
			},
			sipFilter: &persistence.SIPFilter{
				Name: ref.New("small"),
			},
			want: results{
				data: []*datatypes.SIP{
					{
						ID:          2,
						UUID:        sipUUID2,
						Name:        "small.zip",
						AIPID:       aipID2,
						Status:      enums.SIPStatusProcessing,
						CreatedAt:   time.Now(),
						StartedAt:   started2,
						CompletedAt: completed2,
					},
				},
				page: &persistence.Page{
					Limit: entfilter.DefaultPageSize,
					Total: 1,
				},
			},
		},
		{
			name: "Returns SIPs filtered by AIPID",
			data: []*datatypes.SIP{
				{
					UUID:        sipUUID,
					Name:        "Test SIP 1",
					AIPID:       aipID,
					Status:      enums.SIPStatusIngested,
					StartedAt:   started,
					CompletedAt: completed,
				},
				{
					UUID:        sipUUID2,
					Name:        "Test SIP 2",
					AIPID:       aipID2,
					Status:      enums.SIPStatusProcessing,
					StartedAt:   started2,
					CompletedAt: completed2,
				},
			},
			sipFilter: &persistence.SIPFilter{
				AIPID: &aipID2.UUID,
			},
			want: results{
				data: []*datatypes.SIP{
					{
						ID:          2,
						UUID:        sipUUID2,
						Name:        "Test SIP 2",
						AIPID:       aipID2,
						Status:      enums.SIPStatusProcessing,
						CreatedAt:   time.Now(),
						StartedAt:   started2,
						CompletedAt: completed2,
					},
				},
				page: &persistence.Page{
					Limit: entfilter.DefaultPageSize,
					Total: 1,
				},
			},
		},
		{
			name: "Returns SIPs filtered by status",
			data: []*datatypes.SIP{
				{
					UUID:        sipUUID,
					Name:        "Test SIP 1",
					AIPID:       aipID,
					Status:      enums.SIPStatusIngested,
					StartedAt:   started,
					CompletedAt: completed,
				},
				{
					UUID:        sipUUID2,
					Name:        "Test SIP 2",
					AIPID:       aipID2,
					Status:      enums.SIPStatusProcessing,
					StartedAt:   started2,
					CompletedAt: completed2,
				},
			},
			sipFilter: &persistence.SIPFilter{
				Status: ref.New(enums.SIPStatusProcessing),
			},
			want: results{
				data: []*datatypes.SIP{
					{
						ID:          2,
						UUID:        sipUUID2,
						Name:        "Test SIP 2",
						AIPID:       aipID2,
						Status:      enums.SIPStatusProcessing,
						CreatedAt:   time.Now(),
						StartedAt:   started2,
						CompletedAt: completed2,
					},
				},
				page: &persistence.Page{
					Limit: entfilter.DefaultPageSize,
					Total: 1,
				},
			},
		},
		{
			name: "Returns SIPs filtered by CreatedAt",
			data: []*datatypes.SIP{
				{
					UUID:        sipUUID,
					Name:        "Test SIP 1",
					AIPID:       aipID,
					Status:      enums.SIPStatusIngested,
					StartedAt:   started,
					CompletedAt: completed,
				},
				{
					UUID:        sipUUID2,
					Name:        "Test SIP 2",
					AIPID:       aipID2,
					Status:      enums.SIPStatusProcessing,
					StartedAt:   started2,
					CompletedAt: completed2,
				},
			},
			sipFilter: &persistence.SIPFilter{
				CreatedAt: func(t *testing.T) *timerange.Range {
					r, err := timerange.New(
						time.Now().Add(-1*time.Minute),
						time.Now().Add(time.Minute),
					)
					if err != nil {
						t.Fatalf("Error: %v", err)
					}
					return &r
				}(t),
			},
			want: results{
				data: []*datatypes.SIP{
					{
						ID:          1,
						UUID:        sipUUID,
						Name:        "Test SIP 1",
						AIPID:       aipID,
						Status:      enums.SIPStatusIngested,
						CreatedAt:   time.Now(),
						StartedAt:   started,
						CompletedAt: completed,
					},
					{
						ID:          2,
						UUID:        sipUUID2,
						Name:        "Test SIP 2",
						AIPID:       aipID2,
						Status:      enums.SIPStatusProcessing,
						CreatedAt:   time.Now(),
						StartedAt:   started2,
						CompletedAt: completed2,
					},
				},
				page: &persistence.Page{
					Limit: entfilter.DefaultPageSize,
					Total: 2,
				},
			},
		},
		{
			name: "Returns no results when no SIPs match CreatedAt range",
			data: []*datatypes.SIP{
				{
					UUID:        sipUUID,
					Name:        "Test SIP 1",
					AIPID:       aipID,
					Status:      enums.SIPStatusIngested,
					StartedAt:   started,
					CompletedAt: completed,
				},
				{
					UUID:        sipUUID2,
					Name:        "Test SIP 2",
					AIPID:       aipID2,
					Status:      enums.SIPStatusProcessing,
					StartedAt:   started2,
					CompletedAt: completed2,
				},
			},
			sipFilter: &persistence.SIPFilter{
				CreatedAt: func(t *testing.T) *timerange.Range {
					r, err := timerange.New(
						time.Now().Add(time.Minute),
						time.Now().Add(2*time.Minute),
					)
					if err != nil {
						t.Fatalf("Error: %v", err)
					}
					return &r
				}(t),
			},
			want: results{
				data: []*datatypes.SIP{},
				page: &persistence.Page{
					Limit: entfilter.DefaultPageSize,
					Total: 0,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			entc, svc := setUpClient(t, logr.Discard())
			ctx := t.Context()

			if len(tt.data) > 0 {
				for _, sip := range tt.data {
					q := entc.SIP.Create().
						SetUUID(sip.UUID).
						SetName(sip.Name).
						SetStatus(sip.Status).
						SetCreatedAt(time.Now()).
						SetStartedAt(sip.StartedAt.Time).
						SetCompletedAt(sip.CompletedAt.Time)

					if sip.AIPID.Valid {
						q.SetAipID(sip.AIPID.UUID)
					}
					if sip.FailedAs.IsValid() {
						q.SetFailedAs(sip.FailedAs)
					}
					if sip.FailedKey != "" {
						q.SetFailedKey(sip.FailedKey)
					}
					if sip.Uploader != nil {
						user, err := createUser(t, entc, uploaderID)
						assert.NilError(t, err)
						q.SetUser(user)
					}

					_, err := q.Save(ctx)
					assert.NilError(t, err)
				}
			}

			got, pg, err := svc.ListSIPs(ctx, tt.sipFilter)
			assert.NilError(t, err)

			assert.DeepEqual(t, got, tt.want.data,
				cmpopts.EquateApproxTime(time.Millisecond*100),
				cmpopts.IgnoreUnexported(db.SIP{}, db.SIPEdges{}),
			)
			assert.DeepEqual(t, pg, tt.want.page)
		})
	}
}
