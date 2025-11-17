package design

import (
	. "goa.design/goa/v3/dsl" //nolint:staticcheck
)

// This type is used as a StreamingResult, Goa generates two OpenAPI schemas,
// one for the declared type itself and one for its streaming representation.
// To avoid a name clash the latter is suffixed as IngestEvent2.
var IngestEvent = Type("IngestEvent", func() {
	OneOf("value", func() {
		Attribute("ingest_ping_event", IngestPingEvent)
		Attribute("sip_created_event", SIPCreatedEvent)
		Attribute("sip_updated_event", SIPUpdatedEvent)
		Attribute("sip_status_updated_event", SIPStatusUpdatedEvent)
		Attribute("sip_workflow_created_event", SIPWorkflowCreatedEvent)
		Attribute("sip_workflow_updated_event", SIPWorkflowUpdatedEvent)
		Attribute("sip_task_created_event", SIPTaskCreatedEvent)
		Attribute("sip_task_updated_event", SIPTaskUpdatedEvent)
		Attribute("batch_created_event", BatchCreatedEvent)
		Attribute("batch_updated_event", BatchUpdatedEvent)
	})
})

var IngestPingEvent = Type("IngestPingEvent", func() {
	Attribute("message", String)
})

var SIPCreatedEvent = Type("SIPCreatedEvent", func() {
	TypedAttributeUUID("uuid", "Identifier of SIP")
	Attribute("item", SIP, func() { View("default") })
	Required("uuid", "item")
})

var SIPUpdatedEvent = Type("SIPUpdatedEvent", func() {
	TypedAttributeUUID("uuid", "Identifier of SIP")
	Attribute("item", SIP, func() { View("default") })
	Required("uuid", "item")
})

var SIPStatusUpdatedEvent = Type("SIPStatusUpdatedEvent", func() {
	TypedAttributeUUID("uuid", "Identifier of SIP")
	Attribute("status", String, func() {
		EnumSIPStatus()
	})
	Required("uuid", "status")
})

var SIPWorkflowCreatedEvent = Type("SIPWorkflowCreatedEvent", func() {
	TypedAttributeUUID("uuid", "Identifier of workflow")
	Attribute("item", SIPWorkflow, func() {
		View("simple")
	})
	Required("uuid", "item")
})

var SIPWorkflowUpdatedEvent = Type("SIPWorkflowUpdatedEvent", func() {
	TypedAttributeUUID("uuid", "Identifier of workflow")
	Attribute("item", SIPWorkflow, func() {
		View("simple")
	})
	Required("uuid", "item")
})

var SIPTaskCreatedEvent = Type("SIPTaskCreatedEvent", func() {
	TypedAttributeUUID("uuid", "Identifier of task")
	Attribute("item", SIPTask, func() {
		View("default")
	})
	Required("uuid", "item")
})

var SIPTaskUpdatedEvent = Type("SIPTaskUpdatedEvent", func() {
	TypedAttributeUUID("uuid", "Identifier of task")
	Attribute("item", SIPTask, func() {
		View("default")
	})
	Required("uuid", "item")
})

var BatchCreatedEvent = Type("BatchCreatedEvent", func() {
	TypedAttributeUUID("uuid", "Identifier of Batch")
	Attribute("item", Batch)
	Required("uuid", "item")
})

var BatchUpdatedEvent = Type("BatchUpdatedEvent", func() {
	TypedAttributeUUID("uuid", "Identifier of Batch")
	Attribute("item", Batch)
	Required("uuid", "item")
})
