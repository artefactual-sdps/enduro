package ingest

import "github.com/google/uuid"

type Config struct {
	Storage StorageConfig
}

type StorageConfig struct {
	Address                    string
	DefaultPermanentLocationID uuid.UUID
}
