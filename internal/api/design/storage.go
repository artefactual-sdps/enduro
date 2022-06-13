package design

import (
	. "goa.design/goa/v3/dsl"
)

var _ = Service("storage", func() {
	Description("The storage service manages XXX.")
	HTTP(func() {
		Path("/storage")
	})
	Method("submit", func() {
		Description("XXX")
		Payload(func() {
			Attribute("key", String)
			Required("key")
		})
		Result(SubmitResult)
		Error("not_available")
		Error("not_valid")
		HTTP(func() {
			POST("/submit")
			Response(StatusAccepted)
			Response("not_available", StatusConflict)
			Response("not_valid", StatusBadRequest)
		})
	})
	Method("update", func() {
		Description("Signal the storage service that an upload is complete")
		Payload(func() {
			Attribute("workflow_id", String)
			Required("workflow_id")
		})
		Result(UpdateResult)
		Error("not_available")
		Error("not_valid")
		HTTP(func() {
			POST("/update")
			Response(StatusAccepted)
			Response("not_available", StatusConflict)
			Response("not_valid", StatusBadRequest)
		})
	})
})

var SubmitResult = Type("SubmitResult", func() {
	Attribute("url", String)
	Attribute("workflow_id", String)
	Required("url", "workflow_id")
})

var UpdateResult = Type("UpdateResult", func() {
	Attribute("ok", Boolean)
	Required("ok")
})
