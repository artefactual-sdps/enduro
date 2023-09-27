package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/viper"

	"github.com/artefactual-sdps/enduro/internal/a3m"
	"github.com/artefactual-sdps/enduro/internal/am"
	"github.com/artefactual-sdps/enduro/internal/api"
	"github.com/artefactual-sdps/enduro/internal/db"
	"github.com/artefactual-sdps/enduro/internal/event"
	"github.com/artefactual-sdps/enduro/internal/storage"
	"github.com/artefactual-sdps/enduro/internal/temporal"
	"github.com/artefactual-sdps/enduro/internal/upload"
	"github.com/artefactual-sdps/enduro/internal/watcher"
)

type ConfigurationValidator interface {
	Validate() error
}

type Configuration struct {
	Verbosity   int
	Debug       bool
	DebugListen string
	API         api.Config
	Event       event.Config
	Database    db.Config
	Temporal    temporal.Config
	Watcher     watcher.Config
	Storage     storage.Config
	Upload      upload.Config
	A3m         a3m.Config
	Am          am.Config
}

func (c Configuration) Validate() error {
	// TODO: should this validate all the fields in Configuration?
	if config, ok := interface{}(c.Upload).(ConfigurationValidator); ok {
		err := config.Validate()
		if err != nil {
			return err
		}
	}
	if config, ok := interface{}(c.API.Auth).(ConfigurationValidator); ok {
		err := config.Validate()
		if err != nil {
			return err
		}
	}
	return nil
}

func Read(config *Configuration, configFile string) (found bool, configFileUsed string, err error) {
	v := viper.New()

	v.AddConfigPath(".")
	v.AddConfigPath("$HOME/.config/")
	v.AddConfigPath("/etc")
	v.SetConfigName("enduro")
	v.SetDefault("api.processing", a3m.ProcessingDefault)
	v.SetDefault("debugListen", "127.0.0.1:9001")
	v.SetDefault("api.listen", "127.0.0.1:9000")
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

	err = v.Unmarshal(config)
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
