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
			"package_deleted_event",
			PackageDeletedEvent,
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

var PackageDeletedEvent = ResultType("application/vnd.enduro.package-deleted-event", func() {
	Attributes(func() {
		Attribute("id", UInt, "Identifier of package")
		Required("id")
	})

	View("default", func() {
		Attribute("id")
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
		Attribute("location", String)
		Required("id", "location")
	})

	View("default", func() {
		Attribute("id")
		Attribute("location")
	})
})
