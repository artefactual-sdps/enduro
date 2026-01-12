package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"reflect"
	"strings"
	"time"

	"github.com/artefactual-sdps/temporal-activities/archiveextract"
	"github.com/artefactual-sdps/temporal-activities/bagcreate"
	"github.com/google/uuid"
	"github.com/mitchellh/mapstructure"
	"github.com/spf13/viper"

	"github.com/artefactual-sdps/enduro/internal/a3m"
	"github.com/artefactual-sdps/enduro/internal/am"
	"github.com/artefactual-sdps/enduro/internal/api"
	"github.com/artefactual-sdps/enduro/internal/auditlog"
	"github.com/artefactual-sdps/enduro/internal/batch"
	"github.com/artefactual-sdps/enduro/internal/db"
	"github.com/artefactual-sdps/enduro/internal/event"
	"github.com/artefactual-sdps/enduro/internal/ingest"
	"github.com/artefactual-sdps/enduro/internal/poststorage"
	"github.com/artefactual-sdps/enduro/internal/premis"
	"github.com/artefactual-sdps/enduro/internal/preprocessing"
	"github.com/artefactual-sdps/enduro/internal/pres"
	"github.com/artefactual-sdps/enduro/internal/sipsource"
	"github.com/artefactual-sdps/enduro/internal/storage"
	"github.com/artefactual-sdps/enduro/internal/telemetry"
	"github.com/artefactual-sdps/enduro/internal/temporal"
	"github.com/artefactual-sdps/enduro/internal/watcher"
)

type ConfigurationValidator interface {
	Validate() error
}

type Configuration struct {
	// Debug toggles the encoding of log messages to support different reader
	// contexts.
	//
	// If Debug is true, the logger will output human readable logs with ANSI
	// color codes. This is useful for debugging and development.
	//
	// If Debug is false the logger will output JSON formatted messages with no
	// color codes. This is useful for data analysis and log aggregation.
	Debug bool

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
	Batch           batch.Config
	Database        db.Config
	Event           event.Config
	ExtractActivity archiveextract.Config
	Poststorage     []poststorage.Config
	Preprocessing   preprocessing.Config
	Preservation    pres.Config
	SIPSource       sipsource.Config
	Storage         storage.Config
	Temporal        temporal.Config
	InternalStorage InternalStorageConfig
	Upload          ingest.UploadConfig
	Watcher         watcher.Config
	Telemetry       telemetry.Config
	ValidatePREMIS  premis.Config
	Auditlog        auditlog.Config
}

func (c *Configuration) Validate() error {
	return errors.Join(
		c.InternalStorage.Validate(),
		c.A3m.Validate(),
		c.API.Auth.Validate(),
		c.BagIt.Validate(),
		c.Preprocessing.Validate(),
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
	v.SetDefault("debugListen", "127.0.0.1:9001")
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
func setCORSOriginEnv(config *Configuration) error {
	if config.API.CORSOrigin == "" {
		// Default to the API URI to disallow all cross-origin requests.
		config.API.CORSOrigin = config.API.Listen
	}

	if err := os.Setenv("ENDURO_API_CORS_ORIGIN", config.API.CORSOrigin); err != nil {
		return err
	}

	return nil
}

// stringToUUIDHookFunc decodes a string to a uuid.UUID. Copied from
// https://github.com/go-saas/kit/blob/main/pkg/mapstructure/mapstructure.go
func stringToUUIDHookFunc() mapstructure.DecodeHookFunc {
	return func(f, t reflect.Type, data any) (any, error) {
		if f.Kind() != reflect.String || t != reflect.TypeOf(uuid.UUID{}) {
			return data, nil
		}

		return uuid.Parse(data.(string))
	}
}

// stringToMapHookFunc decodes a JSON string to a map[string][]string.
func stringToMapHookFunc() mapstructure.DecodeHookFunc {
	return func(f, t reflect.Type, data any) (any, error) {
		value := map[string][]string{}
		if f.Kind() != reflect.String || t != reflect.TypeOf(value) {
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
