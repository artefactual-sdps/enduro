package about

import (
	"context"
	"errors"

	"github.com/go-logr/logr"
	"goa.design/goa/v3/security"

	"github.com/artefactual-sdps/enduro/internal/api/auth"
	goaabout "github.com/artefactual-sdps/enduro/internal/api/gen/about"
	"github.com/artefactual-sdps/enduro/internal/childwf"
	"github.com/artefactual-sdps/enduro/internal/ingest"
	"github.com/artefactual-sdps/enduro/internal/temporal"
	"github.com/artefactual-sdps/enduro/internal/version"
)

type Service struct {
	logger        logr.Logger
	presTaskQueue string
	cwfConfigs    childwf.Configs
	uploadConfig  ingest.UploadConfig
	tokenVerifier auth.TokenVerifier
}

var _ goaabout.Service = (*Service)(nil)

var ErrUnauthorized error = goaabout.Unauthorized("Unauthorized")

func NewService(
	logger logr.Logger,
	presTaskQueue string,
	cwfConfigs childwf.Configs,
	uploadConfig ingest.UploadConfig,
	tokenVerifier auth.TokenVerifier,
) *Service {
	return &Service{
		logger:        logger,
		presTaskQueue: presTaskQueue,
		cwfConfigs:    cwfConfigs,
		uploadConfig:  uploadConfig,
		tokenVerifier: tokenVerifier,
	}
}

func (s *Service) JWTAuth(ctx context.Context, token string, scheme *security.JWTScheme) (context.Context, error) {
	claims, err := s.tokenVerifier.Verify(ctx, token)
	if err != nil {
		if !errors.Is(err, auth.ErrUnauthorized) {
			s.logger.V(1).Info("failed to verify token", "err", err)
		}
		return ctx, ErrUnauthorized
	}

	ctx = auth.WithUserClaims(ctx, claims)

	return ctx, nil
}

func (s *Service) About(context.Context, *goaabout.AboutPayload) (*goaabout.EnduroAbout, error) {
	res := &goaabout.EnduroAbout{
		Version:       version.Short,
		UploadMaxSize: s.uploadConfig.MaxSize,
	}

	res.PreservationSystem = "Unknown"
	switch s.presTaskQueue {
	case temporal.AmWorkerTaskQueue:
		res.PreservationSystem = "Archivematica"
	case temporal.A3mWorkerTaskQueue:
		res.PreservationSystem = "a3m"
	}

	// Add child workflows to the response.
	if len(s.cwfConfigs) > 0 {
		res.ChildWorkflows = make([]*goaabout.EnduroChildworkflow, 0, len(s.cwfConfigs))
		for _, cfg := range s.cwfConfigs {
			res.ChildWorkflows = append(res.ChildWorkflows, &goaabout.EnduroChildworkflow{
				Type:         cfg.Type.String(),
				TaskQueue:    cfg.TaskQueue,
				WorkflowName: cfg.WorkflowName,
			})
		}

	}

	return res, nil
}
