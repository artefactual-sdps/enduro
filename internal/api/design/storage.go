package design

import (
	. "goa.design/goa/v3/dsl"

	"github.com/artefactual-sdps/enduro/internal/storage/types"
)

var _ = Service("storage", func() {
	Description("The storage service manages the storage of packages.")
	Error("unauthorized", String, "Unauthorized")
	Error("forbidden", String, "Forbidden")
	HTTP(func() {
		Path("/storage")
		Response("unauthorized", StatusUnauthorized)
		Response("forbidden", StatusForbidden)
	})
	Method("create", func() {
		Description("Create a new package")
		Security(JWTAuth, func() {
			Scope("storage:package:create")
		})
		Payload(func() {
			AttributeUUID("aip_id", "Identifier of AIP")
			Attribute("name", String, "Name of the package")
			AttributeUUID("object_key", "ObjectKey of AIP")
			Attribute("status", String, "Status of the package", func() {
				EnumStoragePackageStatus()
				Default("unspecified")
			})
			TypedAttributeUUID("location_id", "Identifier of the package's storage location")
			Token("token", String)
			Required("aip_id", "name", "object_key")
		})
		Result(StoragePackage)
		Error("not_valid")
		HTTP(func() {
			POST("/package")
			Response(StatusOK)
			Response("not_valid", StatusBadRequest)
		})
	})
	Method("submit", func() {
		Description("Start the submission of a package")
		Security(JWTAuth, func() {
			Scope("storage:package:submit")
		})
		Payload(func() {
			AttributeUUID("aip_id", "Identifier of AIP")
			Attribute("name", String)
			Token("token", String)
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
		Description("Signal that a package submission is complete")
		Security(JWTAuth, func() {
			Scope("storage:package:submit")
		})
		Payload(func() {
			AttributeUUID("aip_id", "Identifier of AIP")
			Token("token", String)
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
		Security(JWTAuth, func() {
			Scope("storage:package:download")
		})
		Payload(func() {
			AttributeUUID("aip_id", "Identifier of AIP")
			Token("token", String)
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
	Method("move", func() {
		Description("Move a package to a permanent storage location")
		Security(JWTAuth, func() {
			Scope("storage:package:move")
		})
		Payload(func() {
			AttributeUUID("aip_id", "Identifier of AIP")
			TypedAttributeUUID("location_id", "Identifier of storage location")
			Token("token", String)
			Required("aip_id", "location_id")
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
		Security(JWTAuth, func() {
			Scope("storage:package:move")
		})
		Payload(func() {
			AttributeUUID("aip_id", "Identifier of AIP")
			Token("token", String)
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
		Security(JWTAuth, func() {
			Scope("storage:package:review")
		})
		Payload(func() {
			AttributeUUID("aip_id", "Identifier of AIP")
			Token("token", String)
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
		Security(JWTAuth, func() {
			Scope("storage:package:read")
		})
		Payload(func() {
			AttributeUUID("aip_id", "Identifier of AIP")
			Token("token", String)
			Required("aip_id")
		})
		Result(StoragePackage)
		Error("not_found", StoragePackageNotFound, "Storage package not found")
		HTTP(func() {
			GET("/package/{aip_id}")
			Response(StatusOK)
			Response("not_found", StatusNotFound)
		})
	})
	Method("locations", func() {
		Description("List locations")
		Security(JWTAuth, func() {
			Scope("storage:location:list")
		})
		Payload(func() {
			Token("token", String)
		})
		Result(CollectionOf(Location), func() { View("default") })
		HTTP(func() {
			GET("/location")
			Response(StatusOK)
		})
	})
	Method("add_location", func() {
		Description("Create a storage location")
		Security(JWTAuth, func() {
			Scope("storage:location:create")
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
		Result(AddLocationResult)
		Error("not_valid")
		HTTP(func() {
			POST("/location")
			Response(StatusCreated)
			Response("not_valid", StatusBadRequest)
		})
	})
	Method("show_location", func() {
		Description("Show location by UUID")
		Security(JWTAuth, func() {
			Scope("storage:location:read")
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
			GET("/location/{uuid}")
			Response(StatusOK)
			Response("not_found", StatusNotFound)
		})
	})
	Method("location_packages", func() {
		Description("List all the packages stored in the location with UUID")
		Security(JWTAuth, func() {
			Scope("storage:location:listPackages")
		})
		Payload(func() {
			// TODO: explore how we can use uuid.UUID that are also URL params.
			AttributeUUID("uuid", "Identifier of location")
			Token("token", String)
			Required("uuid")
		})
		Result(CollectionOf(StoragePackage), func() { View("default") })
		Error("not_found", LocationNotFound, "Storage location not found")
		Error("not_valid")
		HTTP(func() {
			GET("/location/{uuid}/packages")
			Response(StatusOK)
			Response("not_found", StatusNotFound)
			Response("not_valid", StatusBadRequest)
		})
	})
})

var SubmitResult = Type("SubmitResult", func() {
	Attribute("url", String)
	Required("url")
})

var StoragePackageNotFound = Type("StoragePackageNotFound", func() {
	Description("Storage package not found.")
	TypeName("PackageNotFound")
	Attribute("message", String, "Message of error", func() {
		Meta("struct:error:message")
	})
	Attribute("aip_id", String, "Identifier of missing package", func() {
		Meta("struct:field:type", "uuid.UUID", "github.com/google/uuid")
	})
	Required("message", "aip_id")
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

var Location = ResultType("application/vnd.enduro.storage-location", func() {
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

var EnumLocationSource = func() {
	Enum("unspecified", "minio", "sftp", "amss")
}

var EnumLocationPurpose = func() {
	Enum("unspecified", "aip_store")
}

var AddLocationResult = Type("AddLocationResult", func() {
	Attribute("uuid", String)
	Required("uuid")
})

var MoveStatusResult = Type("MoveStatusResult", func() {
	Attribute("done", Boolean)
	Required("done")
})

var StoragePackage = ResultType("application/vnd.enduro.storage-package", func() {
	Description("A Package describes a package retrieved by the storage service.")
	TypeName("Package")

	Attributes(func() {
		Attribute("name", String)
		Attribute("aip_id", String, func() {
			Meta("struct:field:type", "uuid.UUID", "github.com/google/uuid")
		})
		Attribute("status", String, "Status of the package", func() {
			EnumStoragePackageStatus()
			Default("unspecified")
		})
		Attribute("object_key", String, func() {
			Meta("struct:field:type", "uuid.UUID", "github.com/google/uuid")
		})
		TypedAttributeUUID("location_id", "Identifier of storage location")
		Attribute("created_at", String, "Creation datetime", func() {
			Format(FormatDateTime)
		})
		Required("name", "aip_id", "status", "object_key", "created_at")
	})

	View("default", func() {
		Attribute("name")
		Attribute("aip_id")
		Attribute("status")
		Attribute("object_key")
		Attribute("location_id")
		Attribute("created_at")
	})
})

var EnumStoragePackageStatus = func() {
	Enum("unspecified", "in_review", "rejected", "stored", "moving")
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
