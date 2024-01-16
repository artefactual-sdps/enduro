package storage

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/google/uuid"
	goahttp "goa.design/goa/v3/http"
	"gocloud.dev/blob"

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
		// Package reader returns an external processes' return (such as an AIP from archivematica) given by a particular package
		// we can encapsulate the return into something else later besides an anytype, we might even do a type switch depending on
		// configuration.
		reader, resp, err := svc.PackageReader(ctx, pkg)
		_ = resp
		if err != nil {
			rw.WriteHeader(http.StatusNotFound)
			return
		}
		defer reader.Close()

		filename := fmt.Sprintf("%s-%s.7z", fsutil.BaseNoExt(pkg.Name), pkg.AipID)
		if reader == nil {
			fillResponseResp(rw, resp, filename)
			return
		}

		fillResponseBlob(rw, reader, filename)
	}
}

// Fill the header and body of the http response with the blob's response information.
func fillResponseBlob(rw http.ResponseWriter, responseInfo *blob.Reader, filename string) {
	rw.Header().Add("Content-Type", responseInfo.ContentType())
	rw.Header().Add("Content-Length", strconv.FormatInt(responseInfo.Size(), 10))
	rw.Header().Add("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))

	// Copy reader contents into the response.
	_, err := io.Copy(rw, responseInfo)
	if err != nil {
		rw.WriteHeader(http.StatusNotFound)
		return
	}
}

// Fill the header and body of the http response with the resps's response information.
func fillResponseResp(rw http.ResponseWriter, responseInfo *http.Response, filename string) {
	rw.Header().Add("Content-Type", responseInfo.Header.Get("Content-Type"))
	rw.Header().Add("Content-Length", responseInfo.Header.Get("Content-Length"))
	rw.Header().Add("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))

	// Copy reader contents into the response.
	_, err := io.Copy(rw, responseInfo.Body)
	if err != nil {
		rw.WriteHeader(http.StatusNotFound)
		return
	}
}
