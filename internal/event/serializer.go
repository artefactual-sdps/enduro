package event

import (
	"encoding/json"

	ingestclient "github.com/artefactual-sdps/enduro/internal/api/gen/http/ingest/client"
	ingestserver "github.com/artefactual-sdps/enduro/internal/api/gen/http/ingest/server"
	storageclient "github.com/artefactual-sdps/enduro/internal/api/gen/http/storage/client"
	storageserver "github.com/artefactual-sdps/enduro/internal/api/gen/http/storage/server"
	goaingest "github.com/artefactual-sdps/enduro/internal/api/gen/ingest"
	goastorage "github.com/artefactual-sdps/enduro/internal/api/gen/storage"
)

// EventSerializer handles serialization/deserialization of events for Redis.
type EventSerializer[T any] interface {
	Marshal(event T) ([]byte, error)
	Unmarshal(data []byte) (T, error)
}

// IngestEventSerializer handles ingest events.
type IngestEventSerializer struct{}

func (s *IngestEventSerializer) Marshal(event *goaingest.IngestEvent) ([]byte, error) {
	return json.Marshal(ingestserver.NewMonitorResponseBody(event))
}

func (s *IngestEventSerializer) Unmarshal(data []byte) (*goaingest.IngestEvent, error) {
	payload := ingestclient.MonitorResponseBody{}
	if err := json.Unmarshal(data, &payload); err != nil {
		return nil, err
	}
	if err := ingestclient.ValidateMonitorResponseBody(&payload); err != nil {
		return nil, err
	}
	return ingestclient.NewMonitorIngestEventOK(&payload), nil
}

// StorageEventSerializer handles storage events
type StorageEventSerializer struct{}

func (s *StorageEventSerializer) Marshal(event *goastorage.StorageEvent) ([]byte, error) {
	return json.Marshal(storageserver.NewMonitorResponseBody(event))
}

func (s *StorageEventSerializer) Unmarshal(data []byte) (*goastorage.StorageEvent, error) {
	payload := storageclient.MonitorResponseBody{}
	if err := json.Unmarshal(data, &payload); err != nil {
		return nil, err
	}
	if err := storageclient.ValidateMonitorResponseBody(&payload); err != nil {
		return nil, err
	}
	return storageclient.NewMonitorStorageEventOK(&payload), nil
}
