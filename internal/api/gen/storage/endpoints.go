// Code generated by goa v3.15.2, DO NOT EDIT.
//
// storage endpoints
//
// Command:
// $ goa gen github.com/artefactual-sdps/enduro/internal/api/design -o
// internal/api

package storage

import (
	"context"

	goa "goa.design/goa/v3/pkg"
	"goa.design/goa/v3/security"
)

// Endpoints wraps the "storage" service endpoints.
type Endpoints struct {
	ListAips         goa.Endpoint
	CreateAip        goa.Endpoint
	SubmitAip        goa.Endpoint
	UpdateAip        goa.Endpoint
	DownloadAip      goa.Endpoint
	MoveAip          goa.Endpoint
	MoveAipStatus    goa.Endpoint
	RejectAip        goa.Endpoint
	ShowAip          goa.Endpoint
	ListLocations    goa.Endpoint
	CreateLocation   goa.Endpoint
	ShowLocation     goa.Endpoint
	ListLocationAips goa.Endpoint
}

// NewEndpoints wraps the methods of the "storage" service with endpoints.
func NewEndpoints(s Service) *Endpoints {
	// Casting service to Auther interface
	a := s.(Auther)
	return &Endpoints{
		ListAips:         NewListAipsEndpoint(s, a.JWTAuth),
		CreateAip:        NewCreateAipEndpoint(s, a.JWTAuth),
		SubmitAip:        NewSubmitAipEndpoint(s, a.JWTAuth),
		UpdateAip:        NewUpdateAipEndpoint(s, a.JWTAuth),
		DownloadAip:      NewDownloadAipEndpoint(s, a.JWTAuth),
		MoveAip:          NewMoveAipEndpoint(s, a.JWTAuth),
		MoveAipStatus:    NewMoveAipStatusEndpoint(s, a.JWTAuth),
		RejectAip:        NewRejectAipEndpoint(s, a.JWTAuth),
		ShowAip:          NewShowAipEndpoint(s, a.JWTAuth),
		ListLocations:    NewListLocationsEndpoint(s, a.JWTAuth),
		CreateLocation:   NewCreateLocationEndpoint(s, a.JWTAuth),
		ShowLocation:     NewShowLocationEndpoint(s, a.JWTAuth),
		ListLocationAips: NewListLocationAipsEndpoint(s, a.JWTAuth),
	}
}

// Use applies the given middleware to all the "storage" service endpoints.
func (e *Endpoints) Use(m func(goa.Endpoint) goa.Endpoint) {
	e.ListAips = m(e.ListAips)
	e.CreateAip = m(e.CreateAip)
	e.SubmitAip = m(e.SubmitAip)
	e.UpdateAip = m(e.UpdateAip)
	e.DownloadAip = m(e.DownloadAip)
	e.MoveAip = m(e.MoveAip)
	e.MoveAipStatus = m(e.MoveAipStatus)
	e.RejectAip = m(e.RejectAip)
	e.ShowAip = m(e.ShowAip)
	e.ListLocations = m(e.ListLocations)
	e.CreateLocation = m(e.CreateLocation)
	e.ShowLocation = m(e.ShowLocation)
	e.ListLocationAips = m(e.ListLocationAips)
}

// NewListAipsEndpoint returns an endpoint function that calls the method
// "list_aips" of service "storage".
func NewListAipsEndpoint(s Service, authJWTFn security.AuthJWTFunc) goa.Endpoint {
	return func(ctx context.Context, req any) (any, error) {
		p := req.(*ListAipsPayload)
		var err error
		sc := security.JWTScheme{
			Name:           "jwt",
			Scopes:         []string{"ingest:sips:actions:list", "ingest:sips:list", "ingest:sips:move", "ingest:sips:read", "ingest:sips:review", "ingest:sips:upload", "storage:aips:create", "storage:aips:download", "storage:aips:list", "storage:aips:move", "storage:aips:read", "storage:aips:review", "storage:aips:submit", "storage:locations:aips:list", "storage:locations:create", "storage:locations:list", "storage:locations:read"},
			RequiredScopes: []string{"storage:aips:list"},
		}
		var token string
		if p.Token != nil {
			token = *p.Token
		}
		ctx, err = authJWTFn(ctx, token, &sc)
		if err != nil {
			return nil, err
		}
		res, err := s.ListAips(ctx, p)
		if err != nil {
			return nil, err
		}
		vres := NewViewedAIPs(res, "default")
		return vres, nil
	}
}

