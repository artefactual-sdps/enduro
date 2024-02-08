package ssblob

import (
	"fmt"
	"net/http"
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
