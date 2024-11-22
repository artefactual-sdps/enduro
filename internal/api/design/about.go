package design

import . "goa.design/goa/v3/dsl"

var _ = Service("about", func() {
	Description("The about service provides information about the system.")
	Error("unauthorized", String, "Unauthorized")
	HTTP(func() {
		Path("/about")
		Response("unauthorized", StatusUnauthorized)
	})
	Method("about", func() {
		Description("Get information about the system")
		Security(JWTAuth)
		Payload(func() {
			Token("token", String)
		})
		Result(About)
		HTTP(func() {
			GET("/")
			Response(StatusOK)
		})
	})
})

var Preprocessing = ResultType("application/vnd.enduro.preprocessing", func() {
	Attribute("enabled", Boolean)
	Attribute("workflow_name", String)
	Attribute("task_queue", String)
	Required("enabled", "workflow_name", "task_queue")
})

var Poststorage = ResultType("application/vnd.enduro.poststorage", func() {
	Attribute("workflow_name", String)
	Attribute("task_queue", String)
	Required("workflow_name", "task_queue")
})

var About = ResultType("application/vnd.enduro.about", func() {
	Attribute("version", String)
	Attribute("preservation_system", String)
	Attribute("preprocessing", Preprocessing)
	Attribute("poststorage", CollectionOf(Poststorage))
	Required("version", "preservation_system", "preprocessing")
})
