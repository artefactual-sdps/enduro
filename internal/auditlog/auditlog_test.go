package auditlog_test

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/google/uuid"
	"gotest.tools/v3/assert"

	"github.com/artefactual-sdps/enduro/internal/auditlog"
)

func TestNewFromConfig(t *testing.T) {
	t.Parallel()

	resID := uuid.MustParse("e8d32bd5-faa4-4ce1-bb50-55d9c28b306d")

	type test struct {
		name    string
		cfg     auditlog.Config
		event   *auditlog.Event
		want    string
		wantErr error
	}
	for _, tc := range []test{
		{
			name: "audit log disabled when no filepath is configured",
			cfg:  auditlog.Config{},
			event: &auditlog.Event{
				Level:      auditlog.LevelInfo,
				Msg:        "SIP ingest started",
				Type:       "SIP.ingest",
				ResourceID: resID.String(),
				User:       "test@example.com",
			},
			want:    "",
			wantErr: os.ErrNotExist,
		},
		{
			name: "writes audit log",
			cfg: auditlog.Config{
				Filepath: filepath.Join(t.TempDir(), "audit.log"),
				MaxSize:  1, // 1 MB
				Compress: true,
			},
			event: &auditlog.Event{
				Level:      auditlog.LevelInfo,
				Msg:        "SIP ingest started",
				Type:       "SIP.ingest",
				ResourceID: resID.String(),
				User:       "test@example.com",
			},
			want: `"level":"INFO","msg":"SIP ingest started","type":"SIP.ingest","resourceID":"e8d32bd5-faa4-4ce1-bb50-55d9c28b306d","user":"test@example.com"`,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			al := auditlog.NewFromConfig(tc.cfg)
			al.Log(context.Background(), tc.event)
			al.Close()

			got, err := os.ReadFile(tc.cfg.Filepath)
			if tc.wantErr != nil {
				assert.ErrorIs(t, err, tc.wantErr)
				return
			}
			assert.NilError(t, err)

			assert.Assert(t,
				strings.Contains(string(got), tc.want),
				fmt.Sprintf("expected %s to contain %s", string(got), tc.want),
			)
		})
	}
}
