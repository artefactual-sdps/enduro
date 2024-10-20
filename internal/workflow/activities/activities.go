// Package activities implements Enduro's workflow activities.
package activities

const (
	DownloadActivityName                   = "download-activity"
	BundleActivityName                     = "bundle-activity"
	CleanUpActivityName                    = "clean-up-activity"
	DeleteOriginalActivityName             = "delete-original-activity"
	DisposeOriginalActivityName            = "dispose-original-activity"
	CreateStoragePackageActivityName       = "create-storage-package-activity"
	MoveToPermanentStorageActivityName     = "move-to-permanent-storage-activity"
	PollMoveToPermanentStorageActivityName = "poll-move-to-permanent-storage-activity"
	RejectPackageActivityName              = "reject-package-activity"
	UploadActivityName                     = "upload-activity"
	SendToFailedSIPsName                   = "send-to-failed-sips"
	SendToFailedPIPsName                   = "send-to-failed-pips"
)
