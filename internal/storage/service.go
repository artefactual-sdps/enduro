package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/go-logr/logr"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	temporalsdk_client "go.temporal.io/sdk/client"
	goahttp "goa.design/goa/v3/http"
	"gocloud.dev/blob"
	"gocloud.dev/blob/s3blob"

	"github.com/artefactual-labs/enduro/internal/api/gen/http/storage/server"
	goastorage "github.com/artefactual-labs/enduro/internal/api/gen/storage"
)

var submitURLExpirationTime = 15 * time.Minute

type Service interface {
	Submit(context.Context, *goastorage.SubmitPayload) (res *goastorage.SubmitResult, err error)
	Update(context.Context, *goastorage.UpdatePayload) (res *goastorage.UpdateResult, err error)
	Download(context.Context, *goastorage.DownloadPayload) ([]byte, error)
	HTTPDownload(mux goahttp.Muxer, dec func(r *http.Request) goahttp.Decoder) http.HandlerFunc
}

type serviceImpl struct {
	logger logr.Logger
	db     *sqlx.DB
	tc     temporalsdk_client.Client
	config Config
	bucket *blob.Bucket
}

var _ Service = (*serviceImpl)(nil)

func NewService(logger logr.Logger, db *sql.DB, tc temporalsdk_client.Client, config Config) (*serviceImpl, error) {
	s := &serviceImpl{
		logger: logger,
		db:     sqlx.NewDb(db, "mysql"),
		tc:     tc,
		config: config,
	}

	var err error
	s.bucket, err = s.openBucket(&config)
	if err != nil {
		return nil, fmt.Errorf("error opening bucket: %v", err)
	}

	return s, nil
}

func (s *serviceImpl) openBucket(config *Config) (*blob.Bucket, error) {
	sessOpts := session.Options{}
	sessOpts.Config.WithRegion(s.config.Region)
	sessOpts.Config.WithEndpoint(s.config.Endpoint)
	sessOpts.Config.WithS3ForcePathStyle(s.config.PathStyle)
	sessOpts.Config.WithCredentials(
		credentials.NewStaticCredentials(
			s.config.Key, s.config.Secret, s.config.Token,
		),
	)
	sess, err := session.NewSessionWithOptions(sessOpts)
	if err != nil {
		return nil, err
	}
	return s3blob.OpenBucket(context.Background(), sess, s.config.Bucket, nil)
}

func (s *serviceImpl) Submit(ctx context.Context, payload *goastorage.SubmitPayload) (*goastorage.SubmitResult, error) {
	workflowReq := &StorageWorkflowRequest{AIPID: payload.AipID}
	_, err := InitStorageWorkflow(ctx, s.tc, workflowReq)
	if err != nil {
		return nil, goastorage.MakeNotAvailable(errors.New("cannot perform operation"))
	}

	p := Package{
		Name:      payload.Name,
		AIPID:     payload.AipID,
		Status:    StatusUnspecified,
		ObjectKey: uuid.New().String(),
	}
	err = s.createPackage(ctx, &p)
	if err != nil {
		return nil, goastorage.MakeNotValid(errors.New("cannot persist package"))
	}

	url, err := s.bucket.SignedURL(ctx, p.ObjectKey, &blob.SignedURLOptions{Expiry: submitURLExpirationTime, Method: http.MethodPut})
	if err != nil {
		return nil, goastorage.MakeNotValid(errors.New("cannot persist package"))
	}

	result := &goastorage.SubmitResult{
		URL: url,
	}
	return result, nil
}

func (s *serviceImpl) Update(ctx context.Context, payload *goastorage.UpdatePayload) (*goastorage.UpdateResult, error) {
	signal := UploadDoneSignal{}
	workflowID := fmt.Sprintf("%s-%s", StorageWorkflowName, payload.AipID)
	err := s.tc.SignalWorkflow(ctx, workflowID, "", UploadDoneSignalName, signal)
	if err != nil {
		return nil, goastorage.MakeNotAvailable(errors.New("cannot perform operation"))
	}
	// Uptade the package status to in_review
	err = s.updatePackageStatus(ctx, StatusInReview, payload.AipID)
	if err != nil {
		return nil, goastorage.MakeNotValid(errors.New("cannot persist package"))
	}

	result := &goastorage.UpdateResult{OK: true}
	return result, nil
}

func (s *serviceImpl) Download(ctx context.Context, payload *goastorage.DownloadPayload) ([]byte, error) {
	return []byte{}, nil
}

func (s *serviceImpl) HTTPDownload(mux goahttp.Muxer, dec func(r *http.Request) goahttp.Decoder) http.HandlerFunc {
	return func(rw http.ResponseWriter, req *http.Request) {
		// Decode request payload.
		payload, err := server.DecodeDownloadRequest(mux, dec)(req)
		if err != nil {
			rw.WriteHeader(http.StatusBadRequest)
			return
		}
		p := payload.(*goastorage.DownloadPayload)

		// Read storage package.
		ctx := context.Background()
		pkg, err := s.readPackage(ctx, p.AipID)
		if err != nil {
			rw.WriteHeader(http.StatusNotFound)
			return
		}

		// Get MinIO bucket reader for object key.
		reader, err := s.bucket.NewReader(ctx, pkg.ObjectKey, nil)
		if err != nil {
			rw.WriteHeader(http.StatusNotFound)
			return
		}
		defer reader.Close()

		rw.Header().Add("Content-Type", reader.ContentType())
		rw.Header().Add("Content-Length", strconv.FormatInt(reader.Size(), 10))

		// Copy reader contents into the response.
		_, err = io.Copy(rw, reader)
		if err != nil {
			rw.WriteHeader(http.StatusNotFound)
			return
		}
	}
}

func SetBucket(s *serviceImpl, b *blob.Bucket) {
	s.bucket = b
}

func (s *serviceImpl) createPackage(ctx context.Context, p *Package) error {
	query := `INSERT INTO storage_package (name, aip_id, status, object_key) VALUES (?, ?, ?, ?)`
	args := []interface{}{
		p.Name,
		p.AIPID,
		p.Status,
		p.ObjectKey,
	}

	res, err := s.db.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("error inserting package: %w", err)
	}

	var id int64
	if id, err = res.LastInsertId(); err != nil {
		return fmt.Errorf("error retrieving insert ID: %w", err)
	}

	p.ID = uint(id)

	return nil
}

func (s *serviceImpl) readPackage(ctx context.Context, AIPID string) (*Package, error) {
	query := "SELECT id, name, aip_id, status, object_key FROM storage_package WHERE aip_id = (?)"
	args := []interface{}{AIPID}
	p := Package{}

	if err := s.db.GetContext(ctx, &p, query, args...); err != nil {
		return nil, err
	}

	return &p, nil
}

func (s *serviceImpl) updatePackageStatus(ctx context.Context, status PackageStatus, aipID string) error {
	s.logger.Info("updating package status", "status", status, "aip_id", aipID)

	query := `UPDATE storage_package SET status=? WHERE aip_id=?`
	args := []interface{}{
		status,
		aipID,
	}

	_, err := s.db.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("error updating package status: %w", err)
	}

	return nil
}
