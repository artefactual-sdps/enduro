package design

import (
	. "goa.design/goa/v3/dsl"

	"github.com/artefactual-sdps/enduro/internal/enums"
)

var _ = Service("package", func() {
	Description("The package service manages packages being transferred to a3m.")
	Error("unauthorized", String, "Unauthorized")
	Error("forbidden", String, "Forbidden")
	HTTP(func() {
		Path("/package")
		Response("unauthorized", StatusUnauthorized)
		Response("forbidden", StatusForbidden)
	})
	Method("monitor_request", func() {
		Description("Request access to the /monitor WebSocket")
		// For now, the monitor websocket requires all the scopes from this service.
		Security(JWTAuth, func() {
			Scope("package:list")
			Scope("package:listActions")
			Scope("package:move")
			Scope("package:read")
			Scope("package:review")
			Scope("package:upload")
		})
		Payload(func() {
			Token("token", String)
		})
		Result(func() {
			Attribute("ticket", String)
		})
		Error("not_available")
		HTTP(func() {
			POST("/monitor")
			Response("not_available", StatusInternalServerError)
			Response(StatusOK, func() {
				Cookie("ticket:enduro-ws-ticket")
				CookieMaxAge(5)
				CookieSecure()
				CookieHTTPOnly()
			})
		})
	})
	Method("monitor", func() {
		Description("Obtain access to the /monitor WebSocket")
		// Disable JWTAuth security (it validates the previous method cookie).
		NoSecurity()
		Payload(func() {
			Attribute("ticket", String)
		})
		StreamingResult(MonitorEvent)
		Error("not_available")
		HTTP(func() {
			GET("/monitor")
			Response("not_available", StatusInternalServerError)
			Response(StatusOK)
			Cookie("ticket:enduro-ws-ticket")
		})
	})
	Method("list", func() {
		Description("List all stored packages")
		Security(JWTAuth, func() {
			Scope("package:list")
		})
		Payload(func() {
			Attribute("name", String)
			AttributeUUID("aip_id", "Identifier of AIP")
			Attribute("earliest_created_time", String, func() {
				Format(FormatDateTime)
			})
			Attribute("latest_created_time", String, func() {
				Format(FormatDateTime)
			})
			AttributeUUID("location_id", "Identifier of storage location")
			Attribute("status", String, func() {
				EnumPackageStatus()
			})
			Attribute("limit", Int, "Limit number of results to return")
			Attribute("offset", Int, "Offset from the beginning of the found set")

			Token("token", String)
		})
		Result(PackageList)
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
				Param("limit")
				Param("offset")
			})
		})
	})
	Method("show", func() {
		Description("Show package by ID")
		Security(JWTAuth, func() {
			Scope("package:read")
		})
		Payload(func() {
			Attribute("id", UInt, "Identifier of package to show")
			Token("token", String)
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
		Security(JWTAuth, func() {
			Scope("package:listActions")
		})
		Payload(func() {
			Attribute("id", UInt, "Identifier of package to look up")
			Token("token", String)
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
		Security(JWTAuth, func() {
			Scope("package:review")
		})
		Payload(func() {
			Attribute("id", UInt, "Identifier of package to look up")
			TypedAttributeUUID("location_id", "Identifier of storage location")
			Token("token", String)
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
		Security(JWTAuth, func() {
			Scope("package:review")
		})
		Payload(func() {
			Attribute("id", UInt, "Identifier of package to look up")
			Token("token", String)
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
		Security(JWTAuth, func() {
			Scope("package:move")
		})
		Payload(func() {
			Attribute("id", UInt, "Identifier of package to move")
			TypedAttributeUUID("location_id", "Identifier of storage location")
			Token("token", String)
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
		Security(JWTAuth, func() {
			Scope("package:move")
		})
		Payload(func() {
			Attribute("id", UInt, "Identifier of package to move")
			Token("token", String)
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
	Method("upload", func() {
		Description("Upload a package to trigger an ingest workflow")
		Security(JWTAuth, func() {
			Scope("package:upload")
		})
		Payload(func() {
			Attribute("content_type", String, "Content-Type header, must define value for multipart boundary.", func() {
				Default("multipart/form-data; boundary=goa")
				Pattern("multipart/[^;]+; boundary=.+")
				Example("multipart/form-data; boundary=goa")
			})
			Token("token", String)
		})

		Error(
			"invalid_media_type",
			ErrorResult,
			"Error returned when the Content-Type header does not define a multipart request.",
		)
		Error(
			"invalid_multipart_request",
			ErrorResult,
			"Error returned when the request body is not a valid multipart content.",
		)
		Error("internal_error", ErrorResult, "Fault while processing upload.")

		HTTP(func() {
			POST("/upload")
			Header("content_type:Content-Type")

			// Bypass request body decoder code generation to alleviate need for
			// loading the entire request body in memory. The service gets
			// direct access to the HTTP request body reader.
			SkipRequestBodyEncodeDecode()

			// Define error HTTP statuses.
			Response("invalid_media_type", StatusBadRequest)
			Response("invalid_multipart_request", StatusBadRequest)
			Response("internal_error", StatusInternalServerError)
		})
	})
})

var EnumPackageStatus = func() {
	Enum(enums.SIPStatusInterfaces()...)
}

var Package_ = Type("Package", func() {
	Description("Package describes a package to be stored.")
	Attribute("name", String, "Name of the package")
	TypedAttributeUUID("location_id", "Identifier of storage location")
	Attribute("status", String, "Status of the package", func() {
		EnumPackageStatus()
		Default(enums.SIPStatusNew.String())
	})
	AttributeUUID("workflow_id", "Identifier of processing workflow")
	AttributeUUID("run_id", "Identifier of latest processing workflow run")
	AttributeUUID("aip_id", "Identifier of AIP")
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

var PackageList = ResultType("application/vnd.enduro.packages", func() {
	Attribute("items", CollectionOf(StoredPackage))
	Attribute("page", Page)
	Required("items", "page")
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
	Enum(enums.PreservationActionTypeInterfaces()...)
}

var EnumPreservationActionStatus = func() {
	Enum(enums.PreservationActionStatusInterfaces()...)
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
	Enum(enums.PreservationTaskStatusInterfaces()...)
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
