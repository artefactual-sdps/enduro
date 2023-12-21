// Code generated by goa v3.14.1, DO NOT EDIT.
//
// upload endpoints
//
// Command:
// $ goa gen github.com/artefactual-sdps/enduro/internal/api/design -o
// internal/api

package upload

import (
	"context"
	"io"

	goa "goa.design/goa/v3/pkg"
	"goa.design/goa/v3/security"
)

// Endpoints wraps the "upload" service endpoints.
type Endpoints struct {
	Upload goa.Endpoint
}

// UploadRequestData holds both the payload and the HTTP request body reader of
// the "upload" method.
type UploadRequestData struct {
	// Payload is the method payload.
	Payload *UploadPayload
	// Body streams the HTTP request body.
	Body io.ReadCloser
}

// NewEndpoints wraps the methods of the "upload" service with endpoints.
func NewEndpoints(s Service) *Endpoints {
	// Casting service to Auther interface
	a := s.(Auther)
	return &Endpoints{
		Upload: NewUploadEndpoint(s, a.OAuth2Auth),
	}
}

// Use applies the given middleware to all the "upload" service endpoints.
func (e *Endpoints) Use(m func(goa.Endpoint) goa.Endpoint) {
	e.Upload = m(e.Upload)
}

// NewUploadEndpoint returns an endpoint function that calls the method
// "upload" of service "upload".
func NewUploadEndpoint(s Service, authOAuth2Fn security.AuthOAuth2Func) goa.Endpoint {
	return func(ctx context.Context, req any) (any, error) {
		ep := req.(*UploadRequestData)
		var err error
		sc := security.OAuth2Scheme{
			Name:           "oauth2",
			Scopes:         []string{},
			RequiredScopes: []string{},
			Flows: []*security.OAuthFlow{
				&security.OAuthFlow{
					Type:       "client_credentials",
					TokenURL:   "/oauth2/token",
					RefreshURL: "/oauth2/refresh",
				},
			},
		}
		var token string
		if ep.Payload.OauthToken != nil {
			token = *ep.Payload.OauthToken
		}
		ctx, err = authOAuth2Fn(ctx, token, &sc)
		if err != nil {
			return nil, err
		}
		return nil, s.Upload(ctx, ep.Payload, ep.Body)
	}
}
