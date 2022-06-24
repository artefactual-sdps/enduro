// Package activities implements Enduro's workflow activities.
package activities

const (
	DownloadActivityName                   = "download-activity"
	BundleActivityName                     = "bundle-activity"
	CleanUpActivityName                    = "clean-up-activity"
	DeleteOriginalActivityName             = "delete-original-activity"
	DisposeOriginalActivityName            = "dispose-original-activity"
	MoveToPermanentStorageActivityName     = "move-to-permanent-storage-activity"
	PollMoveToPermanentStorageActivityName = "poll-move-to-permanent-storage-activity"
	ValidateTransferActivityName           = "validate-transfer-activity"
	UploadActivityName                     = "upload-activity"
)
