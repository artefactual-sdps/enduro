package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/go-logr/logr"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	temporalapi_enums "go.temporal.io/api/enums/v1"
	temporalapi_serviceerror "go.temporal.io/api/serviceerror"
	temporalsdk_client "go.temporal.io/sdk/client"
	goahttp "goa.design/goa/v3/http"
	"gocloud.dev/blob"
	"gocloud.dev/blob/s3blob"

	"github.com/artefactual-labs/enduro/internal/api/gen/http/storage/server"
	goastorage "github.com/artefactual-labs/enduro/internal/api/gen/storage"
)

var SubmitURLExpirationTime = 15 * time.Minute

type Location interface {
	Name() string
	OpenBucket() (*blob.Bucket, error)
}

type locationImpl struct {
	name   string
	config LocationConfig
}

func (l *locationImpl) Name() string {
	return l.name
}

func (l *locationImpl) OpenBucket() (*blob.Bucket, error) {
	sessOpts := session.Options{}
	sessOpts.Config.WithRegion(l.config.Region)
	sessOpts.Config.WithEndpoint(l.config.Endpoint)
	sessOpts.Config.WithS3ForcePathStyle(l.config.PathStyle)
	sessOpts.Config.WithCredentials(
		credentials.NewStaticCredentials(
			l.config.Key, l.config.Secret, l.config.Token,
		),
	)
	sess, err := session.NewSessionWithOptions(sessOpts)
	if err != nil {
		return nil, err
	}
	return s3blob.OpenBucket(context.Background(), sess, l.config.Bucket, nil)
}

func NewLocation(config LocationConfig) Location {
	return &locationImpl{
		name:   config.Name,
		config: config,
	}
}

type Service interface {
	Submit(context.Context, *goastorage.SubmitPayload) (res *goastorage.SubmitResult, err error)
	Update(context.Context, *goastorage.UpdatePayload) (err error)
	Download(context.Context, *goastorage.DownloadPayload) ([]byte, error)
	List(context.Context) (res goastorage.StoredLocationCollection, err error)
	Move(context.Context, *goastorage.MovePayload) (err error)
	MoveStatus(context.Context, *goastorage.MoveStatusPayload) (res *goastorage.MoveStatusResult, err error)

	Bucket() *blob.Bucket
	Location(name string) (Location, error)
	ReadPackage(ctx context.Context, AIPID string) (*Package, error)
	UpdatePackageStatus(ctx context.Context, status PackageStatus, aipID string) error
	UpdatePackageLocation(ctx context.Context, location string, aipID string) error

	HTTPDownload(mux goahttp.Muxer, dec func(r *http.Request) goahttp.Decoder) http.HandlerFunc
}

