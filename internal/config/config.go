package config

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"reflect"
	"strings"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/container"
	"github.com/artefactual-sdps/temporal-activities/archiveextract"
	"github.com/artefactual-sdps/temporal-activities/bagcreate"
	"github.com/google/uuid"
	"github.com/mitchellh/mapstructure"
	"github.com/spf13/viper"
	"go.artefactual.dev/tools/bucket"
	"gocloud.dev/blob"
	"gocloud.dev/blob/azureblob"

	"github.com/artefactual-sdps/enduro/internal/a3m"
	"github.com/artefactual-sdps/enduro/internal/am"
	"github.com/artefactual-sdps/enduro/internal/api"
	"github.com/artefactual-sdps/enduro/internal/db"
	"github.com/artefactual-sdps/enduro/internal/event"
	"github.com/artefactual-sdps/enduro/internal/ingest"
	"github.com/artefactual-sdps/enduro/internal/poststorage"
	"github.com/artefactual-sdps/enduro/internal/premis"
	"github.com/artefactual-sdps/enduro/internal/preprocessing"
	"github.com/artefactual-sdps/enduro/internal/pres"
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
	Database        db.Config
	Event           event.Config
	ExtractActivity archiveextract.Config
	Poststorage     []poststorage.Config
	Preprocessing   preprocessing.Config
	Preservation    pres.Config
	Storage         storage.Config
	Temporal        temporal.Config
	InternalStorage InternalStorageConfig
	Upload          ingest.UploadConfig
	Watcher         watcher.Config
	Telemetry       telemetry.Config
	ValidatePREMIS  premis.Config
}

type InternalStorageConfig struct {
	Bucket bucket.Config
	Azure  Azure
}

type Azure struct {
	StorageAccount string
	StorageKey     string
}

func (u *InternalStorageConfig) OpenBucket(ctx context.Context) (*blob.Bucket, error) {
	if u.Azure.StorageAccount != "" && strings.HasPrefix(u.Bucket.URL, "azblob") {
		makeClient := func(svcURL azureblob.ServiceURL, containerName azureblob.ContainerName) (*container.Client, error) {
			sharedKeyCredential, err := container.NewSharedKeyCredential(u.Azure.StorageAccount, u.Azure.StorageKey)
			if err != nil {
				return nil, err
			}

			containerURL := fmt.Sprintf("%s/%s", svcURL, containerName)
			sharedKeyCredentialClient, err := container.NewClientWithSharedKeyCredential(
				containerURL,
				sharedKeyCredential,
				nil,
			)
			if err != nil {
				return nil, err
			}

			return sharedKeyCredentialClient, nil
		}

		urlOpener := azureblob.URLOpener{
			MakeClient: makeClient,
			ServiceURLOptions: azureblob.ServiceURLOptions{
				AccountName: u.Azure.StorageAccount,
			},
		}

		urlMux := new(blob.URLMux)
		urlMux.RegisterBucket(azureblob.Scheme, &urlOpener)
		return urlMux.OpenBucket(ctx, u.Bucket.URL)
	}

	b, err := bucket.NewWithConfig(ctx, &u.Bucket)
	if err != nil {
		return nil, err
	}
	return b, nil
}

func (c *Configuration) Validate() error {
	var err error
	if c.InternalStorage.Bucket.URL != "" &&
		(c.InternalStorage.Bucket.Bucket != "" || c.InternalStorage.Bucket.Region != "") {
		err = errors.New("the [internalBucket] URL option and the other configuration options are mutually exclusive")
	} else if strings.HasPrefix(c.InternalStorage.Bucket.URL, "azblob") {
		if c.InternalStorage.Azure.StorageAccount == "" || c.InternalStorage.Azure.StorageKey == "" {
			err = errors.New("the [internalBucket] Azure credentials are undefined")
		}
	} else if c.InternalStorage.Azure.StorageAccount != "" && !strings.HasPrefix(c.InternalStorage.Bucket.URL, "azblob://") {
		err = errors.New("the [internalBucket] URL Azure option is invalid, should be in the form azblob://my-bucket")
	}

	return errors.Join(
		err,
		c.A3m.Validate(),
		c.API.Auth.Validate(),
		c.BagIt.Validate(),
		c.Preprocessing.Validate(),
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
	v.SetDefault("a3m.capacity", 1)
	v.SetDefault("a3m.processing", a3m.ProcessingDefault)
	v.SetDefault("am.capacity", 1)
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