// NewCreateAipEndpoint returns an endpoint function that calls the method
// "create_aip" of service "storage".
func NewCreateAipEndpoint(s Service, authJWTFn security.AuthJWTFunc) goa.Endpoint {
	return func(ctx context.Context, req any) (any, error) {
		p := req.(*CreateAipPayload)
		var err error
		sc := security.JWTScheme{
			Name:           "jwt",
			Scopes:         []string{"ingest:sips:actions:list", "ingest:sips:list", "ingest:sips:move", "ingest:sips:read", "ingest:sips:review", "ingest:sips:upload", "storage:aips:create", "storage:aips:download", "storage:aips:list", "storage:aips:move", "storage:aips:read", "storage:aips:review", "storage:aips:submit", "storage:locations:aips:list", "storage:locations:create", "storage:locations:list", "storage:locations:read"},
			RequiredScopes: []string{"storage:aips:create"},
		}
		var token string
		if p.Token != nil {
			token = *p.Token
		}
		ctx, err = authJWTFn(ctx, token, &sc)
		if err != nil {
			return nil, err
		}
		res, err := s.CreateAip(ctx, p)
		if err != nil {
			return nil, err
		}
		vres := NewViewedAIP(res, "default")
		return vres, nil
	}
}

// NewSubmitAipEndpoint returns an endpoint function that calls the method
// "submit_aip" of service "storage".
func NewSubmitAipEndpoint(s Service, authJWTFn security.AuthJWTFunc) goa.Endpoint {
	return func(ctx context.Context, req any) (any, error) {
		p := req.(*SubmitAipPayload)
		var err error
		sc := security.JWTScheme{
			Name:           "jwt",
			Scopes:         []string{"ingest:sips:actions:list", "ingest:sips:list", "ingest:sips:move", "ingest:sips:read", "ingest:sips:review", "ingest:sips:upload", "storage:aips:create", "storage:aips:download", "storage:aips:list", "storage:aips:move", "storage:aips:read", "storage:aips:review", "storage:aips:submit", "storage:locations:aips:list", "storage:locations:create", "storage:locations:list", "storage:locations:read"},
			RequiredScopes: []string{"storage:aips:submit"},
		}
		var token string
		if p.Token != nil {
			token = *p.Token
		}
		ctx, err = authJWTFn(ctx, token, &sc)
		if err != nil {
			return nil, err
		}
		return s.SubmitAip(ctx, p)
	}
}

// NewUpdateAipEndpoint returns an endpoint function that calls the method
// "update_aip" of service "storage".
func NewUpdateAipEndpoint(s Service, authJWTFn security.AuthJWTFunc) goa.Endpoint {
	return func(ctx context.Context, req any) (any, error) {
		p := req.(*UpdateAipPayload)
		var err error
		sc := security.JWTScheme{
			Name:           "jwt",
			Scopes:         []string{"ingest:sips:actions:list", "ingest:sips:list", "ingest:sips:move", "ingest:sips:read", "ingest:sips:review", "ingest:sips:upload", "storage:aips:create", "storage:aips:download", "storage:aips:list", "storage:aips:move", "storage:aips:read", "storage:aips:review", "storage:aips:submit", "storage:locations:aips:list", "storage:locations:create", "storage:locations:list", "storage:locations:read"},
			RequiredScopes: []string{"storage:aips:submit"},
		}
		var token string
		if p.Token != nil {
			token = *p.Token
		}
		ctx, err = authJWTFn(ctx, token, &sc)
		if err != nil {
			return nil, err
		}
		return nil, s.UpdateAip(ctx, p)
	}
}

// NewDownloadAipEndpoint returns an endpoint function that calls the method
// "download_aip" of service "storage".
func NewDownloadAipEndpoint(s Service, authJWTFn security.AuthJWTFunc) goa.Endpoint {
	return func(ctx context.Context, req any) (any, error) {
		p := req.(*DownloadAipPayload)
		var err error
		sc := security.JWTScheme{
			Name:           "jwt",
			Scopes:         []string{"ingest:sips:actions:list", "ingest:sips:list", "ingest:sips:move", "ingest:sips:read", "ingest:sips:review", "ingest:sips:upload", "storage:aips:create", "storage:aips:download", "storage:aips:list", "storage:aips:move", "storage:aips:read", "storage:aips:review", "storage:aips:submit", "storage:locations:aips:list", "storage:locations:create", "storage:locations:list", "storage:locations:read"},
			RequiredScopes: []string{"storage:aips:download"},
		}
		var token string
		if p.Token != nil {
			token = *p.Token
		}
		ctx, err = authJWTFn(ctx, token, &sc)
		if err != nil {
			return nil, err
		}
		return s.DownloadAip(ctx, p)
	}
}

