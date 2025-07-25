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
//   - Meta("openapi:typename", "IngestPingEvent")
//     It guarantees that the schema is not omitted because there is another
//     type structurally equivalent, which is the default behavior in Goa.
//

var IngestEvent = Type("IngestEvent", func() {
	// The IngestEvent and the StorageEvent share a similar structure.
	// To differentiate them we use distinct keys for the event value.
	// If we use the Meta("type:generate:force") option in this type we
	// get two versions in the design because it's used as an streaming
	// response, and we get HTTP and WebSocket schemas. Without the meta
	// and the same structure we only get the IngestEvent in the schema.
	OneOf("ingest_value", func() {
		Attribute("ingest_ping_event", IngestPingEvent)
		Attribute("sip_created_event", SIPCreatedEvent)
		Attribute("sip_updated_event", SIPUpdatedEvent)
		Attribute("sip_status_updated_event", SIPStatusUpdatedEvent)
		Attribute("sip_workflow_created_event", SIPWorkflowCreatedEvent)
		Attribute("sip_workflow_updated_event", SIPWorkflowUpdatedEvent)
		Attribute("sip_task_created_event", SIPTaskCreatedEvent)
		Attribute("sip_task_updated_event", SIPTaskUpdatedEvent)
	})
})

var IngestPingEvent = Type("IngestPingEvent", func() {
	Attribute("message", String)

	Meta("type:generate:force")
	Meta("openapi:typename", "IngestPingEvent")
})

var SIPCreatedEvent = Type("SIPCreatedEvent", func() {
	TypedAttributeUUID("uuid", "Identifier of SIP")
	Attribute("item", SIP, func() { View("default") })
	Required("uuid", "item")

	Meta("type:generate:force")
	Meta("openapi:typename", "SIPCreatedEvent")
})

var SIPUpdatedEvent = Type("SIPUpdatedEvent", func() {
	TypedAttributeUUID("uuid", "Identifier of SIP")
	Attribute("item", SIP, func() { View("default") })
	Required("uuid", "item")

	Meta("type:generate:force")
	Meta("openapi:typename", "SIPUpdatedEvent")
})

var SIPStatusUpdatedEvent = Type("SIPStatusUpdatedEvent", func() {
	TypedAttributeUUID("uuid", "Identifier of SIP")
	Attribute("status", String, func() {
		EnumSIPStatus()
	})
	Required("uuid", "status")

	Meta("type:generate:force")
	Meta("openapi:typename", "SIPStatusUpdatedEvent")
})

var SIPWorkflowCreatedEvent = Type("SIPWorkflowCreatedEvent", func() {
	TypedAttributeUUID("uuid", "Identifier of workflow")
	Attribute("item", SIPWorkflow, func() {
		View("simple")
	})
	Required("uuid", "item")

	Meta("type:generate:force")
	Meta("openapi:typename", "SIPWorkflowCreatedEvent")
})

var SIPWorkflowUpdatedEvent = Type("SIPWorkflowUpdatedEvent", func() {
	TypedAttributeUUID("uuid", "Identifier of workflow")
	Attribute("item", SIPWorkflow, func() {
		View("simple")
	})
	Required("uuid", "item")

	Meta("type:generate:force")
	Meta("openapi:typename", "SIPWorkflowUpdatedEvent")
})

var SIPTaskCreatedEvent = Type("SIPTaskCreatedEvent", func() {
	TypedAttributeUUID("uuid", "Identifier of task")
	Attribute("item", SIPTask, func() {
		View("default")
	})
	Required("uuid", "item")

	Meta("type:generate:force")
	Meta("openapi:typename", "SIPTaskCreatedEvent")
})

var SIPTaskUpdatedEvent = Type("SIPTaskUpdatedEvent", func() {
	TypedAttributeUUID("uuid", "Identifier of task")
	Attribute("item", SIPTask, func() {
		View("default")
	})
	Required("uuid", "item")

	Meta("type:generate:force")
	Meta("openapi:typename", "SIPTaskUpdatedEvent")
})
