package activities

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
)

const CalcFileChecksumActivityName = "calc-file-checksum-activity"

type (
	CalcFileChecksumActivityParams struct {
		Path string
	}
	CalcFileChecksumActivityResult struct {
		Algo string
		Hash string
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
	f, err := os.Open(params.Path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("calculate file checksum: file not found: %s", params.Path)
		}
		return nil, fmt.Errorf("calculate file checksum: open file: %v", err)
	}
	defer f.Close()

	if fi, err := f.Stat(); err != nil {
		return nil, fmt.Errorf("calculate file checksum: stat file: %v", err)
	} else if fi.IsDir() {
		return nil, fmt.Errorf("calculate file checksum: not a file: %s", params.Path)
	}

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return nil, fmt.Errorf("calculate file checksum: compute hash: %v", err)
	}

	return &CalcFileChecksumActivityResult{
		Algo: "SHA-256",
		Hash: hex.EncodeToString(h.Sum(nil)),
	}, nil
}
