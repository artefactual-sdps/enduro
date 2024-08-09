package activities

import (
	"context"
	"fmt"
	"os"

	"gocloud.dev/blob"
)

const (
	SendToFailedBucketName = "send-to-failed-bucket"
	FailureSIP             = "failure-sip"
	FailurePIP             = "failure-pip"
)

type SendToFailedBucketParams struct {
	Type string
	Path string
	Key  string
}

type SendToFailedBucketResult struct {
	FailedKey string
}

type SendToFailedBucketActivity struct {
	failedSIPs *blob.Bucket
	failedPIPs *blob.Bucket
}

func NewSendToFailedBuckeActivity(failedSIPs, failedPIPs *blob.Bucket) *SendToFailedBucketActivity {
	return &SendToFailedBucketActivity{
		failedSIPs: failedSIPs,
		failedPIPs: failedPIPs,
	}
}

func (sf *SendToFailedBucketActivity) Execute(
	ctx context.Context,
	params *SendToFailedBucketParams,
) (*SendToFailedBucketResult, error) {
	res := &SendToFailedBucketResult{}

	f, err := os.Open(params.Path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	res.FailedKey = "Failed_" + params.Key
	writerOptions := &blob.WriterOptions{
		ContentType: "application/octet-stream",
		BufferSize:  100_000_000,
	}

	switch params.Type {
	case FailureSIP:
		err = sf.failedSIPs.Upload(ctx, res.FailedKey, f, writerOptions)
	case FailurePIP:
		err = sf.failedPIPs.Upload(ctx, res.FailedKey, f, writerOptions)
	default:
		err = fmt.Errorf("SendToFailedBucketActivity: unexpected type %q", params.Type)
	}
	if err != nil {
		return nil, err
	}

	return res, nil
}
