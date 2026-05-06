package config

import "github.com/artefactual-sdps/enduro/internal/api/auth"

type APIConfig struct {
	Listen     string
	Debug      bool
	Auth       auth.Config
	CORSOrigin string
}
