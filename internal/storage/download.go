package storage

import (
	"context"
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
func Download(svc Service, mux goahttp.Muxer, preservationTaskQueue string, dec func(r *http.Request) goahttp.Decoder) http.HandlerFunc {
	return func(rw http.ResponseWriter, req *http.Request) {
		// Decode request payload.
		payload, err := server.DecodeDownloadRequest(mux, dec)(req)
		if err != nil {
			rw.WriteHeader(http.StatusBadRequest)
			return
		}
		p := payload.(*goastorage.DownloadPayload)

		aipID, err := uuid.Parse(p.AipID)
		if err != nil {
			rw.WriteHeader(http.StatusBadRequest)
			return
		}

		// Read storage package.
		ctx := context.Background()
		pkg, err := svc.ReadPackage(ctx, aipID)
		if err != nil {
			rw.WriteHeader(http.StatusNotFound)
			return
		}

		// Download aip from the archivematica storage service if preservationTaskQueue field is set to am
		if preservationTaskQueue == "am" {
			panic("Not Implemented")
		}
		// Get MinIO bucket reader for object key.
		reader, err := svc.PackageReader(ctx, pkg)
		if err != nil {
			rw.WriteHeader(http.StatusNotFound)
			return
		}
		defer reader.Close()

		filename := fmt.Sprintf("%s-%s.7z", fsutil.BaseNoExt(pkg.Name), pkg.AipID)

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