// NewMoveAipEndpoint returns an endpoint function that calls the method
// "move_aip" of service "storage".
func NewMoveAipEndpoint(s Service, authJWTFn security.AuthJWTFunc) goa.Endpoint {
	return func(ctx context.Context, req any) (any, error) {
		p := req.(*MoveAipPayload)
		var err error
		sc := security.JWTScheme{
			Name:           "jwt",
			Scopes:         []string{"ingest:sips:actions:list", "ingest:sips:list", "ingest:sips:move", "ingest:sips:read", "ingest:sips:review", "ingest:sips:upload", "storage:aips:create", "storage:aips:download", "storage:aips:list", "storage:aips:move", "storage:aips:read", "storage:aips:review", "storage:aips:submit", "storage:locations:aips:list", "storage:locations:create", "storage:locations:list", "storage:locations:read"},
			RequiredScopes: []string{"storage:aips:move"},
		}
		var token string
		if p.Token != nil {
			token = *p.Token
		}
		ctx, err = authJWTFn(ctx, token, &sc)
		if err != nil {
			return nil, err
		}
		return nil, s.MoveAip(ctx, p)
	}
}

// NewMoveAipStatusEndpoint returns an endpoint function that calls the method
// "move_aip_status" of service "storage".
func NewMoveAipStatusEndpoint(s Service, authJWTFn security.AuthJWTFunc) goa.Endpoint {
	return func(ctx context.Context, req any) (any, error) {
		p := req.(*MoveAipStatusPayload)
		var err error
		sc := security.JWTScheme{
			Name:           "jwt",
			Scopes:         []string{"ingest:sips:actions:list", "ingest:sips:list", "ingest:sips:move", "ingest:sips:read", "ingest:sips:review", "ingest:sips:upload", "storage:aips:create", "storage:aips:download", "storage:aips:list", "storage:aips:move", "storage:aips:read", "storage:aips:review", "storage:aips:submit", "storage:locations:aips:list", "storage:locations:create", "storage:locations:list", "storage:locations:read"},
			RequiredScopes: []string{"storage:aips:move"},
		}
		var token string
		if p.Token != nil {
			token = *p.Token
		}
		ctx, err = authJWTFn(ctx, token, &sc)
		if err != nil {
			return nil, err
		}
		return s.MoveAipStatus(ctx, p)
	}
}

// NewRejectAipEndpoint returns an endpoint function that calls the method
// "reject_aip" of service "storage".
func NewRejectAipEndpoint(s Service, authJWTFn security.AuthJWTFunc) goa.Endpoint {
	return func(ctx context.Context, req any) (any, error) {
		p := req.(*RejectAipPayload)
		var err error
		sc := security.JWTScheme{
			Name:           "jwt",
			Scopes:         []string{"ingest:sips:actions:list", "ingest:sips:list", "ingest:sips:move", "ingest:sips:read", "ingest:sips:review", "ingest:sips:upload", "storage:aips:create", "storage:aips:download", "storage:aips:list", "storage:aips:move", "storage:aips:read", "storage:aips:review", "storage:aips:submit", "storage:locations:aips:list", "storage:locations:create", "storage:locations:list", "storage:locations:read"},
			RequiredScopes: []string{"storage:aips:review"},
		}
		var token string
		if p.Token != nil {
			token = *p.Token
		}
		ctx, err = authJWTFn(ctx, token, &sc)
		if err != nil {
			return nil, err
		}
		return nil, s.RejectAip(ctx, p)
	}
}

// NewShowAipEndpoint returns an endpoint function that calls the method
// "show_aip" of service "storage".
func NewShowAipEndpoint(s Service, authJWTFn security.AuthJWTFunc) goa.Endpoint {
	return func(ctx context.Context, req any) (any, error) {
		p := req.(*ShowAipPayload)
		var err error
		sc := security.JWTScheme{
			Name:           "jwt",
			Scopes:         []string{"ingest:sips:actions:list", "ingest:sips:list", "ingest:sips:move", "ingest:sips:read", "ingest:sips:review", "ingest:sips:upload", "storage:aips:create", "storage:aips:download", "storage:aips:list", "storage:aips:move", "storage:aips:read", "storage:aips:review", "storage:aips:submit", "storage:locations:aips:list", "storage:locations:create", "storage:locations:list", "storage:locations:read"},
			RequiredScopes: []string{"storage:aips:read"},
		}
		var token string
		if p.Token != nil {
			token = *p.Token
		}
		ctx, err = authJWTFn(ctx, token, &sc)
		if err != nil {
			return nil, err
		}
		res, err := s.ShowAip(ctx, p)
		if err != nil {
			return nil, err
		}
		vres := NewViewedAIP(res, "default")
		return vres, nil
	}
}

