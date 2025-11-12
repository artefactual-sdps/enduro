package design

import (
	. "goa.design/goa/v3/dsl" //nolint:staticcheck
)

// This type is used as a StreamingResult, Goa generates two OpenAPI schemas,
// one for the declared type itself and one for its streaming representation.
// To avoid a name clash the latter is suffixed as StorageEvent2.
var StorageEvent = Type("StorageEvent", func() {
	OneOf("value", func() {
		Attribute("storage_ping_event", StoragePingEvent)
		Attribute("location_created_event", LocationCreatedEvent)
		Attribute("aip_created_event", AIPCreatedEvent)
		Attribute("aip_status_updated_event", AIPStatusUpdatedEvent)
		Attribute("aip_location_updated_event", AIPLocationUpdatedEvent)
		Attribute("aip_workflow_created_event", AIPWorkflowCreatedEvent)
		Attribute("aip_workflow_updated_event", AIPWorkflowUpdatedEvent)
		Attribute("aip_task_created_event", AIPTaskCreatedEvent)
		Attribute("aip_task_updated_event", AIPTaskUpdatedEvent)
	})
})

var StoragePingEvent = Type("StoragePingEvent", func() {
	Attribute("message", String)
})

var LocationCreatedEvent = Type("LocationCreatedEvent", func() {
	TypedAttributeUUID("uuid", "Identifier of Location")
	Attribute("item", Location)
	Required("uuid", "item")
})

var AIPCreatedEvent = Type("AIPCreatedEvent", func() {
	TypedAttributeUUID("uuid", "Identifier of AIP")
	Attribute("item", AIP)
	Required("uuid", "item")
})

var AIPStatusUpdatedEvent = Type("AIPStatusUpdatedEvent", func() {
	TypedAttributeUUID("uuid", "Identifier of AIP")
	Attribute("status", String, func() {
		EnumAIPStatus()
	})
	Required("uuid", "status")
})

var AIPLocationUpdatedEvent = Type("AIPLocationUpdatedEvent", func() {
	TypedAttributeUUID("uuid", "Identifier of AIP")
	TypedAttributeUUID("location_uuid", "Identifier of Location")
	Required("uuid", "location_uuid")
})

var AIPWorkflowCreatedEvent = Type("AIPWorkflowCreatedEvent", func() {
	TypedAttributeUUID("uuid", "Identifier of workflow")
	Attribute("item", AIPWorkflow)
	Required("uuid", "item")
})

var AIPWorkflowUpdatedEvent = Type("AIPWorkflowUpdatedEvent", func() {
	TypedAttributeUUID("uuid", "Identifier of workflow")
	Attribute("item", AIPWorkflow)
	Required("uuid", "item")
})

var AIPTaskCreatedEvent = Type("AIPTaskCreatedEvent", func() {
	TypedAttributeUUID("uuid", "Identifier of task")
	Attribute("item", AIPTask)
	Required("uuid", "item")
})

var AIPTaskUpdatedEvent = Type("AIPTaskUpdatedEvent", func() {
	TypedAttributeUUID("uuid", "Identifier of task")
	Attribute("item", AIPTask)
	Required("uuid", "item")
})
