package design

import (
	. "goa.design/goa/v3/dsl"
)

var MonitorEvent = ResultType("application/vnd.enduro.monitor-event", func() {
	// TODO: use OneOf when possible.
	Attributes(func() {
		Attribute(
			"monitor_ping_event",
			MonitorPingEvent,
			func() { View("default") },
		)
		Attribute(
			"package_created_event",
			PackageCreatedEvent,
			func() { View("default") },
		)
		Attribute(
			"package_updated_event",
			PackageUpdatedEvent,
			func() { View("default") },
		)
		Attribute(
			"package_status_updated_event",
			PackageStatusUpdatedEvent,
			func() { View("default") },
		)
		Attribute(
			"package_location_updated_event",
			PackageLocationUpdatedEvent,
			func() { View("default") },
		)
		Attribute(
			"preservation_action_created_event",
			PreservationActionCreatedEvent,
			func() { View("default") },
		)
		Attribute(
			"preservation_action_updated_event",
			PreservationActionUpdatedEvent,
			func() { View("default") },
		)
		Attribute(
			"preservation_task_created_event",
			PreservationTaskCreatedEvent,
			func() { View("default") },
		)
		Attribute(
			"preservation_task_updated_event",
			PreservationTaskUpdatedEvent,
			func() { View("default") },
		)
	})
})

var MonitorPingEvent = ResultType("application/vnd.enduro.monitor-ping-event", func() {
	Attributes(func() {
		Attribute("message", String)
	})

	View("default", func() {
		Attribute("message")
	})
})

var PackageCreatedEvent = ResultType("application/vnd.enduro.package-created-event", func() {
	Attributes(func() {
		Attribute("id", UInt, "Identifier of package")
		Attribute("item", StoredPackage, func() { View("default") })
		Required("id", "item")
	})

	View("default", func() {
		Attribute("id")
		Attribute("item")
	})
})

var PackageUpdatedEvent = ResultType("application/vnd.enduro.package-updated-event", func() {
	Attributes(func() {
		Attribute("id", UInt, "Identifier of package")
		Attribute("item", StoredPackage, func() { View("default") })
		Required("id", "item")
	})

	View("default", func() {
		Attribute("id")
		Attribute("item")
	})
})

var PackageStatusUpdatedEvent = ResultType("application/vnd.enduro.package-status-updated-event", func() {
	Attributes(func() {
		Attribute("id", UInt, "Identifier of package")
		Attribute("status", String, func() {
			EnumPackageStatus()
		})
		Required("id", "status")
	})

	View("default", func() {
		Attribute("id")
		Attribute("status")
	})
})

var PackageLocationUpdatedEvent = ResultType("application/vnd.enduro.package-location-updated-event", func() {
	Attributes(func() {
		Attribute("id", UInt, "Identifier of package")
		TypedAttributeUUID("location_id", "Identifier of storage location")
		Required("id", "location_id")
	})

	View("default", func() {
		Attribute("id")
		Attribute("location_id")
	})
})

var PreservationActionCreatedEvent = ResultType("application/vnd.enduro.preservation-action-created-event", func() {
	Attributes(func() {
		Attribute("id", UInt, "Identifier of preservation action")
		Attribute("item", PreservationAction, func() {
			View("simple")
		})
		Required("id", "item")
	})

	View("default", func() {
		Attribute("id")
		Attribute("item")
	})
})

var PreservationActionUpdatedEvent = ResultType("application/vnd.enduro.preservation-action-updated-event", func() {
	Attributes(func() {
		Attribute("id", UInt, "Identifier of preservation action")
		Attribute("item", PreservationAction, func() {
			View("simple")
		})
		Required("id", "item")
	})

	View("default", func() {
		Attribute("id")
		Attribute("item")
	})
})

var PreservationTaskCreatedEvent = ResultType("application/vnd.enduro.preservation-task-created-event", func() {
	Attributes(func() {
		Attribute("id", UInt, "Identifier of preservation task")
		Attribute("item", PreservationTask, func() {
			View("default")
		})
		Required("id", "item")
	})

	View("default", func() {
		Attribute("id")
		Attribute("item")
	})
})

var PreservationTaskUpdatedEvent = ResultType("application/vnd.enduro.preservation-task-updated-event", func() {
	Attributes(func() {
		Attribute("id", UInt, "Identifier of preservation task")
		Attribute("item", PreservationTask, func() {
			View("default")
		})
		Required("id", "item")
	})

	View("default", func() {
		Attribute("id")
		Attribute("item")
	})
})
