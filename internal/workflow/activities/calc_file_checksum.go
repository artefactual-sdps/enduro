package activities

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"

	"github.com/artefactual-sdps/enduro/internal/datatypes"
)

const CalcFileChecksumActivityName = "calc-file-checksum-activity"

type (
	CalcFileChecksumActivityParams struct {
		Path string
	}
	CalcFileChecksumActivityResult struct {
		Checksum datatypes.Checksum
	}
	CalcFileChecksumActivity struct{}
)

func NewCalcFileChecksumActivity() *CalcFileChecksumActivity {
	return &CalcFileChecksumActivity{}
}

// Execute calculates the SHA-256 checksum of the file at params.Path.
//
// If params.Path is a directory, Execute returns an error.
func (a *CalcFileChecksumActivity) Execute(
	ctx context.Context,
	params *CalcFileChecksumActivityParams,
) (*CalcFileChecksumActivityResult, error) {
	fi, err := os.Stat(params.Path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("calculate file checksum: file not found: %s", params.Path)
		}
		return nil, fmt.Errorf("calculate file checksum: stat file: %v", err)
	}

	if fi.IsDir() {
		return nil, fmt.Errorf("calculate file checksum: not a file: %s", params.Path)
	}

	f, err := os.Open(params.Path)
	if err != nil {
		return nil, fmt.Errorf("calculate file checksum: open file: %v", err)
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return nil, fmt.Errorf("calculate file checksum: compute hash: %v", err)
	}

	return &CalcFileChecksumActivityResult{
		Checksum: datatypes.Checksum{
			Algorithm: datatypes.ChecksumAlgoSHA256,
			Hash:      hex.EncodeToString(h.Sum(nil)),
		},
	}, nil
}
