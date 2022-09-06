package design

import (
	. "goa.design/goa/v3/dsl"
)

var _ = Service("package", func() {
	Description("The package service manages packages being transferred to a3m.")
	HTTP(func() {
		Path("/package")
	})
	Method("monitor", func() {
		StreamingResult(MonitorEvent, func() {
			View("default")
		})
		HTTP(func() {
			GET("/monitor")
			Response(StatusOK)
		})
	})
	Method("list", func() {
		Description("List all stored packages")
		Payload(func() {
			Attribute("name", String)
			Attribute("aip_id", String, func() {
				Format(FormatUUID)
			})
			Attribute("earliest_created_time", String, func() {
				Format(FormatDateTime)
			})
			Attribute("latest_created_time", String, func() {
				Format(FormatDateTime)
			})
			Attribute("location_id", String, func() {
				Format(FormatUUID)
			})
			Attribute("status", String, func() {
				EnumPackageStatus()
			})
			Attribute("cursor", String, "Pagination cursor")
		})
		Result(PaginatedCollectionOf(StoredPackage))
		HTTP(func() {
			GET("/")
			Response(StatusOK)
			Params(func() {
				Param("name")
				Param("aip_id")
				Param("earliest_created_time")
				Param("latest_created_time")
				Param("location_id")
				Param("status")
				Param("cursor")
			})
		})
	})
	Method("show", func() {
		Description("Show package by ID")
		Payload(func() {
			Attribute("id", UInt, "Identifier of package to show")
			Required("id")
		})
		Result(StoredPackage)
		Error("not_found", PackageNotFound, "Package not found")
		Error("not_available")
		HTTP(func() {
			GET("/{id}")
			Response(StatusOK)
			Response("not_found", StatusNotFound)
			Response("not_available", StatusConflict)
		})
	})
	Method("preservation_actions", func() {
		Description("List all preservation actions by ID")
		Payload(func() {
			Attribute("id", UInt, "Identifier of package to look up")
			Required("id")
		})
		Result(PreservationActions)
		Error("not_found", PackageNotFound, "Package not found")
		HTTP(func() {
			GET("/{id}/preservation-actions")
			Response(StatusOK)
			Response("not_found", StatusNotFound)
		})
	})
	Method("confirm", func() {
		Description("Signal the package has been reviewed and accepted")
		Payload(func() {
			Attribute("id", UInt, "Identifier of package to look up")
			Attribute("location_id", String, func() {
				Meta("struct:field:type", "uuid.UUID", "github.com/google/uuid")
			})
			Required("id", "location_id")
		})
		Error("not_found", PackageNotFound, "Package not found")
		Error("not_available")
		Error("not_valid")
		HTTP(func() {
			POST("/{id}/confirm")
			Response(StatusAccepted)
			Response("not_found", StatusNotFound)
			Response("not_available", StatusConflict)
			Response("not_valid", StatusBadRequest)
		})
	})
	Method("reject", func() {
		Description("Signal the package has been reviewed and rejected")
		Payload(func() {
			Attribute("id", UInt, "Identifier of package to look up")
			Required("id")
		})
		Error("not_found", PackageNotFound, "Package not found")
		Error("not_available")
		Error("not_valid")
		HTTP(func() {
			POST("/{id}/reject")
			Response(StatusAccepted)
			Response("not_found", StatusNotFound)
			Response("not_available", StatusConflict)
			Response("not_valid", StatusBadRequest)
		})
	})
	Method("move", func() {
		Description("Move a package to a permanent storage location")
		Payload(func() {
			Attribute("id", UInt, "Identifier of package to move")
			Attribute("location_id", String, func() {
				Meta("struct:field:type", "uuid.UUID", "github.com/google/uuid")
			})
			Required("id", "location_id")
		})
		Error("not_found", PackageNotFound, "Package not found")
		Error("not_available")
		Error("not_valid")
		HTTP(func() {
			POST("/{id}/move")
			Response(StatusAccepted)
			Response("not_found", StatusNotFound)
			Response("not_available", StatusConflict)
			Response("not_valid", StatusBadRequest)
		})
	})
	Method("move_status", func() {
		Description("Retrieve the status of a permanent storage location move of the package")
		Payload(func() {
			Attribute("id", UInt, "Identifier of package to move")
			Required("id")
		})
		Result(MoveStatusResult)
		Error("not_found", PackageNotFound, "Package not found")
		Error("failed_dependency")
		HTTP(func() {
			GET("/{id}/move")
			Response(StatusOK)
			Response("not_found", StatusNotFound)
			Response("failed_dependency", StatusFailedDependency)
		})
	})
})

