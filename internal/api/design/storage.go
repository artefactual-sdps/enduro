package design

import (
	. "goa.design/goa/v3/dsl"

	"github.com/artefactual-sdps/enduro/internal/storage/types"
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
			POST("/package/{aip_id}/submit")
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
			POST("/package/{aip_id}/update")
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
			GET("/package/{aip_id}/download")
			Response(StatusOK)
			Response("not_found", StatusNotFound)
		})
	})
	Method("locations", func() {
		Description("List locations")
		Result(CollectionOf(StoredLocation), func() { View("default") })
		HTTP(func() {
			GET("/location")
			Response(StatusOK)
		})
	})
	Method("add_location", func() {
		Description("Add a storage location")
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
				Attribute("s3", S3Config)
			})
			Required("name", "source", "purpose")
		})
		Result(AddLocationResult)
		Error("not_valid")
		HTTP(func() {
			POST("/location")
			Response(StatusCreated)
			Response("not_valid", StatusBadRequest)
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
			POST("/package/{aip_id}/store")
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
			GET("/package/{aip_id}/store")
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
			POST("/package/{aip_id}/reject")
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
			GET("/package/{aip_id}")
			Response(StatusOK)
			Response("not_found", StatusNotFound)
		})
	})
	Method("show-location", func() {
		Description("Show location by UUID")
		Payload(func() {
			Attribute("uuid", String)
			Required("uuid")
		})
		Result(StoredLocation)
		Error("not_found", StorageLocationNotFound, "Storage location not found")
		HTTP(func() {
			GET("/location/{uuid}")
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

var StorageLocationNotFound = Type("StorageLocationNotfound", func() {
	Description("Storage location not found.")
	Attribute("message", String, "Message of error", func() {
		Meta("struct:error:name")
	})
	Attribute("uuid", String, "Identifier of missing location")
	Required("message", "uuid")
})

var StoredLocation = ResultType("application/vnd.enduro.stored-location", func() {
	Description("A StoredLocation describes a location retrieved by the storage service.")
	Reference(Location)
	TypeName("StoredLocation")

	Attributes(func() {
		Attribute("id", UInt, "ID is the unique id of the location.")
		Field(2, "name")
		Field(3, "description")
		Field(4, "source")
		Field(5, "purpose")
		Field(6, "uuid")
	})

	View("default", func() {
		Attribute("name")
		Attribute("description")
		Attribute("source")
		Attribute("purpose")
		Attribute("uuid")
	})

	Required("id", "name", "source", "purpose")
})

var EnumLocationSource = func() {
	Enum("unspecified", "minio")
}

var EnumLocationPurpose = func() {
	Enum("unspecified", "aip_store")
}

var Location = Type("Location", func() {
	Description("Location describes a physical entity used to store AIPs.")
	Attribute("id", UInt)
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
	Attribute("uuid", String)
})

var AddLocationResult = Type("AddLocationResult", func() {
	Attribute("uuid", String)
	Required("uuid")
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
