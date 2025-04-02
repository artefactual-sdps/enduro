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
//   - Meta("openapi:typename", "MonitorPingEvent")
//     It guarantees that the schema is not omitted because there is another
//     type structurally equivalent, which is the default behavior in Goa.
//

var MonitorEvent = Type("MonitorEvent", func() {
	OneOf("event", func() {
		Attribute("monitor_ping_event", MonitorPingEvent)
		Attribute("sip_created_event", SIPCreatedEvent)
		Attribute("sip_updated_event", SIPUpdatedEvent)
		Attribute("sip_status_updated_event", SIPStatusUpdatedEvent)
		Attribute("sip_workflow_created_event", SIPWorkflowCreatedEvent)
		Attribute("sip_workflow_updated_event", SIPWorkflowUpdatedEvent)
		Attribute("sip_task_created_event", SIPTaskCreatedEvent)
		Attribute("sip_task_updated_event", SIPTaskUpdatedEvent)
	})
})

var MonitorPingEvent = Type("MonitorPingEvent", func() {
	Attribute("message", String)

	Meta("type:generate:force")
	Meta("openapi:typename", "MonitorPingEvent")
})

var SIPCreatedEvent = Type("SIPCreatedEvent", func() {
	Attribute("id", UInt, "Identifier of SIP")
	Attribute("item", SIP, func() { View("default") })
	Required("id", "item")

	Meta("type:generate:force")
	Meta("openapi:typename", "SIPCreatedEvent")
})

var SIPUpdatedEvent = Type("SIPUpdatedEvent", func() {
	Attribute("id", UInt, "Identifier of SIP")
	Attribute("item", SIP, func() { View("default") })
	Required("id", "item")

	Meta("type:generate:force")
	Meta("openapi:typename", "SIPUpdatedEvent")
})

var SIPStatusUpdatedEvent = Type("SIPStatusUpdatedEvent", func() {
	Attribute("id", UInt, "Identifier of SIP")
	Attribute("status", String, func() {
		EnumSIPStatus()
	})
	Required("id", "status")

	Meta("type:generate:force")
	Meta("openapi:typename", "SIPStatusUpdatedEvent")
})

var SIPWorkflowCreatedEvent = Type("SIPWorkflowCreatedEvent", func() {
	Attribute("id", UInt, "Identifier of workflow")
	Attribute("item", SIPWorkflow, func() {
		View("simple")
	})
	Required("id", "item")

	Meta("type:generate:force")
	Meta("openapi:typename", "SIPWorkflowCreatedEvent")
})

var SIPWorkflowUpdatedEvent = Type("SIPWorkflowUpdatedEvent", func() {
	Attribute("id", UInt, "Identifier of workflow")
	Attribute("item", SIPWorkflow, func() {
		View("simple")
	})
	Required("id", "item")

	Meta("type:generate:force")
	Meta("openapi:typename", "SIPWorkflowUpdatedEvent")
})

var SIPTaskCreatedEvent = Type("SIPTaskCreatedEvent", func() {
	Attribute("id", UInt, "Identifier of task")
	Attribute("item", SIPTask, func() {
		View("default")
	})
	Required("id", "item")

	Meta("type:generate:force")
	Meta("openapi:typename", "SIPTaskCreatedEvent")
})

var SIPTaskUpdatedEvent = Type("SIPTaskUpdatedEvent", func() {
	Attribute("id", UInt, "Identifier of task")
	Attribute("item", SIPTask, func() {
		View("default")
	})
	Required("id", "item")

	Meta("type:generate:force")
	Meta("openapi:typename", "SIPTaskUpdatedEvent")
})
