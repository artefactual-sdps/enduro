// Code generated by goa v3.15.1, DO NOT EDIT.
//
// package endpoints
//
// Command:
// $ goa gen github.com/artefactual-sdps/enduro/internal/api/design -o
// internal/api

package package_

import (
	"context"

	goa "goa.design/goa/v3/pkg"
	"goa.design/goa/v3/security"
)

// Endpoints wraps the "package" service endpoints.
type Endpoints struct {
	MonitorRequest      goa.Endpoint
	Monitor             goa.Endpoint
	List                goa.Endpoint
	Show                goa.Endpoint
	PreservationActions goa.Endpoint
	Confirm             goa.Endpoint
	Reject              goa.Endpoint
	Move                goa.Endpoint
	MoveStatus          goa.Endpoint
}

// MonitorEndpointInput holds both the payload and the server stream of the
// "monitor" method.
type MonitorEndpointInput struct {
	// Payload is the method payload.
	Payload *MonitorPayload
	// Stream is the server stream used by the "monitor" method to send data.
	Stream MonitorServerStream
}

// NewEndpoints wraps the methods of the "package" service with endpoints.
func NewEndpoints(s Service) *Endpoints {
	// Casting service to Auther interface
	a := s.(Auther)
	return &Endpoints{
		MonitorRequest:      NewMonitorRequestEndpoint(s, a.OAuth2Auth),
		Monitor:             NewMonitorEndpoint(s),
		List:                NewListEndpoint(s, a.OAuth2Auth),
		Show:                NewShowEndpoint(s, a.OAuth2Auth),
		PreservationActions: NewPreservationActionsEndpoint(s, a.OAuth2Auth),
		Confirm:             NewConfirmEndpoint(s, a.OAuth2Auth),
		Reject:              NewRejectEndpoint(s, a.OAuth2Auth),
		Move:                NewMoveEndpoint(s, a.OAuth2Auth),
		MoveStatus:          NewMoveStatusEndpoint(s, a.OAuth2Auth),
	}
}

// Use applies the given middleware to all the "package" service endpoints.
func (e *Endpoints) Use(m func(goa.Endpoint) goa.Endpoint) {
	e.MonitorRequest = m(e.MonitorRequest)
	e.Monitor = m(e.Monitor)
	e.List = m(e.List)
	e.Show = m(e.Show)
	e.PreservationActions = m(e.PreservationActions)
	e.Confirm = m(e.Confirm)
	e.Reject = m(e.Reject)
	e.Move = m(e.Move)
	e.MoveStatus = m(e.MoveStatus)
}

// NewMonitorRequestEndpoint returns an endpoint function that calls the method
// "monitor_request" of service "package".
func NewMonitorRequestEndpoint(s Service, authOAuth2Fn security.AuthOAuth2Func) goa.Endpoint {
	return func(ctx context.Context, req any) (any, error) {
		p := req.(*MonitorRequestPayload)
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
		if p.OauthToken != nil {
			token = *p.OauthToken
		}
		ctx, err = authOAuth2Fn(ctx, token, &sc)
		if err != nil {
			return nil, err
		}
		return s.MonitorRequest(ctx, p)
	}
}

// NewMonitorEndpoint returns an endpoint function that calls the method
// "monitor" of service "package".
func NewMonitorEndpoint(s Service) goa.Endpoint {
	return func(ctx context.Context, req any) (any, error) {
		ep := req.(*MonitorEndpointInput)
		return nil, s.Monitor(ctx, ep.Payload, ep.Stream)
	}
}

// NewListEndpoint returns an endpoint function that calls the method "list" of
// service "package".
func NewListEndpoint(s Service, authOAuth2Fn security.AuthOAuth2Func) goa.Endpoint {
	return func(ctx context.Context, req any) (any, error) {
		p := req.(*ListPayload)
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
		if p.OauthToken != nil {
			token = *p.OauthToken
		}
		ctx, err = authOAuth2Fn(ctx, token, &sc)
		if err != nil {
			return nil, err
		}
		return s.List(ctx, p)
	}
}

// NewShowEndpoint returns an endpoint function that calls the method "show" of
// service "package".
func NewShowEndpoint(s Service, authOAuth2Fn security.AuthOAuth2Func) goa.Endpoint {
	return func(ctx context.Context, req any) (any, error) {
		p := req.(*ShowPayload)
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
		if p.OauthToken != nil {
			token = *p.OauthToken
		}
		ctx, err = authOAuth2Fn(ctx, token, &sc)
		if err != nil {
			return nil, err
		}
		res, err := s.Show(ctx, p)
		if err != nil {
			return nil, err
		}
		vres := NewViewedEnduroStoredPackage(res, "default")
		return vres, nil
	}
}

// NewPreservationActionsEndpoint returns an endpoint function that calls the
// method "preservation_actions" of service "package".
func NewPreservationActionsEndpoint(s Service, authOAuth2Fn security.AuthOAuth2Func) goa.Endpoint {
	return func(ctx context.Context, req any) (any, error) {
		p := req.(*PreservationActionsPayload)
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
		if p.OauthToken != nil {
			token = *p.OauthToken
		}
		ctx, err = authOAuth2Fn(ctx, token, &sc)
		if err != nil {
			return nil, err
		}
		res, err := s.PreservationActions(ctx, p)
		if err != nil {
			return nil, err
		}
		vres := NewViewedEnduroPackagePreservationActions(res, "default")
		return vres, nil
	}
}

// NewConfirmEndpoint returns an endpoint function that calls the method
// "confirm" of service "package".
func NewConfirmEndpoint(s Service, authOAuth2Fn security.AuthOAuth2Func) goa.Endpoint {
	return func(ctx context.Context, req any) (any, error) {
		p := req.(*ConfirmPayload)
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
		if p.OauthToken != nil {
			token = *p.OauthToken
		}
		ctx, err = authOAuth2Fn(ctx, token, &sc)
		if err != nil {
			return nil, err
		}
		return nil, s.Confirm(ctx, p)
	}
}

// NewRejectEndpoint returns an endpoint function that calls the method
// "reject" of service "package".
func NewRejectEndpoint(s Service, authOAuth2Fn security.AuthOAuth2Func) goa.Endpoint {
	return func(ctx context.Context, req any) (any, error) {
		p := req.(*RejectPayload)
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
		if p.OauthToken != nil {
			token = *p.OauthToken
		}
		ctx, err = authOAuth2Fn(ctx, token, &sc)
		if err != nil {
			return nil, err
		}
		return nil, s.Reject(ctx, p)
	}
}

// NewMoveEndpoint returns an endpoint function that calls the method "move" of
// service "package".
func NewMoveEndpoint(s Service, authOAuth2Fn security.AuthOAuth2Func) goa.Endpoint {
	return func(ctx context.Context, req any) (any, error) {
		p := req.(*MovePayload)
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
		if p.OauthToken != nil {
			token = *p.OauthToken
		}
		ctx, err = authOAuth2Fn(ctx, token, &sc)
		if err != nil {
			return nil, err
		}
		return nil, s.Move(ctx, p)
	}
}

// NewMoveStatusEndpoint returns an endpoint function that calls the method
// "move_status" of service "package".
func NewMoveStatusEndpoint(s Service, authOAuth2Fn security.AuthOAuth2Func) goa.Endpoint {
	return func(ctx context.Context, req any) (any, error) {
		p := req.(*MoveStatusPayload)
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
		if p.OauthToken != nil {
			token = *p.OauthToken
		}
		ctx, err = authOAuth2Fn(ctx, token, &sc)
		if err != nil {
			return nil, err
		}
		return s.MoveStatus(ctx, p)
	}
}
