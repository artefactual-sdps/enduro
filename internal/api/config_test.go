package api_test

import (
	"log/slog"
	"testing"

	"gotest.tools/v3/assert"

	"github.com/artefactual-sdps/enduro/internal/api"
)

func TestConfigValidate(t *testing.T) {
	type testCase struct {
		name    string
		cfg     *api.Config
		want    *api.Config
		wantErr string
	}

	for _, tc := range []testCase{
		{
			name: "default config",
			cfg:  &api.Config{},
			want: &api.Config{
				Log: api.LogConfig{
					Format: api.LogFormatJSON,
				},
			},
		},
		{
			name: "populated config",
			cfg: &api.Config{
				Listen:     "127.0.0.1:9000",
				CORSOrigin: "http://example.com",
				Log: api.LogConfig{
					Path:   "stdout",
					Level:  slog.LevelWarn,
					Format: api.LogFormatText,
				},
			},
			want: &api.Config{
				Listen:     "127.0.0.1:9000",
				CORSOrigin: "http://example.com",
				Log: api.LogConfig{
					Path:   "stdout",
					Level:  slog.LevelWarn,
					Format: api.LogFormatText,
				},
			},
		},
		{
			name: "invalid log format",
			cfg: &api.Config{
				Log: api.LogConfig{
					Format: "invalid",
				},
			},
			wantErr: `unsupported log format: "invalid", supported formats are "json", "text"`,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.cfg.Validate()
			if tc.wantErr != "" {
				assert.ErrorContains(t, err, tc.wantErr)
				return
			}

			assert.NilError(t, err)
			assert.DeepEqual(t, tc.cfg, tc.want)
		})
	}
}
