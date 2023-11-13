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
			"package_created_event",
			PackageCreatedEvent,
		)
		Attribute(
			"package_updated_event",
			PackageUpdatedEvent,
		)
		Attribute(
			"package_status_updated_event",
			PackageStatusUpdatedEvent,
		)
		Attribute(
			"package_location_updated_event",
			PackageLocationUpdatedEvent,
		)
		Attribute(
			"preservation_action_created_event",
			PreservationActionCreatedEvent,
		)
		Attribute(
			"preservation_action_updated_event",
			PreservationActionUpdatedEvent,
		)
		Attribute(
			"preservation_task_created_event",
			PreservationTaskCreatedEvent,
		)
		Attribute(
			"preservation_task_updated_event",
			PreservationTaskUpdatedEvent,
		)
	})
})

var MonitorPingEvent = Type("MonitorPingEvent", func() {
	Attribute("message", String)

	Meta("type:generate:force")
	Meta("openapi:typename", "MonitorPingEvent")
})

var PackageCreatedEvent = Type("PackageCreatedEvent", func() {
	Attribute("id", UInt, "Identifier of package")
	Attribute("item", StoredPackage, func() { View("default") })
	Required("id", "item")

	Meta("type:generate:force")
	Meta("openapi:typename", "PackageCreatedEvent")
})

var PackageUpdatedEvent = Type("PackageUpdatedEvent", func() {
	Attribute("id", UInt, "Identifier of package")
	Attribute("item", StoredPackage, func() { View("default") })
	Required("id", "item")

	Meta("type:generate:force")
	Meta("openapi:typename", "PackageUpdatedEvent")
})

var PackageStatusUpdatedEvent = Type("PackageStatusUpdatedEvent", func() {
	Attribute("id", UInt, "Identifier of package")
	Attribute("status", String, func() {
		EnumPackageStatus()
	})
	Required("id", "status")

	Meta("type:generate:force")
	Meta("openapi:typename", "PackageStatusUpdatedEvent")
})

var PackageLocationUpdatedEvent = Type("PackageLocationUpdatedEvent", func() {
	Attribute("id", UInt, "Identifier of package")
	TypedAttributeUUID("location_id", "Identifier of storage location")
	Required("id", "location_id")

	Meta("type:generate:force")
	Meta("openapi:typename", "PackageLocationUpdatedEvent")
})

var PreservationActionCreatedEvent = Type("PreservationActionCreatedEvent", func() {
	Attribute("id", UInt, "Identifier of preservation action")
	Attribute("item", PreservationAction, func() {
		View("simple")
	})
	Required("id", "item")

	Meta("type:generate:force")
	Meta("openapi:typename", "PreservationActionCreatedEvent")
})

var PreservationActionUpdatedEvent = Type("PreservationActionUpdatedEvent", func() {
	Attribute("id", UInt, "Identifier of preservation action")
	Attribute("item", PreservationAction, func() {
		View("simple")
	})
	Required("id", "item")

	Meta("type:generate:force")
	Meta("openapi:typename", "PreservationActionUpdatedEvent")
})

var PreservationTaskCreatedEvent = Type("PreservationTaskCreatedEvent", func() {
	Attribute("id", UInt, "Identifier of preservation task")
	Attribute("item", PreservationTask, func() {
		View("default")
	})
	Required("id", "item")

	Meta("type:generate:force")
	Meta("openapi:typename", "PreservationTaskCreatedEvent")
})

var PreservationTaskUpdatedEvent = Type("PreservationTaskUpdatedEvent", func() {
	Attribute("id", UInt, "Identifier of preservation task")
	Attribute("item", PreservationTask, func() {
		View("default")
	})
	Required("id", "item")

	Meta("type:generate:force")
	Meta("openapi:typename", "PreservationTaskUpdatedEvent")
})
