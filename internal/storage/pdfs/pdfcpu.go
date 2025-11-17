package pdfs

import (
	"io"

	pdfcpu_api "github.com/pdfcpu/pdfcpu/pkg/api"
	pdfcpu_model "github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"
)

type PDFCPU struct {
	cfg *pdfcpu_model.Configuration
}

// NewPDFCPU creates a new pdfcpu api wrapper using the default pdfcpu
// configuration. NewPDFCPU is not thread-safe because it reads and writes files
// in a shared directory (os.UserConfigDir + "/pdfcpu").
func NewPDFCPU() *PDFCPU {
	return &PDFCPU{cfg: pdfcpu_api.LoadConfiguration()}
}

func (p *PDFCPU) FillForm(src io.ReadSeeker, data io.Reader, dest io.Writer) error {
	c := *p.cfg // copy cfg to avoid data races from modifying shared config
	return pdfcpu_api.FillForm(src, data, dest, &c)
}

var _ FormFiller = (*PDFCPU)(nil) // ensure interface compliance
