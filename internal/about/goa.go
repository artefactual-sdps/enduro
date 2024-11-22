package about

import (
	"context"
	"errors"

	"github.com/go-logr/logr"
	"goa.design/goa/v3/security"

	"github.com/artefactual-sdps/enduro/internal/api/auth"
	goaabout "github.com/artefactual-sdps/enduro/internal/api/gen/about"
	"github.com/artefactual-sdps/enduro/internal/poststorage"
	"github.com/artefactual-sdps/enduro/internal/preprocessing"
	"github.com/artefactual-sdps/enduro/internal/temporal"
	"github.com/artefactual-sdps/enduro/internal/version"
)

type Service struct {
	logger        logr.Logger
	presTaskQueue string
	ppConfig      preprocessing.Config
	psConfig      []poststorage.Config
	tokenVerifier auth.TokenVerifier
}

var _ goaabout.Service = (*Service)(nil)

var ErrUnauthorized error = goaabout.Unauthorized("Unauthorized")

func NewService(
	logger logr.Logger,
	presTaskQueue string,
	ppConfig preprocessing.Config,
	psConfig []poststorage.Config,
	tokenVerifier auth.TokenVerifier,
) *Service {
	return &Service{
		logger:        logger,
		presTaskQueue: presTaskQueue,
		ppConfig:      ppConfig,
		psConfig:      psConfig,
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
		Version: version.Short,
		Preprocessing: &goaabout.EnduroPreprocessing{
			Enabled:      s.ppConfig.Enabled,
			WorkflowName: s.ppConfig.Temporal.WorkflowName,
			TaskQueue:    s.ppConfig.Temporal.TaskQueue,
		},
	}

	res.PreservationSystem = "Unknown"
	if s.presTaskQueue == temporal.AmWorkerTaskQueue {
		res.PreservationSystem = "Archivematica"
	} else if s.presTaskQueue == temporal.A3mWorkerTaskQueue {
		res.PreservationSystem = "a3m"
	}

	res.Poststorage = make([]*goaabout.EnduroPoststorage, len(s.psConfig))
	for i, cfg := range s.psConfig {
		res.Poststorage[i] = &goaabout.EnduroPoststorage{
			WorkflowName: cfg.WorkflowName,
			TaskQueue:    cfg.TaskQueue,
		}
	}

	return res, nil
}
