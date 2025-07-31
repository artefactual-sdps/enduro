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

// eventSerializer handles serialization/deserialization of events.
type eventSerializer[T any] interface {
	marshal(event T) ([]byte, error)
	unmarshal(data []byte) (T, error)
}

var (
	_ eventSerializer[*goaingest.IngestEvent]   = (*ingestEventSerializer)(nil)
	_ eventSerializer[*goastorage.StorageEvent] = (*storageEventSerializer)(nil)
)

// ingestEventSerializer handles ingest events.
type ingestEventSerializer struct{}

func (s *ingestEventSerializer) marshal(event *goaingest.IngestEvent) ([]byte, error) {
	return json.Marshal(ingestserver.NewMonitorResponseBody(event))
}

func (s *ingestEventSerializer) unmarshal(data []byte) (*goaingest.IngestEvent, error) {
	payload := ingestclient.MonitorResponseBody{}
	if err := json.Unmarshal(data, &payload); err != nil {
		return nil, err
	}
	if err := ingestclient.ValidateMonitorResponseBody(&payload); err != nil {
		return nil, err
	}
	return ingestclient.NewMonitorIngestEventOK(&payload), nil
}

// storageEventSerializer handles storage events.
type storageEventSerializer struct{}

func (s *storageEventSerializer) marshal(event *goastorage.StorageEvent) ([]byte, error) {
	return json.Marshal(storageserver.NewMonitorResponseBody(event))
}

func (s *storageEventSerializer) unmarshal(data []byte) (*goastorage.StorageEvent, error) {
	payload := storageclient.MonitorResponseBody{}
	if err := json.Unmarshal(data, &payload); err != nil {
		return nil, err
	}
	if err := storageclient.ValidateMonitorResponseBody(&payload); err != nil {
		return nil, err
	}
	return storageclient.NewMonitorStorageEventOK(&payload), nil
}
