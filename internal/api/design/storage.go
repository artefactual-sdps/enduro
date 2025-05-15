package design

import (
	. "goa.design/goa/v3/dsl" //nolint:staticcheck

	"github.com/artefactual-sdps/enduro/internal/storage/enums"
	"github.com/artefactual-sdps/enduro/internal/storage/types"
)

var _ = Service("storage", func() {
	Description("The storage service manages locations and AIPs.")
	Error("unauthorized", String, "Unauthorized")
	Error("forbidden", String, "Forbidden")
	HTTP(func() {
		Path("/storage")
		Response("unauthorized", StatusUnauthorized)
		Response("forbidden", StatusForbidden)
	})
	Method("list_aips", func() {
		Description("List all AIPs")
		Security(JWTAuth, func() {
			Scope("storage:aips:list")
		})
		Payload(func() {
			Attribute("name", String)
			Attribute("earliest_created_time", String, func() {
				Format(FormatDateTime)
			})
			Attribute("latest_created_time", String, func() {
				Format(FormatDateTime)
			})
			Attribute("status", String, func() {
				EnumAIPStatus()
			})
			Attribute("limit", Int, "Limit number of results to return")
			Attribute("offset", Int, "Offset from the beginning of the found set")

			Token("token", String)
		})
		Result(AIPs)
		Error("not_available")
		Error("not_valid")
		HTTP(func() {
			GET("/aips")
			Response(StatusOK)
			Response("not_available", StatusConflict)
			Response("not_valid", StatusBadRequest)
			Params(func() {
				Param("name")
				Param("earliest_created_time")
				Param("latest_created_time")
				Param("status")
				Param("limit")
				Param("offset")
			})
		})
	})
	Method("create_aip", func() {
		Description("Create a new AIP")
		Security(JWTAuth, func() {
			Scope("storage:aips:create")
		})
		Payload(func() {
			AttributeUUID("uuid", "Identifier of the AIP")
			Attribute("name", String, "Name of the AIP")
			AttributeUUID("object_key", "ObjectKey of the AIP")
			Attribute("status", String, "Status of the the AIP", func() {
				EnumAIPStatus()
				Default("unspecified")
			})
			TypedAttributeUUID("location_id", "Identifier of the AIP's storage location")
			Token("token", String)
			Required("uuid", "name", "object_key")
		})
		Result(AIP)
		Error("not_valid")
		HTTP(func() {
			POST("/aips")
			Response(StatusOK)
			Response("not_valid", StatusBadRequest)
		})
	})
	Method("submit_aip", func() {
		Description("Start the submission of an AIP")
		Security(JWTAuth, func() {
			Scope("storage:aips:submit")
		})
		Payload(func() {
			AttributeUUID("uuid", "Identifier of AIP")
			Attribute("name", String)
			Token("token", String)
			Required("uuid", "name")
		})
		Result(SubmitAIPResult)
		Error("not_found", AIPNotFound, "AIP not found")
		Error("not_available")
		Error("not_valid")
		HTTP(func() {
			POST("/aips/{uuid}/submit")
			Response(StatusAccepted)
			Response("not_available", StatusConflict)
			Response("not_valid", StatusBadRequest)
		})
	})
	Method("update_aip", func() {
		Description("Signal that an AIP submission is complete")
		Security(JWTAuth, func() {
			Scope("storage:aips:submit")
		})
		Payload(func() {
			AttributeUUID("uuid", "Identifier of AIP")
			Token("token", String)
			Required("uuid")
		})
		Error("not_found", AIPNotFound, "AIP not found")
		Error("not_available")
		Error("not_valid")
		HTTP(func() {
			POST("/aips/{uuid}/update")
			Response(StatusAccepted)
			Response("not_available", StatusConflict)
			Response("not_valid", StatusBadRequest)
		})
	})
	Method("download_aip", func() {
		Description("Download AIP by AIPID")
		Security(JWTAuth, func() {
			Scope("storage:aips:download")
		})
		Payload(func() {
			AttributeUUID("uuid", "Identifier of AIP")
			Token("token", String)
			Required("uuid")
		})
		Result(Bytes)
		Error("not_found", AIPNotFound, "AIP not found")
		HTTP(func() {
			GET("/aips/{uuid}/download")
			Response(StatusOK)
			Response("not_found", StatusNotFound)
		})
	})
	Method("move_aip", func() {
		Description("Move an AIP to a permanent storage location")
		Security(JWTAuth, func() {
			Scope("storage:aips:move")
		})
		Payload(func() {
			AttributeUUID("uuid", "Identifier of AIP")
			TypedAttributeUUID("location_id", "Identifier of storage location")
			Token("token", String)
			Required("uuid", "location_id")
		})
		Error("not_found", AIPNotFound, "AIP not found")
		Error("not_available")
		Error("not_valid")
		HTTP(func() {
			POST("/aips/{uuid}/store")
			Response(StatusAccepted)
			Response("not_found", StatusNotFound)
			Response("not_available", StatusConflict)
			Response("not_valid", StatusBadRequest)
		})
	})
	Method("move_aip_status", func() {
		Description("Retrieve the status of a permanent storage location move of the AIP")
		Security(JWTAuth, func() {
			Scope("storage:aips:move")
		})
		Payload(func() {
			AttributeUUID("uuid", "Identifier of AIP")
			Token("token", String)
			Required("uuid")
		})
		Result(MoveStatusResult)
		Error("not_found", AIPNotFound, "AIP not found")
		Error("failed_dependency")
		HTTP(func() {
			GET("/aips/{uuid}/store")
			Response(StatusOK)
			Response("not_found", StatusNotFound)
			Response("failed_dependency", StatusFailedDependency)
		})
	})
	Method("reject_aip", func() {
		Description("Reject an AIP")
		Security(JWTAuth, func() {
			Scope("storage:aips:review")
		})
		Payload(func() {
			AttributeUUID("uuid", "Identifier of AIP")
			Token("token", String)
			Required("uuid")
		})
		Error("not_found", AIPNotFound, "AIP not found")
		Error("not_available")
		Error("not_valid")
		HTTP(func() {
			POST("/aips/{uuid}/reject")
			Response(StatusAccepted)
			Response("not_found", StatusNotFound)
			Response("not_available", StatusConflict)
			Response("not_valid", StatusBadRequest)
		})
	})
	Method("show_aip", func() {
		Description("Show AIP by AIPID")
		Security(JWTAuth, func() {
			Scope("storage:aips:read")
		})
		Payload(func() {
			AttributeUUID("uuid", "Identifier of AIP")
			Token("token", String)
			Required("uuid")
		})
		Result(AIP)
		Error("not_found", AIPNotFound, "AIP not found")
		HTTP(func() {
			GET("/aips/{uuid}")
			Response(StatusOK)
			Response("not_found", StatusNotFound)
		})
	})
	Method("list_aip_workflows", func() {
		Description("List workflows related to an AIP")
		Security(JWTAuth, func() {
			Scope("storage:aips:workflows:list")
		})
		Payload(func() {
			AttributeUUID("uuid", "Identifier of AIP")
			Attribute("status", String, func() {
				EnumAIPWorkflowStatus()
			})
			Attribute("type", String, func() {
				EnumAIPWorkflowType()
			})
			Token("token", String)
			Required("uuid")
		})
		Result(AIPWorkflows)
		Error("not_found", AIPNotFound, "AIP not found")
		HTTP(func() {
			GET("/aips/{uuid}/workflows")
			Response(StatusOK)
			Response("not_found", StatusNotFound)
		})
	})
	Method("request_aip_deletion", func() {
		Description("Request an AIP deletion")
		Security(JWTAuth, func() {
			Scope("storage:aips:deletion:request")
		})
		Payload(func() {
			AttributeUUID("uuid", "Identifier of AIP")
			Token("token", String)
			Attribute("reason", String)
			Required("uuid", "reason")
		})
		Error("not_found", AIPNotFound, "AIP not found")
		HTTP(func() {
			POST("/aips/{uuid}/deletion-request")
			Response(StatusOK)
			Response("not_found", StatusNotFound)
		})
	})
	Method("review_aip_deletion", func() {
		Description("Review an AIP deletion request")
		Security(JWTAuth, func() {
			Scope("storage:aips:deletion:review")
		})
		Payload(func() {
			AttributeUUID("uuid", "Identifier of AIP")
			Token("token", String)
			Attribute("approved", Boolean)
			Required("uuid", "approved")
		})
		Error("not_found", AIPNotFound, "AIP not found")
		HTTP(func() {
			POST("/aips/{uuid}/deletion-review")
			Response(StatusOK)
			Response("not_found", StatusNotFound)
		})
	})
	Method("cancel_aip_deletion", func() {
		Description("Cancel an AIP deletion request")
		Security(JWTAuth, func() {
			Scope("storage:aips:deletion:request")
		})
		Payload(func() {
			AttributeUUID("uuid", "Identifier of AIP")
			Token("token", String)
			Attribute(
				"check",
				Boolean,
				"If check is true, check user authorization to cancel deletion but don't execute the cancellation.",
			)
			Required("uuid")
		})
		Error("not_found", AIPNotFound, "AIP not found")
		HTTP(func() {
			POST("/aips/{uuid}/deletion-cancel")
			Response(StatusOK)
			Response("not_found", StatusNotFound)
		})
	})
	Method("list_locations", func() {
		Description("List locations")
		Security(JWTAuth, func() {
			Scope("storage:locations:list")
		})
		Payload(func() {
			Token("token", String)
		})
		Result(CollectionOf(Location), func() { View("default") })
		HTTP(func() {
			GET("/locations")
			Response(StatusOK)
		})
	})
	Method("create_location", func() {
		Description("Create a storage location")
		Security(JWTAuth, func() {
			Scope("storage:locations:create")
		})
		Payload(func() {
			Attribute("name", String)
			Attribute("description", String)
			Attribute("source", String, func() {
				EnumLocationSource()
			})
			Attribute("purpose", String, func() {
				EnumLocationPurpose()
			})
			OneOf("config", func() {
				Attribute("amss", AMSSConfig)
				Attribute("s3", S3Config)
				Attribute("sftp", SFTPConfig)
				Attribute("url", URLConfig)
			})
			Token("token", String)
			Required("name", "source", "purpose")
		})
		Result(CreateLocationResult)
		Error("not_valid")
		HTTP(func() {
			POST("/locations")
			Response(StatusCreated)
			Response("not_valid", StatusBadRequest)
		})
	})
	Method("show_location", func() {
		Description("Show location by UUID")
		Security(JWTAuth, func() {
			Scope("storage:locations:read")
		})
		Payload(func() {
			// TODO: explore how we can use uuid.UUID that are also URL params.
			AttributeUUID("uuid", "Identifier of location")
			Token("token", String)
			Required("uuid")
		})
		Result(Location)
		Error("not_found", LocationNotFound, "Storage location not found")
		HTTP(func() {
			GET("/locations/{uuid}")
			Response(StatusOK)
			Response("not_found", StatusNotFound)
		})
	})
	Method("list_location_aips", func() {
		Description("List all the AIPs stored in the location with UUID")
		Security(JWTAuth, func() {
			Scope("storage:locations:aips:list")
		})
		Payload(func() {
			// TODO: explore how we can use uuid.UUID that are also URL params.
			AttributeUUID("uuid", "Identifier of location")
			Token("token", String)
			Required("uuid")
		})
		Result(CollectionOf(AIP), func() { View("default") })
		Error("not_found", LocationNotFound, "Storage location not found")
		Error("not_valid")
		HTTP(func() {
			GET("/locations/{uuid}/aips")
			Response(StatusOK)
			Response("not_found", StatusNotFound)
			Response("not_valid", StatusBadRequest)
		})
	})
})

