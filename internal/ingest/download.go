package ingest

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/google/uuid"
	goahttp "goa.design/goa/v3/http"
	"gocloud.dev/gcerrors"

	"github.com/artefactual-sdps/enduro/internal/api/gen/http/ingest/server"
	goaingest "github.com/artefactual-sdps/enduro/internal/api/gen/ingest"
	"github.com/artefactual-sdps/enduro/internal/persistence"
)

type errorResponse struct {
	Message string `json:"message"`
}

func writeJSONError(rw http.ResponseWriter, code int, msg string) {
	rw.Header().Set("Content-Type", "application/json")
	rw.WriteHeader(code)
	_ = json.NewEncoder(rw).Encode(errorResponse{Message: msg})
}

// DownloadSIP returns an HTTP handler that sets the headers for the SIP when it is downloaded.
// The headers are the Content Type, Content Length, and the Content Disposition.
// If there is an error with the file download, it will return http-status not found (404).
func (svc *ingestImpl) DownloadSIP(mux goahttp.Muxer, dec func(r *http.Request) goahttp.Decoder) http.HandlerFunc {
	return func(rw http.ResponseWriter, req *http.Request) {
		// Decode request payload.
		payload, err := server.DecodeDownloadSipRequest(mux, dec)(req)
		if err != nil {
			writeJSONError(rw, http.StatusBadRequest, "invalid request")
			return
		}
		p := payload.(*goaingest.DownloadSipPayload)

		sipUUID, err := uuid.Parse(p.UUID)
		if err != nil {
			writeJSONError(rw, http.StatusBadRequest, "invalid UUID format")
			return
		}

		// Read the persisted SIP.
		ctx := req.Context()
		sip, err := svc.perSvc.ReadSIP(ctx, sipUUID)
		if err != nil {
			if errors.Is(err, persistence.ErrNotFound) {
				writeJSONError(rw, http.StatusNotFound, "SIP not found")
			} else {
				writeJSONError(rw, http.StatusInternalServerError, "error reading SIP")
			}
			return
		}

		// Check failed as and failed key are set.
		if sip.FailedAs == "" || sip.FailedKey == "" {
			writeJSONError(rw, http.StatusBadRequest, "SIP has no failed values")
			return
		}

		// Get reader from internal storage for failed key.
		reader, err := svc.internalStorage.NewReader(ctx, sip.FailedKey, nil)
		if err != nil {
			if gcerrors.Code(err) == gcerrors.NotFound {
				writeJSONError(rw, http.StatusNotFound, "SIP file not found")
			} else {
				writeJSONError(rw, http.StatusInternalServerError, "error reading SIP file")
			}
			return
		}
		defer reader.Close()

		rw.Header().Add("Content-Type", reader.ContentType())
		rw.Header().Add("Content-Length", strconv.FormatInt(reader.Size(), 10))
		rw.Header().Add("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", sip.FailedKey))

		// Copy reader contents into the response.
		_, err = io.Copy(rw, reader)
		if err != nil {
			writeJSONError(rw, http.StatusInternalServerError, "error streaming SIP file")
			return
		}
	}
}
