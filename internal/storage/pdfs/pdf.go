package pdfs

import "io"

// FormFiller defines the interface for filling PDF forms.
type FormFiller interface {
	// FillForm fills the src PDF form fields with JSON formatted data and
	// writes the resulting filled PDF to dest.
	FillForm(src io.ReadSeeker, data io.Reader, dest io.Writer) error
}