var SubmitAIPResult = Type("SubmitAIPResult", func() {
	TypeName("SubmitAIPResult")
	Attribute("url", String)
	Required("url")
})

var AIPNotFound = Type("AIPNotFound", func() {
	Description("AIP not found.")
	TypeName("AIPNotFound")
	Attribute("message", String, "Message of error", func() {
		Meta("struct:error:message")
	})
	Attribute("uuid", String, "Identifier of missing AIP", func() {
		Meta("struct:field:type", "uuid.UUID", "github.com/google/uuid")
	})
	Required("message", "uuid")
})

var LocationNotFound = Type("LocationNotFound", func() {
	Description("Storage location not found.")
	TypeName("LocationNotFound")
	Attribute("message", String, "Message of error", func() {
		Meta("struct:error:message")
	})
	Attribute("uuid", String, func() {
		Meta("struct:field:type", "uuid.UUID", "github.com/google/uuid")
	})
	Required("message", "uuid")
})

var Location = ResultType("application/vnd.enduro.storage.location", func() {
	Description("A Location describes a location retrieved by the storage service.")
	TypeName("Location")

	Attributes(func() {
		Attribute("name", String, "Name of location")
		Attribute("description", String, "Description of the location")
		Attribute("source", String, "Data source of the location", func() {
			EnumLocationSource()
			Default("unspecified")
		})
		Attribute("purpose", String, "Purpose of the location", func() {
			EnumLocationPurpose()
			Default("unspecified")
		})
		Attribute("uuid", String, func() {
			Meta("struct:field:type", "uuid.UUID", "github.com/google/uuid")
		})
		OneOf("config", func() {
			Attribute("amss", AMSSConfig)
			Attribute("s3", S3Config)
			Attribute("sftp", SFTPConfig)
			Attribute("url", URLConfig)
		})
		Attribute("created_at", String, "Creation datetime", func() {
			Format(FormatDateTime)
		})
		Required("name", "source", "purpose", "uuid", "created_at")
	})

	View("default", func() {
		Attribute("name")
		Attribute("description")
		Attribute("source")
		Attribute("purpose")
		Attribute("uuid")
		Attribute("created_at")
	})
})

