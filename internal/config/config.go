package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"reflect"
	"slices"
	"strings"
	"time"

	"github.com/artefactual-sdps/temporal-activities/archiveextract"
	"github.com/artefactual-sdps/temporal-activities/bagcreate"
	"github.com/google/uuid"
	"github.com/mitchellh/mapstructure"
	"github.com/spf13/viper"
	"go.artefactual.dev/tools/bucket"
	"go.artefactual.dev/tools/log"

	"github.com/artefactual-sdps/enduro/internal/a3m"
	"github.com/artefactual-sdps/enduro/internal/am"
	"github.com/artefactual-sdps/enduro/internal/api"
	"github.com/artefactual-sdps/enduro/internal/auditlog"
	"github.com/artefactual-sdps/enduro/internal/bagit"
	"github.com/artefactual-sdps/enduro/internal/childwf"
	"github.com/artefactual-sdps/enduro/internal/db"
	"github.com/artefactual-sdps/enduro/internal/event"
	"github.com/artefactual-sdps/enduro/internal/ingest"
	"github.com/artefactual-sdps/enduro/internal/premis"
	"github.com/artefactual-sdps/enduro/internal/pres"
	"github.com/artefactual-sdps/enduro/internal/sipsource"
	"github.com/artefactual-sdps/enduro/internal/storage"
	"github.com/artefactual-sdps/enduro/internal/telemetry"
	"github.com/artefactual-sdps/enduro/internal/temporal"
	"github.com/artefactual-sdps/enduro/internal/watcher"
)

var logLevels = []string{
	"debug",
	"info",
	"warn",
	"error",
}

type ConfigurationValidator interface {
	Validate() error
}

type LogFormat string

const (
	LogFormatJSON LogFormat = "json"
	LogFormatText LogFormat = "text"
)

func (f LogFormat) Validate() error {
	switch f {
	case LogFormatJSON, LogFormatText:
		return nil
	default:
		return fmt.Errorf("LogFormat: unsupported value %q (use %q or %q)", f, LogFormatJSON, LogFormatText)
	}
}

// LoggerFormat returns the corresponding application logger format.
func (f LogFormat) LoggerFormat() log.Format {
	switch f {
	case LogFormatJSON:
		return log.FormatJSON
	case LogFormatText:
		return log.FormatText
	default:
		panic(fmt.Sprintf("config: invalid log format %q", f))
	}
}

type Configuration struct {
	// LogFormat controls the encoding of application log messages. Supported
	// values are "json" for structured output and "text" for human-readable,
	// colorized output.
	LogFormat LogFormat

	// DebugListen is the HTTP address of the observability server.
	DebugListen string

	// Verbosity controls the verbosity of log messages. The default is 0 which
	// will only log the most important messages. The development environment
	// log level is 2 which will log most messages. See the developer
	// documentation for more information on logging levels.
	Verbosity int

	A3m             a3m.Config
	AM              am.Config
	InternalAPI     api.Config
	API             api.Config
	BagIt           bagcreate.Config
	BagItValidator  bagit.ValidatorConfig
	ChildWorkflows  childwf.Configs
	Database        db.Config
	Event           event.Config
	ExtractActivity archiveextract.Config
	Ingest          ingest.Config
	Preservation    pres.Config
	SIPSource       sipsource.Config
	Storage         storage.Config
	Temporal        temporal.Config
	InternalStorage bucket.Config
	Upload          ingest.UploadConfig
	Watcher         watcher.Config
	Telemetry       telemetry.Config
	ValidatePREMIS  premis.Config
	Auditlog        auditlog.Config
}

func (c *Configuration) Validate() error {
	return errors.Join(
		c.LogFormat.Validate(),
		c.A3m.Validate(),
		c.API.Validate(),
		c.InternalAPI.Validate(),
		c.BagIt.Validate(),
		c.BagItValidator.Validate(),
		c.ChildWorkflows.Validate(),
		c.Ingest.Validate(),
		c.SIPSource.Validate(),
		c.ValidatePREMIS.Validate(),
		c.Watcher.Validate(),
	)
}

