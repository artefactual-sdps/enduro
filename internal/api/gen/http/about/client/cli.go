// Code generated by goa v3.15.2, DO NOT EDIT.
//
// about HTTP client CLI support package
//
// Command:
// $ goa gen github.com/artefactual-sdps/enduro/internal/api/design -o
// internal/api

package client

import (
	about "github.com/artefactual-sdps/enduro/internal/api/gen/about"
)

// BuildAboutPayload builds the payload for the about about endpoint from CLI
// flags.
func BuildAboutPayload(aboutAboutToken string) (*about.AboutPayload, error) {
	var token *string
	{
		if aboutAboutToken != "" {
			token = &aboutAboutToken
		}
	}
	v := &about.AboutPayload{}
	v.Token = token

	return v, nil
}
