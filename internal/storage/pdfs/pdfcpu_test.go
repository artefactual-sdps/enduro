package pdfs_test

import (
	"bytes"
	"os"
	"testing"

	"gotest.tools/v3/assert"

	"github.com/artefactual-sdps/enduro/internal/storage/pdfs"
)

func TestPDFCPU_FillForm(t *testing.T) {
	t.Parallel()

	templatePath := "../../../assets/Enduro_AIP_deletion_report_v3.tmpl.pdf"
	buf := &bytes.Buffer{}
	jsonData := `{
	"forms": [
		{
			"textfield": [
				{
					"name": "deleted_at",
					"value": "2025-10-27T08:20:43Z"
				}
			]
		}
	]
}
`

	src, err := os.Open(templatePath)
	if err != nil {
		t.Fatalf("couldn't open PDF template: %v", err)
	}
	defer src.Close()

	ff := pdfs.NewPDFCPU()
	err = ff.FillForm(src, bytes.NewReader([]byte(jsonData)), buf)
	assert.NilError(t, err)
	assert.Assert(t, buf.Len() > 0)
}