var EnumLocationPurpose = func() {
	Enum(enums.LocationPurposeInterfaces()...)
}

var EnumLocationSource = func() {
	Enum(enums.LocationSourceInterfaces()...)
}

var CreateLocationResult = Type("CreateLocationResult", func() {
	Attribute("uuid", String)
	Required("uuid")
})

var MoveStatusResult = Type("MoveStatusResult", func() {
	Attribute("done", Boolean)
	Required("done")
})

var AIP = ResultType("application/vnd.enduro.storage.aip", func() {
	Description("An AIP describes an AIP retrieved by the storage service.")
	TypeName("AIP")

	Attributes(func() {
		Attribute("name", String)
		Attribute("uuid", String, func() {
			Meta("struct:field:type", "uuid.UUID", "github.com/google/uuid")
		})
		Attribute("status", String, "Status of the AIP", func() {
			EnumAIPStatus()
			Default("unspecified")
		})
		Attribute("object_key", String, func() {
			Meta("struct:field:type", "uuid.UUID", "github.com/google/uuid")
		})
		TypedAttributeUUID("location_id", "Identifier of storage location")
		Attribute("created_at", String, "Creation datetime", func() {
			Format(FormatDateTime)
		})
		Required("name", "uuid", "status", "object_key", "created_at")
	})

	View("default", func() {
		Attribute("name")
		Attribute("uuid")
		Attribute("status")
		Attribute("object_key")
		Attribute("location_id")
		Attribute("created_at")
	})
})

