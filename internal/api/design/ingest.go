package design

import (
	. "goa.design/goa/v3/dsl"

	"github.com/artefactual-sdps/enduro/internal/enums"
)

var _ = Service("ingest", func() {
	Description("The ingest service manages ingested SIPs.")
	Error("unauthorized", String, "Unauthorized")
	Error("forbidden", String, "Forbidden")
	HTTP(func() {
		Path("/ingest")
		Response("unauthorized", StatusUnauthorized)
		Response("forbidden", StatusForbidden)
	})
	Method("monitor_request", func() {
		Description("Request access to the /monitor WebSocket")
		// For now, the monitor websocket requires all the scopes from this service.
		Security(JWTAuth, func() {
			Scope("ingest:sips:list")
			Scope("ingest:sips:actions:list")
			Scope("ingest:sips:move")
			Scope("ingest:sips:read")
			Scope("ingest:sips:review")
			Scope("ingest:sips:upload")
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
	Method("list_sips", func() {
		Description("List all ingested SIPs")
		Security(JWTAuth, func() {
			Scope("ingest:sips:list")
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
				EnumSIPStatus()
			})
			Attribute("limit", Int, "Limit number of results to return")
			Attribute("offset", Int, "Offset from the beginning of the found set")

			Token("token", String)
		})
		Result(SIPs)
		HTTP(func() {
			GET("/sips")
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
	Method("show_sip", func() {
		Description("Show SIP by ID")
		Security(JWTAuth, func() {
			Scope("ingest:sips:read")
		})
		Payload(func() {
			Attribute("id", UInt, "Identifier of SIP to show")
			Token("token", String)
			Required("id")
		})
		Result(SIP)
		Error("not_found", SIPNotFound, "SIP not found")
		Error("not_available")
		HTTP(func() {
			GET("/sips/{id}")
			Response(StatusOK)
			Response("not_found", StatusNotFound)
			Response("not_available", StatusConflict)
		})
	})
	Method("list_sip_preservation_actions", func() {
		Description("List all preservation actions for a SIP")
		Security(JWTAuth, func() {
			Scope("ingest:sips:actions:list")
		})
		Payload(func() {
			Attribute("id", UInt, "Identifier of SIP to look up")
			Token("token", String)
			Required("id")
		})
		Result(SIPPreservationActions)
		Error("not_found", SIPNotFound, "SIP not found")
		HTTP(func() {
			GET("/sips/{id}/preservation-actions")
			Response(StatusOK)
			Response("not_found", StatusNotFound)
		})
	})
	Method("confirm_sip", func() {
		Description("Signal the SIP has been reviewed and accepted")
		Security(JWTAuth, func() {
			Scope("ingest:sips:review")
		})
		Payload(func() {
			Attribute("id", UInt, "Identifier of SIP to look up")
			TypedAttributeUUID("location_id", "Identifier of storage location")
			Token("token", String)
			Required("id", "location_id")
		})
		Error("not_found", SIPNotFound, "SIP not found")
		Error("not_available")
		Error("not_valid")
		HTTP(func() {
			POST("/sips/{id}/confirm")
			Response(StatusAccepted)
			Response("not_found", StatusNotFound)
			Response("not_available", StatusConflict)
			Response("not_valid", StatusBadRequest)
		})
	})
	Method("reject_sip", func() {
		Description("Signal the SIP has been reviewed and rejected")
		Security(JWTAuth, func() {
			Scope("ingest:sips:review")
		})
		Payload(func() {
			Attribute("id", UInt, "Identifier of SIP to look up")
			Token("token", String)
			Required("id")
		})
		Error("not_found", SIPNotFound, "SIP not found")
		Error("not_available")
		Error("not_valid")
		HTTP(func() {
			POST("/sips/{id}/reject")
			Response(StatusAccepted)
			Response("not_found", StatusNotFound)
			Response("not_available", StatusConflict)
			Response("not_valid", StatusBadRequest)
		})
	})
	Method("move_sip", func() {
		Description("Move a SIP to a permanent storage location")
		Security(JWTAuth, func() {
			Scope("ingest:sips:move")
		})
		Payload(func() {
			Attribute("id", UInt, "Identifier of SIP to move")
			TypedAttributeUUID("location_id", "Identifier of storage location")
			Token("token", String)
			Required("id", "location_id")
		})
		Error("not_found", SIPNotFound, "SIP not found")
		Error("not_available")
		Error("not_valid")
		HTTP(func() {
			POST("/sips/{id}/move")
			Response(StatusAccepted)
			Response("not_found", StatusNotFound)
			Response("not_available", StatusConflict)
			Response("not_valid", StatusBadRequest)
		})
	})
	Method("move_sip_status", func() {
		Description("Retrieve the status of a permanent storage location move of the SIP")
		Security(JWTAuth, func() {
			Scope("ingest:sips:move")
		})
		Payload(func() {
			Attribute("id", UInt, "Identifier of SIP to move")
			Token("token", String)
			Required("id")
		})
		Result(MoveStatusResult)
		Error("not_found", SIPNotFound, "SIP not found")
		Error("failed_dependency")
		HTTP(func() {
			GET("/sips/{id}/move")
			Response(StatusOK)
			Response("not_found", StatusNotFound)
			Response("failed_dependency", StatusFailedDependency)
		})
	})
	Method("upload_sip", func() {
		Description("Upload a SIP to trigger an ingest workflow")
		Security(JWTAuth, func() {
			Scope("ingest:sips:upload")
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
			POST("/sips/upload")
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

var EnumSIPStatus = func() {
	Enum(enums.SIPStatusInterfaces()...)
}

var SIP = ResultType("application/vnd.enduro.ingest.sip", func() {
	Description("SIP describes an ingest SIP type.")
	TypeName("SIP")
	Attributes(func() {
		Attribute("id", UInt, "Identifier of SIP")
		Attribute("name", String, "Name of the SIP")
		TypedAttributeUUID("location_id", "Identifier of storage location")
		Attribute("status", String, "Status of the SIP", func() {
			EnumSIPStatus()
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

var SIPs = ResultType("application/vnd.enduro.ingest.sips", func() {
	TypeName("SIPs")
	Attribute("items", CollectionOf(SIP))
	Attribute("page", Page)
	Required("items", "page")
})

var SIPNotFound = Type("SIPNotFound", func() {
	Description("SIP not found.")
	TypeName("SIPNotFound")
	Attribute("message", String, "Message of error", func() {
		Meta("struct:error:message")
	})
	Attribute("id", UInt, "Identifier of missing SIP")
	Required("message", "id")
})

var SIPPreservationActions = ResultType("application/vnd.enduro.ingest.sip.preservation-actions", func() {
	Description("SIPPreservationActions describes the preservation actions of a SIP.")
	TypeName("SIPPreservationActions")
	Attributes(func() {
		Attribute("actions", CollectionOf(SIPPreservationAction))
	})
})

var EnumPreservationActionType = func() {
	Enum(enums.PreservationActionTypeInterfaces()...)
}

var EnumPreservationActionStatus = func() {
	Enum(enums.PreservationActionStatusInterfaces()...)
}

var SIPPreservationAction = ResultType("application/vnd.enduro.ingest.sip.preservation-action", func() {
	Description("SIPPreservationAction describes a preservation action of a SIP.")
	TypeName("SIPPreservationAction")
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
		Attribute("tasks", CollectionOf(SIPPreservationTask))
		Attribute("sip_id", UInt)
	})
	View("simple", func() {
		Attribute("id")
		Attribute("workflow_id")
		Attribute("type")
		Attribute("status")
		Attribute("started_at")
		Attribute("completed_at")
		Attribute("sip_id")
	})
	Required("id", "workflow_id", "type", "status", "started_at")
})

var EnumPreservationTaskStatus = func() {
	Enum(enums.PreservationTaskStatusInterfaces()...)
}

var SIPPreservationTask = ResultType("application/vnd.enduro.ingest.sip.preservation-task", func() {
	Description("SIPPreservationTask describes a SIP preservation action task.")
	TypeName("SIPPreservationTask")
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
