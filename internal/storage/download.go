package storage

import (
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/google/uuid"
	goahttp "goa.design/goa/v3/http"

	"github.com/artefactual-sdps/enduro/internal/api/gen/http/storage/server"
	goastorage "github.com/artefactual-sdps/enduro/internal/api/gen/storage"
	"github.com/artefactual-sdps/enduro/internal/fsutil"
)

// Download returns an HTTP handler that sets the headers for the AIP when it is downloaded.
// The headers are the Content Type, Content Length, and the Content Disposition.
// If there is an error with the file download, it will return http-status not found (404).
func Download(svc Service, mux goahttp.Muxer, dec func(r *http.Request) goahttp.Decoder) http.HandlerFunc {
	return func(rw http.ResponseWriter, req *http.Request) {
		// Decode request payload.
		payload, err := server.DecodeDownloadAipRequest(mux, dec)(req)
		if err != nil {
			rw.WriteHeader(http.StatusBadRequest)
			return
		}
		p := payload.(*goastorage.DownloadAipPayload)

		aipID, err := uuid.Parse(p.UUID)
		if err != nil {
			rw.WriteHeader(http.StatusBadRequest)
			return
		}

		// Read AIP.
		ctx := req.Context()
		aip, err := svc.ReadAip(ctx, aipID)
		if err != nil {
			rw.WriteHeader(http.StatusNotFound)
			return
		}

		// Get MinIO bucket reader for object key.
		reader, err := svc.AipReader(ctx, aip)
		if err != nil {
			rw.WriteHeader(http.StatusNotFound)
			return
		}
		defer reader.Close()

		filename := fmt.Sprintf("%s-%s.7z", fsutil.BaseNoExt(aip.Name), aip.UUID)

		rw.Header().Add("Content-Type", reader.ContentType())
		rw.Header().Add("Content-Length", strconv.FormatInt(reader.Size(), 10))
		rw.Header().Add("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))

		// Copy reader contents into the response.
		_, err = io.Copy(rw, reader)
		if err != nil {
			rw.WriteHeader(http.StatusNotFound)
			return
		}
	}
}