var AIPs = ResultType("application/vnd.enduro.storage.aips", func() {
	TypeName("AIPs")
	Attribute("items", CollectionOf(AIP))
	Attribute("page", Page)
	Required("items", "page")
})

var EnumAIPStatus = func() {
	Enum(enums.AIPStatusInterfaces()...)
}

var AMSSConfig = Type("AMSSConfig", func() {
	ConvertTo(types.AMSSConfig{})

	Attribute("api_key", String)
	Attribute("url", String)
	Attribute("username", String)

	Required("api_key", "url", "username")
})

var S3Config = Type("S3Config", func() {
	ConvertTo(types.S3Config{})

	Attribute("bucket", String)
	Attribute("region", String)
	Attribute("endpoint", String)
	Attribute("path_style", Boolean)
	Attribute("profile", String)
	Attribute("key", String)
	Attribute("secret", String)
	Attribute("token", String)

	Required("bucket", "region")
})

var SFTPConfig = Type("SFTPConfig", func() {
	ConvertTo(types.SFTPConfig{})

	Attribute("address", String)
	Attribute("username", String)
	Attribute("password", String)
	Attribute("directory", String)

	Required("address", "username", "password", "directory")
})

var URLConfig = Type("URLConfig", func() {
	ConvertTo(types.URLConfig{})
	Attribute("url", String)
	Required("url")
})

