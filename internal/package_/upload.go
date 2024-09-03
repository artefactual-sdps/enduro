package package_

import (
	"context"
	"errors"
	"io"
	"mime"
	"mime/multipart"

	"go.artefactual.dev/tools/bucket"
	"gocloud.dev/blob"

	goapackage "github.com/artefactual-sdps/enduro/internal/api/gen/package_"
)

type UploadConfig struct {
	MaxSize int64
	Bucket  bucket.Config
}

// Validate implements config.ConfigurationValidator.
func (c UploadConfig) Validate() error {
	if c.Bucket.URL != "" && (c.Bucket.Bucket != "" || c.Bucket.Region != "") {
		return errors.New("URL and rest of the [upload.bucket] configuration options are mutually exclusive")
	}
	return nil
}

func (w *goaWrapper) Upload(ctx context.Context, payload *goapackage.UploadPayload, req io.ReadCloser) error {
	defer req.Close()

	lr := io.LimitReader(req, int64(w.uploadMaxSize))

	_, params, err := mime.ParseMediaType(payload.ContentType)
	if err != nil {
		return goapackage.MakeInvalidMediaType(errors.New("invalid media type"))
	}
	mr := multipart.NewReader(lr, params["boundary"])

	part, err := mr.NextPart()
	if err == io.EOF {
		return nil
	}
	if err != nil {
		return goapackage.MakeInvalidMultipartRequest(errors.New("invalid multipart request"))
	}

	wr, err := w.uploadBucket.NewWriter(ctx, part.FileName(), &blob.WriterOptions{})
	if err != nil {
		return err
	}

	_, copyErr := io.Copy(wr, part)
	closeErr := wr.Close()

	if copyErr != nil {
		return copyErr
	}
	if closeErr != nil {
		return closeErr
	}

	return nil
}
