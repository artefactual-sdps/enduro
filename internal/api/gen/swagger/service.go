// Code generated by goa v3.15.2, DO NOT EDIT.
//
// swagger service
//
// Command:
// $ goa gen github.com/artefactual-sdps/enduro/internal/api/design -o
// internal/api

package swagger

// The swagger service serves the API swagger definition.
type Service interface {
}

// APIName is the name of the API as defined in the design.
const APIName = "enduro"

// APIVersion is the version of the API as defined in the design.
const APIVersion = "0.0.1"

// ServiceName is the name of the service as defined in the design. This is the
// same value that is set in the endpoint request contexts under the ServiceKey
// key.
const ServiceName = "swagger"

// MethodNames lists the service method names as defined in the design. These
// are the same values that are set in the endpoint request contexts under the
// MethodKey key.
var MethodNames = [0]string{}

type MonitorPingEvent struct {
	Message *string
}

// SIP describes an ingest SIP type.
type SIP struct {
	// Identifier of SIP
	ID uint
	// Name of the SIP
	Name *string
	// Status of the SIP
	Status string
	// Identifier of AIP
	AipID *string
	// Creation datetime
	CreatedAt string
	// Start datetime
	StartedAt *string
	// Completion datetime
	CompletedAt *string
}

type SIPCreatedEvent struct {
	// Identifier of SIP
	ID   uint
	Item *SIP
}

type SIPStatusUpdatedEvent struct {
	// Identifier of SIP
	ID     uint
	Status string
}

// SIPTask describes a SIP workflow task.
type SIPTask struct {
	ID          uint
	TaskID      string
	Name        string
	Status      string
	StartedAt   string
	CompletedAt *string
	Note        *string
	WorkflowID  *uint
}

type SIPTaskCollection []*SIPTask

type SIPTaskCreatedEvent struct {
	// Identifier of task
	ID   uint
	Item *SIPTask
}

type SIPTaskUpdatedEvent struct {
	// Identifier of task
	ID   uint
	Item *SIPTask
}

type SIPUpdatedEvent struct {
	// Identifier of SIP
	ID   uint
	Item *SIP
}

// SIPWorkflow describes a workflow of a SIP.
type SIPWorkflow struct {
	ID          uint
	TemporalID  string
	Type        string
	Status      string
	StartedAt   string
	CompletedAt *string
	Tasks       SIPTaskCollection
	SipID       *uint
}

type SIPWorkflowCreatedEvent struct {
	// Identifier of workflow
	ID   uint
	Item *SIPWorkflow
}

type SIPWorkflowUpdatedEvent struct {
	// Identifier of workflow
	ID   uint
	Item *SIPWorkflow
}
