package clients

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// countingAuthServer serves the token endpoint and records how many times it was
// asked for a token.
func countingAuthServer(t *testing.T, calls *int64) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/identity/resources/auth/v1/api-token" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		atomic.AddInt64(calls, 1)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"accessToken":"` + generateValidJWTToken() + `","expiresIn":3600}`))
	}))
	t.Cleanup(srv.Close)
	return srv
}

func newTestFronteggClient(t *testing.T, endpoint string) *FronteggClient {
	t.Helper()
	c, err := NewFronteggClient(context.Background(), "mzp_"+strings.Repeat("a", 64), endpoint)
	require.NoError(t, err)
	return c
}

func TestTokenIsCachedUntilItExpires(t *testing.T) {
	var calls int64
	srv := countingAuthServer(t, &calls)
	c := newTestFronteggClient(t, srv.URL)
	require.EqualValues(t, 1, atomic.LoadInt64(&calls), "constructing the client fetches one token")

	for i := 0; i < 5; i++ {
		_, err := c.Token(context.Background())
		require.NoError(t, err)
	}
	require.EqualValues(t, 1, atomic.LoadInt64(&calls), "a valid token should be reused")
}

func TestTokenRefreshesOnceForConcurrentCallers(t *testing.T) {
	var calls int64
	srv := countingAuthServer(t, &calls)
	c := newTestFronteggClient(t, srv.URL)

	// Force the cached token past its refresh deadline.
	c.mu.Lock()
	c.tokenExpiry = time.Now().Add(-time.Hour)
	c.mu.Unlock()

	// This is the case that used to fire one token request per goroutine and race
	// on the shared client.
	const goroutines = 20
	var wg sync.WaitGroup
	errs := make(chan error, goroutines)
	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if _, err := c.Token(context.Background()); err != nil {
				errs <- err
			}
		}()
	}
	wg.Wait()
	close(errs)
	for err := range errs {
		require.NoError(t, err)
	}

	require.EqualValues(t, 2, atomic.LoadInt64(&calls),
		"an expired token should be refreshed exactly once regardless of caller count")
}

// stubTokenSource is the whole fake a client needs now.
type stubTokenSource struct {
	token string
	err   error
}

func (s stubTokenSource) Token(context.Context) (string, error) { return s.token, s.err }

type capturingTransport struct {
	gotAuth  string
	gotOther string
}

func (c *capturingTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	c.gotAuth = req.Header.Get("Authorization")
	c.gotOther = req.Header.Get("X-Original")
	return &http.Response{StatusCode: http.StatusOK, Body: http.NoBody, Header: make(http.Header)}, nil
}

func TestBearerTransportAuthorizesEachRequest(t *testing.T) {
	capture := &capturingTransport{}
	rt := &bearerTransport{source: stubTokenSource{token: "abc123"}, base: capture}

	req, err := http.NewRequest(http.MethodGet, "https://example.invalid/x", nil)
	require.NoError(t, err)
	req.Header.Set("X-Original", "kept")

	resp, err := rt.RoundTrip(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, "Bearer abc123", capture.gotAuth, "the token should be applied per request")
	require.Equal(t, "kept", capture.gotOther, "existing headers should survive")
	require.Empty(t, req.Header.Get("Authorization"), "RoundTrip must not mutate the caller's request")
}

func TestBearerTransportPropagatesTokenFailure(t *testing.T) {
	rt := &bearerTransport{source: stubTokenSource{err: context.DeadlineExceeded}, base: &capturingTransport{}}

	req, err := http.NewRequest(http.MethodGet, "https://example.invalid/x", nil)
	require.NoError(t, err)

	_, err = rt.RoundTrip(req)
	require.Error(t, err, "a token failure should surface instead of sending an unauthorized request")
}

func TestAuthorizedClientsAreBounded(t *testing.T) {
	require.Equal(t, requestTimeout, newAuthorizedClient(stubTokenSource{}).Timeout)
	require.Equal(t, requestTimeout, authClient.Timeout, "the token request itself must be bounded too")
}
