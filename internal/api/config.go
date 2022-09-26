package api

import (
	"github.com/artefactual-sdps/enduro/internal/api/auth"
)

type Config struct {
	Listen     string
	Debug      bool
	AppVersion string
	Auth       auth.Config
}