type serviceImpl struct {
	logger    logr.Logger
	db        *sqlx.DB
	tc        temporalsdk_client.Client
	config    Config
	bucket    *blob.Bucket
	mu        sync.RWMutex
	locations map[string]Location
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

	locations := map[string]Location{}
	for _, item := range config.Locations {
		l := NewLocation(item)
		locations[item.Name] = l
	}
	s.locations = locations

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

func (s *serviceImpl) Bucket() *blob.Bucket {
	return s.bucket
}

func (s *serviceImpl) Location(name string) (Location, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	l, ok := s.locations[name]
	if !ok {
		return nil, fmt.Errorf("error loading location: unknown location %s", name)
	}

	return l, nil
}

func (s *serviceImpl) Submit(ctx context.Context, payload *goastorage.SubmitPayload) (*goastorage.SubmitResult, error) {
	_, err := InitStorageWorkflow(ctx, s.tc, &StorageUploadWorkflowRequest{AIPID: payload.AipID})
	if err != nil {
		return nil, goastorage.MakeNotAvailable(errors.New("cannot perform operation"))
	}

	p := Package{
		Name:      payload.Name,
		AIPID:     payload.AipID,
		Status:    StatusUnspecified,
		ObjectKey: uuid.New().String(),
		Location:  "",
	}
	err = s.createPackage(ctx, &p)
	if err != nil {
		return nil, goastorage.MakeNotValid(errors.New("cannot persist package"))
	}

	url, err := s.bucket.SignedURL(ctx, p.ObjectKey, &blob.SignedURLOptions{Expiry: SubmitURLExpirationTime, Method: http.MethodPut})
	if err != nil {
		return nil, goastorage.MakeNotValid(errors.New("cannot persist package"))
	}

	result := &goastorage.SubmitResult{
		URL: url,
	}
	return result, nil
}

func (s *serviceImpl) Update(ctx context.Context, payload *goastorage.UpdatePayload) error {
	signal := UploadDoneSignal{}
	workflowID := fmt.Sprintf("%s-%s", StorageUploadWorkflowName, payload.AipID)
	err := s.tc.SignalWorkflow(ctx, workflowID, "", UploadDoneSignalName, signal)
	if err != nil {
		return goastorage.MakeNotAvailable(errors.New("cannot perform operation"))
	}
	// Uptade the package status to in_review
	err = s.UpdatePackageStatus(ctx, StatusInReview, payload.AipID)
	if err != nil {
		return goastorage.MakeNotValid(errors.New("cannot persist package"))
	}

	return nil
}

func (s *serviceImpl) Download(ctx context.Context, payload *goastorage.DownloadPayload) ([]byte, error) {
	return []byte{}, nil
}

func (s *serviceImpl) List(context.Context) (goastorage.StoredLocationCollection, error) {
	res := []*goastorage.StoredLocation{}
	for _, item := range s.locations {
		l := &goastorage.StoredLocation{
			ID:   item.Name(),
			Name: item.Name(),
		}
		res = append(res, l)
	}
	return res, nil
}

func (s *serviceImpl) Move(ctx context.Context, payload *goastorage.MovePayload) error {
	p, err := s.ReadPackage(ctx, payload.AipID)
	if err == sql.ErrNoRows {
		return &goastorage.StoragePackageNotfound{AipID: payload.AipID, Message: "not_found"}
	} else if err != nil {
		return err
	}

	_, err = InitStorageMoveWorkflow(ctx, s.tc, &StorageMoveWorkflowRequest{
		AIPID:    p.AIPID,
		Location: payload.Location,
	})
	if err != nil {
		s.logger.Error(err, "error initializing move workflow")
		return goastorage.MakeNotAvailable(errors.New("cannot perform operation"))
	}

	return nil
}

func (s *serviceImpl) MoveStatus(ctx context.Context, payload *goastorage.MoveStatusPayload) (*goastorage.MoveStatusResult, error) {
	p, err := s.ReadPackage(ctx, payload.AipID)
	if err == sql.ErrNoRows {
		return nil, &goastorage.StoragePackageNotfound{AipID: payload.AipID, Message: "not_found"}
	} else if err != nil {
		return nil, err
	}

	resp, err := s.tc.DescribeWorkflowExecution(ctx, fmt.Sprintf("%s-%s", StorageMoveWorkflowName, p.AIPID), "")
	if err != nil {
		switch err := err.(type) {
		case *temporalapi_serviceerror.NotFound:
			s.logger.Error(err, "error retrieving workflow")
			// XXX this should be 404, can we create goastorage.MakeNotFound?
			return nil, goastorage.MakeNotAvailable(errors.New("cannot perform operation"))
		default:
			// XXX should this be 404 too?
			s.logger.Error(err, "error retrieving workflow")
			return nil, goastorage.MakeNotAvailable(errors.New("cannot perform operation"))
		}
	}
	if resp.WorkflowExecutionInfo == nil {
		// XXX how to log error when there's no error?
		s.logger.Error(errors.New("error"), "error retrieving workflow execution details")
		return nil, goastorage.MakeNotAvailable(errors.New("cannot perform operation"))
	}

	// XXX: what about WORKFLOW_EXECUTION_STATUS_UNSPECIFIED and WORKFLOW_EXECUTION_STATUS_CONTINUED_AS_NEW?
	var done bool
	switch resp.WorkflowExecutionInfo.Status {
	case
		temporalapi_enums.WORKFLOW_EXECUTION_STATUS_FAILED,
		temporalapi_enums.WORKFLOW_EXECUTION_STATUS_CANCELED,
		temporalapi_enums.WORKFLOW_EXECUTION_STATUS_TERMINATED,
		temporalapi_enums.WORKFLOW_EXECUTION_STATUS_TIMED_OUT:
		// XXX how to log error when there's no error?
		s.logger.Error(errors.New("error"), "workflow execution failed")
		return nil, goastorage.MakeNotAvailable(errors.New("cannot perform operation"))
	case temporalapi_enums.WORKFLOW_EXECUTION_STATUS_COMPLETED:
		done = true
	case temporalapi_enums.WORKFLOW_EXECUTION_STATUS_RUNNING:
		done = false
	}

	return &goastorage.MoveStatusResult{Done: done}, nil
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
		pkg, err := s.ReadPackage(ctx, p.AipID)
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

		filename := fmt.Sprintf("enduro-%s.7z", pkg.AIPID)

		rw.Header().Add("Content-Type", reader.ContentType())
		rw.Header().Add("Content-Length", strconv.FormatInt(reader.Size(), 10))
		rw.Header().Add("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))

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
	query := `INSERT INTO storage_package (name, aip_id, status, object_key, location) VALUES (?, ?, ?, ?, ?)`
	args := []interface{}{
		p.Name,
		p.AIPID,
		p.Status,
		p.ObjectKey,
		p.Location,
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

func (s *serviceImpl) ReadPackage(ctx context.Context, AIPID string) (*Package, error) {
	query := "SELECT id, name, aip_id, status, object_key, location FROM storage_package WHERE aip_id = (?)"
	args := []interface{}{AIPID}
	p := Package{}

	if err := s.db.GetContext(ctx, &p, query, args...); err != nil {
		return nil, err
	}

	return &p, nil
}

func (s *serviceImpl) UpdatePackageStatus(ctx context.Context, status PackageStatus, aipID string) error {
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

func (s *serviceImpl) UpdatePackageLocation(ctx context.Context, location string, aipID string) error {
	query := `UPDATE storage_package SET location=? WHERE aip_id=?`
	args := []interface{}{
		location,
		aipID,
	}

	_, err := s.db.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("error updating package location: %w", err)
	}

	return nil
}