var EnumPackageStatus = func() {
	Enum("new", "in progress", "done", "error", "unknown", "queued", "pending", "abandoned")
}

var Package_ = Type("Package", func() {
	Description("Package describes a package to be stored.")
	Attribute("name", String, "Name of the package")
	Attribute("location_id", String, func() {
		Meta("struct:field:type", "uuid.UUID", "github.com/google/uuid")
	})
	Attribute("status", String, "Status of the package", func() {
		EnumPackageStatus()
		Default("new")
	})
	Attribute("workflow_id", String, "Identifier of processing workflow", func() {
		Format(FormatUUID)
	})
	Attribute("run_id", String, "Identifier of latest processing workflow run", func() {
		Format(FormatUUID)
	})
	Attribute("aip_id", String, "Identifier of Archivematica AIP", func() {
		Format(FormatUUID)
	})
	Attribute("created_at", String, "Creation datetime", func() {
		Format(FormatDateTime)
	})
	Attribute("started_at", String, "Start datetime", func() {
		Format(FormatDateTime)
	})
	Attribute("completed_at", String, "Completion datetime", func() {
		Format(FormatDateTime)
	})
	Required("id", "status", "created_at")
})

var StoredPackage = ResultType("application/vnd.enduro.stored-package", func() {
	Description("StoredPackage describes a package retrieved by the service.")
	Reference(Package_)
	Attributes(func() {
		Attribute("id", UInt, "Identifier of package")
		Attribute("name")
		Attribute("location_id")
		Attribute("status")
		Attribute("workflow_id")
		Attribute("run_id")
		Attribute("aip_id")
		Attribute("created_at")
		Attribute("started_at")
		Attribute("completed_at")
	})
	View("default", func() {
		Attribute("id")
		Attribute("name")
		Attribute("location_id")
		Attribute("status")
		Attribute("workflow_id")
		Attribute("run_id")
		Attribute("aip_id")
		Attribute("created_at")
		Attribute("started_at")
		Attribute("completed_at")
	})
	Required("id", "status", "created_at")
})

var PackageNotFound = Type("PackageNotFound", func() {
	Description("Package not found.")
	TypeName("PackageNotFound")
	Attribute("message", String, "Message of error", func() {
		Meta("struct:error:message")
	})
	Attribute("id", UInt, "Identifier of missing package")
	Required("message", "id")
})

var PreservationActions = ResultType("application/vnd.enduro.package-preservation-actions", func() {
	Description("PreservationActions describes the preservation actions of a package.")
	Attributes(func() {
		Attribute("actions", CollectionOf(PreservationAction))
	})
})

var EnumPreservationActionType = func() {
	Enum("create-aip", "create-and-review-aip", "move-package")
}

var EnumPreservationActionStatus = func() {
	Enum("unspecified", "in progress", "done", "error", "queued", "pending")
}

var PreservationAction = ResultType("application/vnd.enduro.package-preservation-action", func() {
	Description("PreservationAction describes a preservation action.")
	Attributes(func() {
		Attribute("id", UInt)
		Attribute("workflow_id", String)
		Attribute("type", String, func() {
			EnumPreservationActionType()
		})
		Attribute("status", String, func() {
			EnumPreservationActionStatus()
		})
		Attribute("started_at", String, func() {
			Format(FormatDateTime)
		})
		Attribute("completed_at", String, func() {
			Format(FormatDateTime)
		})
		Attribute("tasks", CollectionOf(PreservationTask))
		Attribute("package_id", UInt)
	})
	View("simple", func() {
		Attribute("id")
		Attribute("workflow_id")
		Attribute("type")
		Attribute("status")
		Attribute("started_at")
		Attribute("completed_at")
		Attribute("package_id")
	})
	Required("id", "workflow_id", "type", "status", "started_at")
})

var EnumPreservationTaskStatus = func() {
	Enum("unspecified", "in progress", "done", "error", "queued", "pending")
}

var PreservationTask = ResultType("application/vnd.enduro.package-preservation-task", func() {
	Description("PreservationTask describes a preservation action task.")
	Attributes(func() {
		Attribute("id", UInt)
		Attribute("task_id", String)
		Attribute("name", String)
		Attribute("status", String, func() {
			EnumPreservationTaskStatus()
		})
		Attribute("started_at", String, func() {
			Format(FormatDateTime)
		})
		Attribute("completed_at", String, func() {
			Format(FormatDateTime)
		})
		Attribute("note", String)
		Attribute("preservation_action_id", UInt)
	})
	Required("id", "task_id", "name", "status", "started_at")
})
