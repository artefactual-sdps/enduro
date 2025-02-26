package design

import (
	. "goa.design/goa/v3/dsl"
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
		Attribute(
			"monitor_ping_event",
			MonitorPingEvent,
		)
		Attribute(
			"sip_created_event",
			SIPCreatedEvent,
		)
		Attribute(
			"sip_updated_event",
			SIPUpdatedEvent,
		)
		Attribute(
			"sip_status_updated_event",
			SIPStatusUpdatedEvent,
		)
		Attribute(
			"sip_location_updated_event",
			SIPLocationUpdatedEvent,
		)
		Attribute(
			"sip_preservation_action_created_event",
			SIPPreservationActionCreatedEvent,
		)
		Attribute(
			"sip_preservation_action_updated_event",
			SIPPreservationActionUpdatedEvent,
		)
		Attribute(
			"sip_preservation_task_created_event",
			SIPPreservationTaskCreatedEvent,
		)
		Attribute(
			"sip_preservation_task_updated_event",
			SIPPreservationTaskUpdatedEvent,
		)
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

var SIPLocationUpdatedEvent = Type("SIPLocationUpdatedEvent", func() {
	Attribute("id", UInt, "Identifier of SIP")
	TypedAttributeUUID("location_id", "Identifier of storage location")
	Required("id", "location_id")

	Meta("type:generate:force")
	Meta("openapi:typename", "SIPLocationUpdatedEvent")
})

var SIPPreservationActionCreatedEvent = Type("SIPPreservationActionCreatedEvent", func() {
	Attribute("id", UInt, "Identifier of preservation action")
	Attribute("item", SIPPreservationAction, func() {
		View("simple")
	})
	Required("id", "item")

	Meta("type:generate:force")
	Meta("openapi:typename", "SIPPreservationActionCreatedEvent")
})

var SIPPreservationActionUpdatedEvent = Type("SIPPreservationActionUpdatedEvent", func() {
	Attribute("id", UInt, "Identifier of preservation action")
	Attribute("item", SIPPreservationAction, func() {
		View("simple")
	})
	Required("id", "item")

	Meta("type:generate:force")
	Meta("openapi:typename", "SIPPreservationActionUpdatedEvent")
})

var SIPPreservationTaskCreatedEvent = Type("SIPPreservationTaskCreatedEvent", func() {
	Attribute("id", UInt, "Identifier of preservation task")
	Attribute("item", SIPPreservationTask, func() {
		View("default")
	})
	Required("id", "item")

	Meta("type:generate:force")
	Meta("openapi:typename", "SIPPreservationTaskCreatedEvent")
})

var SIPPreservationTaskUpdatedEvent = Type("SIPPreservationTaskUpdatedEvent", func() {
	Attribute("id", UInt, "Identifier of preservation task")
	Attribute("item", SIPPreservationTask, func() {
		View("default")
	})
	Required("id", "item")

	Meta("type:generate:force")
	Meta("openapi:typename", "SIPPreservationTaskUpdatedEvent")
})
