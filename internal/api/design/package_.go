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
			Attribute("location", String)
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
				Param("location")
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
		HTTP(func() {
			GET("/{id}")
			Response(StatusOK)
			Response("not_found", StatusNotFound)
		})
	})
	Method("delete", func() {
		Description("Delete package by ID")
		Payload(func() {
			Attribute("id", UInt, "Identifier of package to delete")
			Required("id")
		})
		Error("not_found", PackageNotFound, "Package not found")
		HTTP(func() {
			DELETE("/{id}")
			Response(StatusNoContent)
			Response("not_found", StatusNotFound)
		})
	})
	Method("cancel", func() {
		Description("Cancel package processing by ID")
		Payload(func() {
			Attribute("id", UInt, "Identifier of package to remove")
			Required("id")
		})
		Error("not_found", PackageNotFound, "Package not found")
		Error("not_running")
		HTTP(func() {
			POST("/{id}/cancel")
			Response(StatusOK)
			Response("not_found", StatusNotFound)
			Response("not_running", StatusBadRequest)
		})
	})
	Method("retry", func() {
		Description("Retry package processing by ID")
		Payload(func() {
			Attribute("id", UInt, "Identifier of package to retry")
			Required("id")
		})
		Error("not_found", PackageNotFound, "Package not found")
		Error("not_running")
		HTTP(func() {
			POST("/{id}/retry")
			Response(StatusOK)
			Response("not_found", StatusNotFound)
			Response("not_running", StatusBadRequest)
		})
	})
	Method("workflow", func() {
		Description("Retrieve workflow status by ID")
		Payload(func() {
			Attribute("id", UInt, "Identifier of package to look up")
			Required("id")
		})
		Result(WorkflowStatus)
		Error("not_found", PackageNotFound, "Package not found")
		HTTP(func() {
			GET("/{id}/workflow")
			Response(StatusOK)
			Response("not_found", StatusNotFound)
		})
	})
	Method("bulk", func() {
		Description("Bulk operations (retry, cancel...).")
		Payload(func() {
			Attribute("operation", String, func() {
				Enum("retry", "cancel", "abandon")
			})
			Attribute("status", String, func() {
				EnumPackageStatus()
			})
			Attribute("size", UInt, func() {
				Default(100)
			})
			Required("operation", "status")
		})
		Result(BulkResult)
		Error("not_available")
		Error("not_valid")
		HTTP(func() {
			POST("/bulk")
			Response(StatusAccepted)
			Response("not_available", StatusConflict)
			Response("not_valid", StatusBadRequest)
		})
	})
	Method("bulk_status", func() {
		Description("Retrieve status of current bulk operation.")
		Result(BulkStatusResult)
		HTTP(func() {
			GET("/bulk")
			Response(StatusOK)
		})
	})
	Method("preservation-actions", func() {
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
			Attribute("location", String)
			Required("id", "location")
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
			Attribute("location", String)
			Required("id", "location")
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
	Attribute("location", String, "Location of the package")
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
		Attribute("location")
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
		Attribute("location")
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

var WorkflowStatus = ResultType("application/vnd.enduro.package-workflow-status", func() {
	Description("WorkflowStatus describes the processing workflow status of a package.")
	Attributes(func() {
		Attribute("status", String) // TODO
		Attribute("history", CollectionOf(WorkflowHistoryEvent))
	})
})

var WorkflowHistoryEvent = ResultType("application/vnd.enduro.package-workflow-history", func() {
	Description("WorkflowHistoryEvent describes a history event in Temporal.")
	Attributes(func() {
		Attribute("id", UInt, "Identifier of package")
		Attribute("type", String, "Type of the event")
		Attribute("details", Any, "Contents of the event")
	})
})

var PackageNotFound = Type("PackageNotfound", func() {
	Description("Package not found.")
	Attribute("message", String, "Message of error", func() {
		Meta("struct:error:name")
	})
	Attribute("id", UInt, "Identifier of missing package")
	Required("message", "id")
})

var BulkResult = Type("BulkResult", func() {
	Attribute("workflow_id", String)
	Attribute("run_id", String)
	Required("workflow_id", "run_id")
})

var BulkStatusResult = Type("BulkStatusResult", func() {
	Attribute("running", Boolean)
	Attribute("started_at", String, func() {
		Format(FormatDateTime)
	})
	Attribute("closed_at", String, func() {
		Format(FormatDateTime)
	})
	Attribute("status", String)
	Attribute("workflow_id", String)
	Attribute("run_id", String)
	Required("running")
})

var PreservationActions = ResultType("application/vnd.enduro.package-preservation-actions", func() {
	Description("PreservationActions describes the preservation actions of a package.")
	Attributes(func() {
		Attribute("actions", CollectionOf(PreservationAction))
	})
})

var EnumPreservationActionType = func() {
	Enum("create-aip", "move-package")
}

var EnumPreservationActionStatus = func() {
	Enum("unspecified", "complete", "processing", "failed")
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
	})
	Required("id", "workflow_id", "type", "status", "started_at")
})

var EnumPreservationTaskStatus = func() {
	Enum("unspecified", "complete", "processing", "failed")
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
	})
	Required("id", "task_id", "name", "status", "started_at")
})
