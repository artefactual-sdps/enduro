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
)

// Download returns an HTTP handler that se
func Download(svc Service, mux goahttp.Muxer, dec func(r *http.Request) goahttp.Decoder) http.HandlerFunc {
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

		// Get MinIO bucket reader for object key.
		reader, err := svc.PackageReader(ctx, pkg)
		if err != nil {
			rw.WriteHeader(http.StatusNotFound)
			return
		}
		defer reader.Close()

		filename := fmt.Sprintf("enduro-%s.7z", pkg.AipID)

		rw.Header().Add("Content-Type", reader.ContentType())
		rw.Header().Add("Content-Length", strconv.FormatInt(reader.Size(), 10))
		rw.Header().Add("Content-Disposition",
			fmt.Sprintf("attachment; filename=\"%s\"", filename))

		// Copy reader contents into the response.
		_, err = io.Copy(rw, reader)
		if err != nil {
			rw.WriteHeader(http.StatusNotFound)
			return
		}
	}
}
