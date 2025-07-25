package design

import (
	. "goa.design/goa/v3/dsl" //nolint:staticcheck
)

//
// We use a couple of Meta attributes in this file that are important for the
// generator to produce the expected results:
//
//   - Meta("type:generate:force")
//     It guarantees that the schema is included in the OpenAPI spec when it
//     is only listed as a member of an union type (OneOf).
//
//   - Meta("openapi:typename", "StoragePingEvent")
//     It guarantees that the schema is not omitted because there is another
//     type structurally equivalent, which is the default behavior in Goa.
//

var StorageEvent = Type("StorageEvent", func() {
	// The IngestEvent and the StorageEvent share a similar structure.
	// To differentiate them we use distinct keys for the event value.
	// If we use the Meta("type:generate:force") option in this type we
	// get two versions in the design because it's used as an streaming
	// response, and we get HTTP and WebSocket schemas. Without the meta
	// and the same structure we only get the IngestEvent in the schema.
	OneOf("storage_value", func() {
		Attribute("storage_ping_event", StoragePingEvent)
		Attribute("location_created_event", LocationCreatedEvent)
		Attribute("location_updated_event", LocationUpdatedEvent)
		Attribute("aip_created_event", AIPCreatedEvent)
		Attribute("aip_updated_event", AIPUpdatedEvent)
		Attribute("aip_workflow_created_event", AIPWorkflowCreatedEvent)
		Attribute("aip_workflow_updated_event", AIPWorkflowUpdatedEvent)
		Attribute("aip_task_created_event", AIPTaskCreatedEvent)
		Attribute("aip_task_updated_event", AIPTaskUpdatedEvent)
	})
})

var StoragePingEvent = Type("StoragePingEvent", func() {
	Attribute("message", String)

	Meta("type:generate:force")
	Meta("openapi:typename", "StoragePingEvent")
})

var LocationCreatedEvent = Type("LocationCreatedEvent", func() {
	TypedAttributeUUID("uuid", "Identifier of Location")
	Attribute("item", Location)
	Required("uuid", "item")

	Meta("type:generate:force")
	Meta("openapi:typename", "LocationCreatedEvent")
})

var LocationUpdatedEvent = Type("LocationUpdatedEvent", func() {
	TypedAttributeUUID("uuid", "Identifier of Location")
	Attribute("item", Location)
	Required("uuid", "item")

	Meta("type:generate:force")
	Meta("openapi:typename", "LocationUpdatedEvent")
})

var AIPCreatedEvent = Type("AIPCreatedEvent", func() {
	TypedAttributeUUID("uuid", "Identifier of AIP")
	Attribute("item", AIP)
	Required("uuid", "item")

	Meta("type:generate:force")
	Meta("openapi:typename", "AIPCreatedEvent")
})

var AIPUpdatedEvent = Type("AIPUpdatedEvent", func() {
	TypedAttributeUUID("uuid", "Identifier of AIP")
	Attribute("item", AIP)
	Required("uuid", "item")

	Meta("type:generate:force")
	Meta("openapi:typename", "AIPUpdatedEvent")
})

var AIPWorkflowCreatedEvent = Type("AIPWorkflowCreatedEvent", func() {
	TypedAttributeUUID("uuid", "Identifier of workflow")
	Attribute("item", AIPWorkflow)
	Required("uuid", "item")

	Meta("type:generate:force")
	Meta("openapi:typename", "AIPWorkflowCreatedEvent")
})

var AIPWorkflowUpdatedEvent = Type("AIPWorkflowUpdatedEvent", func() {
	TypedAttributeUUID("uuid", "Identifier of workflow")
	Attribute("item", AIPWorkflow)
	Required("uuid", "item")

	Meta("type:generate:force")
	Meta("openapi:typename", "AIPWorkflowUpdatedEvent")
})

var AIPTaskCreatedEvent = Type("AIPTaskCreatedEvent", func() {
	TypedAttributeUUID("uuid", "Identifier of task")
	Attribute("item", AIPTask)
	Required("uuid", "item")

	Meta("type:generate:force")
	Meta("openapi:typename", "AIPTaskCreatedEvent")
})

var AIPTaskUpdatedEvent = Type("AIPTaskUpdatedEvent", func() {
	TypedAttributeUUID("uuid", "Identifier of task")
	Attribute("item", AIPTask)
	Required("uuid", "item")

	Meta("type:generate:force")
	Meta("openapi:typename", "AIPTaskUpdatedEvent")
})
