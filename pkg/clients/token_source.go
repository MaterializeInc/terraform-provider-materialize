package clients

import (
	"context"
	"net/http"
	"time"
)

// requestTimeout bounds every outbound API call. Without it a hung endpoint
// stalls a Terraform operation indefinitely with no way out but Ctrl-C.
const requestTimeout = 30 * time.Second

// authClient issues the unauthenticated token requests themselves, so it must
// not carry a bearerTransport.
var authClient = &http.Client{Timeout: requestTimeout}

// TokenSource supplies a currently valid bearer token, refreshing it when it has
// expired. Implementations must be safe for concurrent use: Terraform runs
// resource operations in parallel.
type TokenSource interface {
	Token(ctx context.Context) (string, error)
}

// bearerTransport authorizes each request with a token taken from source at the
// moment the request is sent. Resolving per request means callers never hold a
// stale token and nothing has to be swapped out when the token is refreshed.
type bearerTransport struct {
	source TokenSource
	base   http.RoundTripper
}

func (t *bearerTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	token, err := t.source.Token(req.Context())
	if err != nil {
		return nil, err
	}

	// RoundTrip must not modify the request it is given.
	r := req.Clone(req.Context())
	r.Header.Set("Authorization", "Bearer "+token)

	base := t.base
	if base == nil {
		base = http.DefaultTransport
	}
	return base.RoundTrip(r)
}

// newAuthorizedClient returns an HTTP client that authorizes every request from
// source and gives up after requestTimeout.
func newAuthorizedClient(source TokenSource) *http.Client {
	return &http.Client{
		Timeout:   requestTimeout,
		Transport: &bearerTransport{source: source},
	}
}
