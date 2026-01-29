package design

import (
	. "goa.design/goa/v3/dsl" //nolint:staticcheck

	"github.com/artefactual-sdps/enduro/internal/enums"
)

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

var ChildWorkflow = ResultType("application/vnd.enduro.childworkflow", func() {
	Attribute("type", String, func() {
		EnumChildWorkflowType()
	})
	Attribute("task_queue", String)
	Attribute("workflow_name", String)
	Required("type", "task_queue", "workflow_name")
})

var EnumChildWorkflowType = func() {
	Enum(enums.ChildWorkflowTypeInterfaces()...)
}

var About = ResultType("application/vnd.enduro.about", func() {
	Attribute("version", String)
	Attribute("preservation_system", String)
	Attribute("child_workflows", CollectionOf(ChildWorkflow))
	Attribute("upload_max_size", Int64)
	Required("version", "preservation_system", "upload_max_size")
})
