package auth

import (
	"context"
	"encoding/json"
	"slices"
	"strings"
)

type Claims struct {
	Email         string `json:"email,omitempty"`
	EmailVerified bool   `json:"email_verified,omitempty"`
	Name          string `json:"name,omitempty"`
	Iss           string `json:"iss,omitempty"`
	Sub           string `json:"sub,omitempty"`
	// The attributes are parsed from a configured claim and added here,
	// they are needed in the JSON representation for the MarshalBinary and
	// UnmarshalBinary methods below. We use the `enduro_internal_attributes`
	// JSON key to reduce the possibility of a conflict when the JWT is parsed.
	Attributes []string `json:"enduro_internal_attributes,omitempty"`
}

// MarshalBinary implements encoding.BinaryMarshaler for Redis compatibility.
func (c *Claims) MarshalBinary() ([]byte, error) {
	return json.Marshal(c)
}

// UnmarshalBinary implements encoding.BinaryUnmarshaler for Redis compatibility.
func (c *Claims) UnmarshalBinary(data []byte) error {
	return json.Unmarshal(data, c)
}

// CheckAttributes verifies all required attributes are present in the claim
// attributes. It always verifies if the claim is nil (authentication disabled)
// or the attributes are nil (access control disabled). Attributes are verified
// by exact match or by having an ancestor with wildcard. For example, a claim
// with "*" or "ingest:*" as one of it's attributes will verify all ingest
// actions, like "ingest:sips:list", "ingest:sips:read", etc.
func (c *Claims) CheckAttributes(required []string) bool {
	// Authentication disabled, access control disabled or all wildcard in claims.
	if c == nil || c.Attributes == nil || slices.Contains(c.Attributes, "*") {
		return true
	}

	// Check for all required attributes considering wildcards.
	for _, attr := range required {
		for !slices.Contains(c.Attributes, attr) {

			attr, _ = strings.CutSuffix(attr, ":*")
			lastColonIndex := strings.LastIndex(attr, ":")
			if lastColonIndex == -1 {
				return false
			}

			attr = attr[:lastColonIndex] + ":*"
		}
	}

	return true
}

type contextUserClaimsType struct{}

var contextUserClaimsKey = &contextUserClaimsType{}

// WithUserClaims puts the user claims into the current context.
func WithUserClaims(ctx context.Context, claims *Claims) context.Context {
	return context.WithValue(ctx, contextUserClaimsKey, claims)
}

// UserClaimsFromContext returns the user claims from the context.
// A nil value is returned if they are not found.
func UserClaimsFromContext(ctx context.Context) *Claims {
	v := ctx.Value(contextUserClaimsKey)
	if v == nil {
		return nil
	}
	c, ok := v.(*Claims)
	if !ok {
		return nil
	}
	return c
}
