package ssblob

import (
	"fmt"
	"net/http"

	"github.com/hashicorp/go-cleanhttp"
)

// transport injects the authentication header.
type transport struct {
	key string
	t   *http.Transport
}

func (t *transport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Authorization", fmt.Sprintf("ApiKey %s", t.key))

	return t.t.RoundTrip(req)
}

func NewClient(user, key string) *http.Client {
	return &http.Client{
		Transport: &transport{
			key: fmt.Sprintf("%s:%s", user, key),
			t:   cleanhttp.DefaultPooledTransport(),
		},
	}
}