func Read(config *Configuration, configFile string) (found bool, configFileUsed string, err error) {
	v := viper.New()

	v.AddConfigPath(".")
	v.AddConfigPath("$HOME/.config/")
	v.AddConfigPath("/etc")
	v.SetConfigName("enduro")
	v.SetDefault("a3m.processing", a3m.ProcessingDefault)
	v.SetDefault("am.capacity", 20)
	v.SetDefault("am.pollInterval", 10*time.Second)
	v.SetDefault("api.listen", "127.0.0.1:9000")
	v.SetDefault("bagitvalidator.poolSize", 1)
	v.SetDefault("debugListen", "127.0.0.1:9001")
	v.SetDefault("logFormat", LogFormatJSON)
	v.SetDefault("preservation.taskqueue", temporal.A3mWorkerTaskQueue)
	v.SetDefault("storage.taskqueue", temporal.GlobalTaskQueue)
	v.SetDefault("temporal.taskqueue", temporal.GlobalTaskQueue)
	v.SetDefault("upload.maxSize", 4294967296)
	v.SetEnvPrefix("enduro")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	if configFile != "" {
		v.SetConfigFile(configFile)
	}

	err = v.ReadInConfig()
	_, ok := err.(viper.ConfigFileNotFoundError)
	if !ok {
		found = true
	}
	if found && err != nil {
		return found, configFileUsed, fmt.Errorf("failed to read configuration file: %w", err)
	}

	decodeHookFunc := mapstructure.ComposeDecodeHookFunc(
		// These are the viper DecodeHookFunc defaults.
		mapstructure.StringToTimeDurationHookFunc(),
		mapstructure.StringToSliceHookFunc(","),
		// StringToUUIDHookFunc is a custom string to UUID decoder.
		stringToUUIDHookFunc(),
		stringToMapHookFunc(),
		stringToLogLevelHookFunc(),
	)

	err = v.Unmarshal(config, viper.DecodeHook(decodeHookFunc))
	if err != nil {
		return found, configFileUsed, fmt.Errorf("failed to unmarshal configuration: %w", err)
	}

	if err := config.Validate(); err != nil {
		return found, configFileUsed, fmt.Errorf("failed to validate the provided config: %w", err)
	}

	configFileUsed = v.ConfigFileUsed()

	if err := setCORSOriginEnv(config); err != nil {
		return found, configFileUsed, fmt.Errorf(
			"failed to set CORS Origin environment variable: %w", err,
		)
	}

	return found, configFileUsed, nil
}

// setCORSOriginEnv sets the CORS Origin environment variable needed by
// Goa-generated code for the API.
func setCORSOriginEnv(cfg *Configuration) error {
	if err := os.Setenv("ENDURO_API_CORS_ORIGIN", cfg.API.CORSOrigin); err != nil {
		return err
	}

	return nil
}

// stringToUUIDHookFunc decodes a string to a uuid.UUID. Copied from
// https://github.com/go-saas/kit/blob/main/pkg/mapstructure/mapstructure.go
func stringToUUIDHookFunc() mapstructure.DecodeHookFunc {
	return func(f, t reflect.Type, data any) (any, error) {
		if f.Kind() != reflect.String || t != reflect.TypeFor[uuid.UUID]() {
			return data, nil
		}

		return uuid.Parse(data.(string))
	}
}

// stringToMapHookFunc decodes a JSON string to a map[string][]string.
func stringToMapHookFunc() mapstructure.DecodeHookFunc {
	return func(f, t reflect.Type, data any) (any, error) {
		value := map[string][]string{}
		if f.Kind() != reflect.String || t != reflect.TypeFor[map[string][]string]() {
			return data, nil
		}

		if data.(string) != "" {
			if err := json.Unmarshal([]byte(data.(string)), &value); err != nil {
				return nil, err
			}
		}

		return value, nil
	}
}

func stringToLogLevelHookFunc() mapstructure.DecodeHookFunc {
	return func(f, t reflect.Type, data any) (any, error) {
		if f.Kind() != reflect.String || t != reflect.TypeFor[slog.Level]() {
			return data, nil
		}

		name := strings.ToLower(data.(string))
		if slices.Contains(logLevels, name) {
			var lvl slog.Level
			if err := lvl.UnmarshalText([]byte(name)); err != nil {
				return nil, fmt.Errorf("failed to unmarshal log level '%s': %w", data.(string), err)
			}
			return lvl, nil
		} else {
			return nil, fmt.Errorf(
				"invalid log level '%s', valid values are: %s",
				data.(string),
				strings.Join(logLevels, ", "),
			)
		}
	}
}
