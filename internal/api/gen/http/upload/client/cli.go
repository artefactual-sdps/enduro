// Code generated by goa v3.15.2, DO NOT EDIT.
//
// upload HTTP client CLI support package
//
// Command:
// $ goa gen github.com/artefactual-sdps/enduro/internal/api/design -o
// internal/api

package client

import (
	upload "github.com/artefactual-sdps/enduro/internal/api/gen/upload"
	goa "goa.design/goa/v3/pkg"
)

// BuildUploadPayload builds the payload for the upload upload endpoint from
// CLI flags.
func BuildUploadPayload(uploadUploadContentType string, uploadUploadOauthToken string) (*upload.UploadPayload, error) {
	var err error
	var contentType string
	{
		if uploadUploadContentType != "" {
			contentType = uploadUploadContentType
			err = goa.MergeErrors(err, goa.ValidatePattern("content_type", contentType, "multipart/[^;]+; boundary=.+"))
			if err != nil {
				return nil, err
			}
		}
	}
	var oauthToken *string
	{
		if uploadUploadOauthToken != "" {
			oauthToken = &uploadUploadOauthToken
		}
	}
	v := &upload.UploadPayload{}
	v.ContentType = contentType
	v.OauthToken = oauthToken

	return v, nil
}
