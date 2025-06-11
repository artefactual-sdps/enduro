package config

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/container"
	"go.artefactual.dev/tools/bucket"
	"gocloud.dev/blob"
	"gocloud.dev/blob/azureblob"
)

type InternalStorageConfig struct {
	Bucket bucket.Config
	Azure  Azure
}

type Azure struct {
	StorageAccount string
	StorageKey     string
}

func (isc *InternalStorageConfig) OpenBucket(ctx context.Context) (*blob.Bucket, error) {
	if isc.Azure.StorageAccount != "" && isc.Azure.StorageKey != "" && strings.HasPrefix(isc.Bucket.URL, "azblob") {
		makeClient := func(svcURL azureblob.ServiceURL, containerName azureblob.ContainerName) (*container.Client, error) {
			sharedKeyCredential, err := container.NewSharedKeyCredential(isc.Azure.StorageAccount, isc.Azure.StorageKey)
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
				AccountName: isc.Azure.StorageAccount,
			},
		}

		urlMux := new(blob.URLMux)
		urlMux.RegisterBucket(azureblob.Scheme, &urlOpener)
		return urlMux.OpenBucket(ctx, isc.Bucket.URL)
	}

	b, err := bucket.NewWithConfig(ctx, &isc.Bucket)
	if err != nil {
		return nil, err
	}
	return b, nil
}

func (isc *InternalStorageConfig) Validate() error {
	var err error
	if isc.Bucket.URL != "" &&
		(isc.Bucket.Bucket != "" || isc.Bucket.Region != "") {
		err = errors.New("the [internalStorage] URL option and the other configuration options are mutually exclusive")
	} else if strings.HasPrefix(isc.Bucket.URL, "azblob") &&
		(isc.Azure.StorageAccount == "" || isc.Azure.StorageKey == "") {
		err = errors.New("the [internalStorage] Azure credentials are undefined")
	}
	return err
}
