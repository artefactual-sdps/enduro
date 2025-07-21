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
//   - Meta("openapi:typename", "StorageMonitorPingEvent")
//     It guarantees that the schema is not omitted because there is another
//     type structurally equivalent, which is the default behavior in Goa.
//

var StorageMonitorEvent = Type("StorageMonitorEvent", func() {
	OneOf("event", func() {
		Attribute("monitor_ping_event", StorageMonitorPingEvent)
		Attribute("location_created_event", LocationCreatedEvent)
		Attribute("location_updated_event", LocationUpdatedEvent)
		Attribute("aip_created_event", AIPCreatedEvent)
		Attribute("aip_updated_event", AIPUpdatedEvent)
		Attribute("workflow_created_event", WorkflowCreatedEvent)
		Attribute("workflow_updated_event", WorkflowUpdatedEvent)
		Attribute("task_created_event", TaskCreatedEvent)
		Attribute("task_updated_event", TaskUpdatedEvent)
	})
})

var StorageMonitorPingEvent = Type("StorageMonitorPingEvent", func() {
	Attribute("message", String)

	Meta("type:generate:force")
	Meta("openapi:typename", "StorageMonitorPingEvent")
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

var WorkflowCreatedEvent = Type("WorkflowCreatedEvent", func() {
	TypedAttributeUUID("uuid", "Identifier of workflow")
	Attribute("item", AIPWorkflow)
	Required("uuid", "item")

	Meta("type:generate:force")
	Meta("openapi:typename", "WorkflowCreatedEvent")
})

var WorkflowUpdatedEvent = Type("WorkflowUpdatedEvent", func() {
	TypedAttributeUUID("uuid", "Identifier of workflow")
	Attribute("item", AIPWorkflow)
	Required("uuid", "item")

	Meta("type:generate:force")
	Meta("openapi:typename", "WorkflowUpdatedEvent")
})

var TaskCreatedEvent = Type("TaskCreatedEvent", func() {
	TypedAttributeUUID("uuid", "Identifier of task")
	Attribute("item", AIPTask)
	Required("uuid", "item")

	Meta("type:generate:force")
	Meta("openapi:typename", "TaskCreatedEvent")
})

var TaskUpdatedEvent = Type("TaskUpdatedEvent", func() {
	TypedAttributeUUID("uuid", "Identifier of task")
	Attribute("item", AIPTask)
	Required("uuid", "item")

	Meta("type:generate:force")
	Meta("openapi:typename", "TaskUpdatedEvent")
})