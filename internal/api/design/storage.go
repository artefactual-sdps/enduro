package design

import (
	. "goa.design/goa/v3/dsl"
)

var _ = Service("storage", func() {
	Description("The storage service manages the storage of packages.")
	HTTP(func() {
		Path("/storage")
	})
	Method("submit", func() {
		Description("Start the submission of a package")
		Payload(func() {
			Attribute("aip_id", String)
			Attribute("name", String)
			Required("aip_id", "name")
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
			Attribute("aip_id", String)
			Required("aip_id")
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
	Required("url")
})

var UpdateResult = Type("UpdateResult", func() {
	Attribute("ok", Boolean)
	Required("ok")
})
