//go:build ignore

// This script performs a manual OAuth2 Authorization Code flow with PKCE
// against a Keycloak (or compatible OIDC) provider. It prints an
// authorization URL, prompts the user to log in via browser, and exchanges
// the returned code for an access token. The decoded token payload is
// printed for inspection. Intended for API testing and development, not
// production use.

package main

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"

	"golang.org/x/oauth2"
)

const usage = `Usage:

    $ go run ./hack/auth/main.go http://keycloak:7470/realms/realm_id client_id scope1,scope2,scope3
`

func main() {
	flag.Usage = func() { fmt.Fprintf(os.Stderr, "%s\n", usage) }
	if len(os.Args) != 4 {
		flag.Usage()
		os.Exit(1)
	}

	host := strings.TrimSuffix(os.Args[1], "/")
	client := os.Args[2]
	scopes := strings.Split(os.Args[3], ",")
	conf := &oauth2.Config{
		ClientID:    client,
		RedirectURL: "urn:ietf:wg:oauth:2.0:oob",
		Scopes:      scopes,
		Endpoint: oauth2.Endpoint{
			AuthURL:  fmt.Sprintf("%s/protocol/openid-connect/auth", host),
			TokenURL: fmt.Sprintf("%s/protocol/openid-connect/token", host),
		},
	}

	cv, err := codeVerifier()
	if err != nil {
		fmt.Printf("Unable to generate code verifier: %v\n", err)
		os.Exit(1)
	}
	cc := codeChallenge(cv)
	authURL := conf.AuthCodeURL(
		"state",
		oauth2.SetAuthURLParam("code_challenge", cc),
		oauth2.SetAuthURLParam("code_challenge_method", "S256"),
	)

	fmt.Printf("Open this URL in your browser and log in:\n\n%s\n", authURL)
	fmt.Print("\nPaste the code from the browser here:\n\n")

	var code string
	fmt.Scanln(&code)

	token, err := conf.Exchange(
		context.Background(),
		code,
		oauth2.SetAuthURLParam("code_verifier", cv),
	)
	if err != nil {
		fmt.Printf("Unable to get token: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("\nAccess token payload:\n\n%s\n", tokenPayload(token.AccessToken))
	fmt.Printf("\nAccess token value:\n\n%s\n", token.AccessToken)
}

func codeVerifier() (string, error) {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

func codeChallenge(verifier string) string {
	h := sha256.New()
	h.Write([]byte(verifier))
	return base64.RawURLEncoding.EncodeToString(h.Sum(nil))
}

func tokenPayload(token string) string {
	parts := strings.Split(token, ".")
	if len(parts) < 2 {
		return "Access token is not a valid JWT."
	}
	b, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return "Failed to decode access token."
	}
	var out map[string]any
	if err := json.Unmarshal(b, &out); err != nil {
		return string(b)
	}
	f, err := json.MarshalIndent(out, "", "  ")
	if err != nil {
		return string(b)
	}
	return string(f)
}
