// Package activities implements Enduro's workflow activities.
package activities

const (
	DownloadActivityName                   = "download-activity"
	DownloadFromInternalBucketActivityName = "download-from-internal-bucket-activity"
	DownloadFromSIPSourceActivityName      = "download-from-sip-source-activity"
	BundleActivityName                     = "bundle-activity"
	CleanUpActivityName                    = "clean-up-activity"
	DeleteOriginalActivityName             = "delete-original-activity"
	DisposeOriginalActivityName            = "dispose-original-activity"
	CreateStorageAIPActivityName           = "create-storage-aip-activity"
	MoveToPermanentStorageActivityName     = "move-to-permanent-storage-activity"
	PollMoveToPermanentStorageActivityName = "poll-move-to-permanent-storage-activity"
	RejectSIPActivityName                  = "reject-sip-activity"
	UploadActivityName                     = "upload-activity"
)
