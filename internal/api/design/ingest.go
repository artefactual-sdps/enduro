package design

import (
	. "goa.design/goa/v3/dsl" //nolint:staticcheck

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
			Scope("ingest:sips:read")
			Scope("ingest:sips:review")
			Scope("ingest:sips:upload")
			Scope("ingest:sips:workflows:list")
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
			Attribute("status", String, func() {
				EnumSIPStatus()
			})
			Attribute("limit", Int, "Limit number of results to return")
			Attribute("offset", Int, "Offset from the beginning of the found set")

			Token("token", String)
		})
		Result(SIPs)
		Error("not_valid")
		HTTP(func() {
			GET("/sips")
			Response(StatusOK)
			Response("not_valid", StatusBadRequest)
			Params(func() {
				Param("name")
				Param("aip_id")
				Param("earliest_created_time")
				Param("latest_created_time")
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
			AttributeUUID("uuid", "Identifier of SIP to show")
			Token("token", String)
			Required("uuid")
		})
		Result(SIP)
		Error("not_found", SIPNotFound, "SIP not found")
		Error("not_available")
		HTTP(func() {
			GET("/sips/{uuid}")
			Response(StatusOK)
			Response("not_found", StatusNotFound)
			Response("not_available", StatusConflict)
		})
	})
	Method("list_sip_workflows", func() {
		Description("List all workflows for a SIP")
		Security(JWTAuth, func() {
			Scope("ingest:sips:workflows:list")
		})
		Payload(func() {
			AttributeUUID("uuid", "Identifier of SIP to look up")
			Token("token", String)
			Required("uuid")
		})
		Result(SIPWorkflows)
		Error("not_found", SIPNotFound, "SIP not found")
		HTTP(func() {
			GET("/sips/{uuid}/workflows")
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
			AttributeUUID("uuid", "Identifier of SIP to look up")
			TypedAttributeUUID("location_id", "Identifier of storage location")
			Token("token", String)
			Required("uuid", "location_id")
		})
		Error("not_found", SIPNotFound, "SIP not found")
		Error("not_available")
		Error("not_valid")
		HTTP(func() {
			POST("/sips/{uuid}/confirm")
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
			AttributeUUID("uuid", "Identifier of SIP to look up")
			Token("token", String)
			Required("uuid")
		})
		Error("not_found", SIPNotFound, "SIP not found")
		Error("not_available")
		Error("not_valid")
		HTTP(func() {
			POST("/sips/{uuid}/reject")
			Response(StatusAccepted)
			Response("not_found", StatusNotFound)
			Response("not_available", StatusConflict)
			Response("not_valid", StatusBadRequest)
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

		Result(func() {
			AttributeUUID("uuid", "Identifier of uploaded SIP")
			Required("uuid")
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

			Response(StatusAccepted)

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
		TypedAttributeUUID("uuid", "Identifier of SIP")
		Attribute("name", String, "Name of the SIP")
		Attribute("status", String, "Status of the SIP", func() {
			EnumSIPStatus()
		})
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
		Attribute("uuid")
		Attribute("name")
		Attribute("status")
		Attribute("aip_id")
		Attribute("created_at")
		Attribute("started_at")
		Attribute("completed_at")
	})
	Required("uuid", "status", "created_at")
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
	AttributeUUID("uuid", "Identifier of missing SIP")
	Required("message", "uuid")
})

var SIPWorkflows = ResultType("application/vnd.enduro.ingest.sip.workflows", func() {
	Description("SIPWorkflows describes the workflows of a SIP.")
	TypeName("SIPWorkflows")
	Attributes(func() {
		Attribute("workflows", CollectionOf(SIPWorkflow))
	})
})

var EnumWorkflowType = func() {
	Enum(enums.WorkflowTypeInterfaces()...)
}

var EnumWorkflowStatus = func() {
	Enum(enums.WorkflowStatusInterfaces()...)
}

var SIPWorkflow = ResultType("application/vnd.enduro.ingest.sip.workflow", func() {
	Description("SIPWorkflow describes a workflow of a SIP.")
	TypeName("SIPWorkflow")
	Attributes(func() {
		Attribute("id", UInt)
		Attribute("temporal_id", String)
		Attribute("type", String, func() {
			EnumWorkflowType()
		})
		Attribute("status", String, func() {
			EnumWorkflowStatus()
		})
		Attribute("started_at", String, func() {
			Format(FormatDateTime)
		})
		Attribute("completed_at", String, func() {
			Format(FormatDateTime)
		})
		Attribute("tasks", CollectionOf(SIPTask))
		TypedAttributeUUID("sip_uuid", "Identifier of related SIP")
	})
	View("simple", func() {
		Attribute("id")
		Attribute("temporal_id")
		Attribute("type")
		Attribute("status")
		Attribute("started_at")
		Attribute("completed_at")
		Attribute("sip_uuid")
	})
	Required("id", "temporal_id", "type", "status", "started_at", "sip_uuid")
})

var EnumTaskStatus = func() {
	Enum(enums.TaskStatusInterfaces()...)
}

var SIPTask = ResultType("application/vnd.enduro.ingest.sip.task", func() {
	Description("SIPTask describes a SIP workflow task.")
	TypeName("SIPTask")
	Attributes(func() {
		Attribute("id", UInt)
		Attribute("task_id", String)
		Attribute("name", String)
		Attribute("status", String, func() {
			EnumTaskStatus()
		})
		Attribute("started_at", String, func() {
			Format(FormatDateTime)
		})
		Attribute("completed_at", String, func() {
			Format(FormatDateTime)
		})
		Attribute("note", String)
		Attribute("workflow_id", UInt)
	})
	Required("id", "task_id", "name", "status", "started_at")
})
