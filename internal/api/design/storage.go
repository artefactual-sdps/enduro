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
		Error("not_found", StoragePackageNotFound, "Storage package not found")
		Error("not_available")
		Error("not_valid")
		HTTP(func() {
			POST("/{aip_id}/submit")
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
		Error("not_found", StoragePackageNotFound, "Storage package not found")
		Error("not_available")
		Error("not_valid")
		HTTP(func() {
			POST("/{aip_id}/update")
			Response(StatusAccepted)
			Response("not_available", StatusConflict)
			Response("not_valid", StatusBadRequest)
		})
	})
	Method("download", func() {
		Description("Download package by AIPID")
		Payload(func() {
			Attribute("aip_id", String)
			Required("aip_id")
		})
		Result(Bytes)
		Error("not_found", StoragePackageNotFound, "Storage package not found")
		HTTP(func() {
			GET("/{aip_id}/download")
			Response(StatusOK)
			Response("not_found", StatusNotFound)
		})
	})
	Method("list", func() {
		Description("List locations")
		Result(CollectionOf(StoredLocation), func() { View("default") })
		HTTP(func() {
			GET("/location")
			Response(StatusOK)
		})
	})
	Method("move", func() {
		Description("Move a package to a permanent storage location")
		Payload(func() {
			Attribute("aip_id", String)
			Attribute("location", String)
			Required("aip_id", "location")
		})
		Error("not_found", StoragePackageNotFound, "Storage package not found")
		Error("not_available")
		Error("not_valid")
		HTTP(func() {
			POST("/{aip_id}/store")
			Response(StatusAccepted)
			Response("not_found", StatusNotFound)
			Response("not_available", StatusConflict)
			Response("not_valid", StatusBadRequest)
		})
	})
	Method("move_status", func() {
		Description("Retrieve the status of a permanent storage location move of the package")
		Payload(func() {
			Attribute("aip_id", String)
			Required("aip_id")
		})
		Result(MoveStatusResult)
		Error("not_found", StoragePackageNotFound, "Storage package not found")
		Error("failed_dependency")
		HTTP(func() {
			GET("/{aip_id}/store")
			Response(StatusOK)
			Response("not_found", StatusNotFound)
			Response("failed_dependency", StatusFailedDependency)
		})
	})
	Method("reject", func() {
		Description("Reject a package")
		Payload(func() {
			Attribute("aip_id", String)
			Required("aip_id")
		})
		Error("not_found", StoragePackageNotFound, "Storage package not found")
		Error("not_available")
		Error("not_valid")
		HTTP(func() {
			POST("/{aip_id}/reject")
			Response(StatusAccepted)
			Response("not_found", StatusNotFound)
			Response("not_available", StatusConflict)
			Response("not_valid", StatusBadRequest)
		})
	})
	Method("show", func() {
		Description("Show package by AIPID")
		Payload(func() {
			Attribute("aip_id", String)
			Required("aip_id")
		})
		Result(StoredStoragePackage)
		Error("not_found", StoragePackageNotFound, "Storage package not found")
		HTTP(func() {
			GET("/{aip_id}")
			Response(StatusOK)
			Response("not_found", StatusNotFound)
		})
	})
})

var SubmitResult = Type("SubmitResult", func() {
	Attribute("url", String)
	Required("url")
})

var StoragePackageNotFound = Type("StoragePackageNotfound", func() {
	Description("Storage package not found.")
	Attribute("message", String, "Message of error", func() {
		Meta("struct:error:name")
	})
	Attribute("aip_id", String, "Identifier of missing package")
	Required("message", "aip_id")
})

var StoredLocation = ResultType("application/vnd.enduro.stored-location", func() {
	Description("A StoredLocation describes a location retrieved by the storage service.")
	Reference(Location)
	TypeName("StoredLocation")

	Attributes(func() {
		Attribute("id", String, "ID is the unique id of the location.")
		Field(2, "name")
	})

	View("default", func() {
		Attribute("id")
		Attribute("name")
	})

	Required("id", "name")
})

var Location = Type("Location", func() {
	Description("Location describes a physical entity used to store AIPs.")
	Attribute("name", String, "Name of location")
})

var MoveStatusResult = Type("MoveStatusResult", func() {
	Attribute("done", Boolean)
	Required("done")
})

var StoredStoragePackage = ResultType("application/vnd.enduro.stored-storage-package", func() {
	Description("A StoredStoragePackage describes a package retrieved by the storage service.")
	Reference(StoragePackage)
	TypeName("StoredStoragePackage")
	Attributes(func() {
		Attribute("id", UInt)
		Attribute("name")
		Attribute("aip_id")
		Attribute("status")
		Attribute("object_key")
		Attribute("location")
	})
	View("default", func() {
		Attribute("name")
		Attribute("aip_id")
		Attribute("status")
		Attribute("object_key")
		Attribute("location")
	})
	Required("id", "name", "aip_id", "status", "object_key")
})

var EnumStoragePackageStatus = func() {
	Enum("unspecified", "in_review", "rejected", "stored", "moving")
}

var StoragePackage = Type("StoragePackage", func() {
	Description("Storage package describes a package of the storage service.")
	Attribute("id", UInt)
	Attribute("name", String)
	Attribute("aip_id", String)
	Attribute("status", String, "Status of the package", func() {
		EnumStoragePackageStatus()
		Default("unspecified")
	})
	Attribute("object_key", String)
	Attribute("location", String)
})
