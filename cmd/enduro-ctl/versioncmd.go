package main

import (
	"context"
	"flag"
	"fmt"
	"io"

	"github.com/peterbourgon/ff/v3/ffcli"

	"github.com/artefactual-sdps/enduro/internal/rootcmd"
	"github.com/artefactual-sdps/enduro/internal/version"
)

func versionCommand(rootConfig *rootcmd.Config, out io.Writer) *ffcli.Command {
	fs := flag.NewFlagSet("enduro-ctl version", flag.ExitOnError)
	rootConfig.RegisterFlags(fs)

	return &ffcli.Command{
		Name:       "version",
		ShortUsage: "enduro-ctl version",
		ShortHelp:  "Print version.",
		FlagSet:    fs,
		Exec: func(ctx context.Context, args []string) error {
			v := version.Short
			if rootConfig.Verbose {
				v = version.Long
			}

			fmt.Fprintf(out, "%s\n", v)

			return nil
		},
	}
}