// NewListLocationsEndpoint returns an endpoint function that calls the method
// "list_locations" of service "storage".
func NewListLocationsEndpoint(s Service, authJWTFn security.AuthJWTFunc) goa.Endpoint {
	return func(ctx context.Context, req any) (any, error) {
		p := req.(*ListLocationsPayload)
		var err error
		sc := security.JWTScheme{
			Name:           "jwt",
			Scopes:         []string{"ingest:sips:actions:list", "ingest:sips:list", "ingest:sips:move", "ingest:sips:read", "ingest:sips:review", "ingest:sips:upload", "storage:aips:create", "storage:aips:download", "storage:aips:list", "storage:aips:move", "storage:aips:read", "storage:aips:review", "storage:aips:submit", "storage:locations:aips:list", "storage:locations:create", "storage:locations:list", "storage:locations:read"},
			RequiredScopes: []string{"storage:locations:list"},
		}
		var token string
		if p.Token != nil {
			token = *p.Token
		}
		ctx, err = authJWTFn(ctx, token, &sc)
		if err != nil {
			return nil, err
		}
		res, err := s.ListLocations(ctx, p)
		if err != nil {
			return nil, err
		}
		vres := NewViewedLocationCollection(res, "default")
		return vres, nil
	}
}

// NewCreateLocationEndpoint returns an endpoint function that calls the method
// "create_location" of service "storage".
func NewCreateLocationEndpoint(s Service, authJWTFn security.AuthJWTFunc) goa.Endpoint {
	return func(ctx context.Context, req any) (any, error) {
		p := req.(*CreateLocationPayload)
		var err error
		sc := security.JWTScheme{
			Name:           "jwt",
			Scopes:         []string{"ingest:sips:actions:list", "ingest:sips:list", "ingest:sips:move", "ingest:sips:read", "ingest:sips:review", "ingest:sips:upload", "storage:aips:create", "storage:aips:download", "storage:aips:list", "storage:aips:move", "storage:aips:read", "storage:aips:review", "storage:aips:submit", "storage:locations:aips:list", "storage:locations:create", "storage:locations:list", "storage:locations:read"},
			RequiredScopes: []string{"storage:locations:create"},
		}
		var token string
		if p.Token != nil {
			token = *p.Token
		}
		ctx, err = authJWTFn(ctx, token, &sc)
		if err != nil {
			return nil, err
		}
		return s.CreateLocation(ctx, p)
	}
}

// NewShowLocationEndpoint returns an endpoint function that calls the method
// "show_location" of service "storage".
func NewShowLocationEndpoint(s Service, authJWTFn security.AuthJWTFunc) goa.Endpoint {
	return func(ctx context.Context, req any) (any, error) {
		p := req.(*ShowLocationPayload)
		var err error
		sc := security.JWTScheme{
			Name:           "jwt",
			Scopes:         []string{"ingest:sips:actions:list", "ingest:sips:list", "ingest:sips:move", "ingest:sips:read", "ingest:sips:review", "ingest:sips:upload", "storage:aips:create", "storage:aips:download", "storage:aips:list", "storage:aips:move", "storage:aips:read", "storage:aips:review", "storage:aips:submit", "storage:locations:aips:list", "storage:locations:create", "storage:locations:list", "storage:locations:read"},
			RequiredScopes: []string{"storage:locations:read"},
		}
		var token string
		if p.Token != nil {
			token = *p.Token
		}
		ctx, err = authJWTFn(ctx, token, &sc)
		if err != nil {
			return nil, err
		}
		res, err := s.ShowLocation(ctx, p)
		if err != nil {
			return nil, err
		}
		vres := NewViewedLocation(res, "default")
		return vres, nil
	}
}

// NewListLocationAipsEndpoint returns an endpoint function that calls the
// method "list_location_aips" of service "storage".
func NewListLocationAipsEndpoint(s Service, authJWTFn security.AuthJWTFunc) goa.Endpoint {
	return func(ctx context.Context, req any) (any, error) {
		p := req.(*ListLocationAipsPayload)
		var err error
		sc := security.JWTScheme{
			Name:           "jwt",
			Scopes:         []string{"ingest:sips:actions:list", "ingest:sips:list", "ingest:sips:move", "ingest:sips:read", "ingest:sips:review", "ingest:sips:upload", "storage:aips:create", "storage:aips:download", "storage:aips:list", "storage:aips:move", "storage:aips:read", "storage:aips:review", "storage:aips:submit", "storage:locations:aips:list", "storage:locations:create", "storage:locations:list", "storage:locations:read"},
			RequiredScopes: []string{"storage:locations:aips:list"},
		}
		var token string
		if p.Token != nil {
			token = *p.Token
		}
		ctx, err = authJWTFn(ctx, token, &sc)
		if err != nil {
			return nil, err
		}
		res, err := s.ListLocationAips(ctx, p)
		if err != nil {
			return nil, err
		}
		vres := NewViewedAIPCollection(res, "default")
		return vres, nil
	}
}
