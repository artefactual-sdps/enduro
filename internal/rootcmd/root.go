package rootcmd

import (
	"context"
	"flag"

	"github.com/artefactual-sdps/enduro/internal/config"
	"github.com/peterbourgon/ff/v3/ffcli"
)

type Config struct {
	config.Configuration
	Verbose bool
}

func New() (*ffcli.Command, *Config) {
	var cfg Config

	fs := flag.NewFlagSet("enduro-ctl", flag.ExitOnError)
	cfg.RegisterFlags(fs)

	return &ffcli.Command{
		Name:       "enduro-ctl",
		ShortUsage: "enduro-ctl [flags] <subcommand> [flags] [<arg>...]",
		FlagSet:    fs,
		Exec:       cfg.Exec,
	}, &cfg
}

func (c *Config) RegisterFlags(fs *flag.FlagSet) {
	fs.BoolVar(&c.Verbose, "v", false, "Log verbose output")
	fs.BoolVar(&c.Debug, "debug", false, "Enable debug mode")
}

func (c *Config) Exec(context.Context, []string) error {
	return flag.ErrHelp
}
