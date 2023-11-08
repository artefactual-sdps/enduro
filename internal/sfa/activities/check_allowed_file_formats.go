package activities

import (
	"context"
	"fmt"

	"github.com/artefactual-sdps/enduro/internal/sfa/fformat"
	"github.com/artefactual-sdps/enduro/internal/sfa/sip"
)

const AllowedFileFormatsName = "allowed-file-formats"

type AllowedFileFormatsActivity struct{}

func NewAllowedFileFormatsActivity() *AllowedFileFormatsActivity {
	return &AllowedFileFormatsActivity{}
}

type AllowedFileFormatsParams struct {
	SipPath string
}

type AllowedFileFormatsResult struct {
	Ok         bool
	Formats    map[string]string
	NotAllowed []string
}

func (md *AllowedFileFormatsActivity) Execute(ctx context.Context, params *AllowedFileFormatsParams) (*AllowedFileFormatsResult, error) {
	res := &AllowedFileFormatsResult{}
	sf := fformat.NewSiegfriedEmbed()
	// TODO(daniel): make allowed list configurable.
	allowed := map[string]struct{}{
		"fmt/95":    {},
		"x-fmt/16":  {},
		"x-fmt/21":  {},
		"x-fmt/22":  {},
		"x-fmt/62":  {},
		"x-fmt/111": {},
		"x-fmt/282": {},
		"x-fmt/283": {},
		"fmt/354":   {},
		"fmt/476":   {},
		"fmt/477":   {},
		"fmt/478":   {},
		"x-fmt/18":  {},
		"fmt/161":   {},
		"fmt/1196":  {},
		"fmt/1777":  {},
		"fmt/353":   {},
		"x-fmt/392": {},
		"fmt/1":     {},
		"fmt/2":     {},
		"fmt/6":     {},
		"fmt/141":   {},
		"fmt/569":   {},
		"fmt/199":   {},
		"fmt/101":   {},
		"x-fmt/280": {},
		"fmt/1014":  {},
		"fmt/1012":  {},
		"fmt/654":   {},
		"fmt/1013":  {},
		"fmt/1011":  {},
		"fmt/653":   {},
	}

	s, err := sip.NewSFASip(params.SipPath)
	if err != nil {
		return nil, err
	}

	res.Formats = make(map[string]string)
	for _, path := range s.Files {
		ff, err := sf.Identify(path)
		if err != nil {
			return nil, err
		}
		res.Formats[path] = ff.ID
	}

	for path, formatID := range res.Formats {
		if _, exists := allowed[formatID]; !exists {
			msg := fmt.Sprintf("File format not allowed: %s", path)
			res.NotAllowed = append(res.NotAllowed, msg)
		}
	}

	if len(res.NotAllowed) == 0 {
		res.Ok = true
	}
	return res, nil
}