var AIPWorkflows = ResultType("application/vnd.enduro.storage.aip.workflows", func() {
	Description("AIPWorkflows describes the workflows of an AIP.")
	TypeName("AIPWorkflows")
	Attributes(func() {
		Attribute("workflows", CollectionOf(AIPWorkflow))
	})
})

var EnumAIPWorkflowType = func() {
	Enum(enums.WorkflowTypeInterfaces()...)
}

var EnumAIPWorkflowStatus = func() {
	Enum(enums.WorkflowStatusInterfaces()...)
}

var AIPWorkflow = ResultType("application/vnd.enduro.storage.aip.workflow", func() {
	Description("AIPWorkflow describes a workflow of an AIP.")
	TypeName("AIPWorkflow")
	Attributes(func() {
		Attribute("uuid", String, func() {
			Meta("struct:field:type", "uuid.UUID", "github.com/google/uuid")
		})
		Attribute("temporal_id", String)
		Attribute("type", String, func() {
			EnumAIPWorkflowType()
		})
		Attribute("status", String, func() {
			EnumAIPWorkflowStatus()
		})
		Attribute("started_at", String, func() {
			Format(FormatDateTime)
		})
		Attribute("completed_at", String, func() {
			Format(FormatDateTime)
		})
		Attribute("tasks", CollectionOf(AIPTask))
	})
	View("simple", func() {
		Attribute("uuid")
		Attribute("temporal_id")
		Attribute("type")
		Attribute("status")
		Attribute("started_at")
		Attribute("completed_at")
	})
	Required("uuid", "temporal_id", "type", "status")
})

var EnumAIPTaskStatus = func() {
	Enum(enums.TaskStatusInterfaces()...)
}

var AIPTask = ResultType("application/vnd.enduro.storage.aip.task", func() {
	Description("AIPTask describes an AIP workflow task.")
	TypeName("AIPTask")
	Attributes(func() {
		Attribute("uuid", String, func() {
			Meta("struct:field:type", "uuid.UUID", "github.com/google/uuid")
		})
		Attribute("name", String)
		Attribute("status", String, func() {
			EnumAIPTaskStatus()
		})
		Attribute("started_at", String, func() {
			Format(FormatDateTime)
		})
		Attribute("completed_at", String, func() {
			Format(FormatDateTime)
		})
		Attribute("note", String)
	})
	Required("uuid", "name", "status")
})
