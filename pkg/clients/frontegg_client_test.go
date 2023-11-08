package clients

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewFronteggClient(t *testing.T) {
	// Start a local HTTP server to mock the Frontegg API
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/identity/resources/auth/v1/api-token" {
			w.Header().Set("Content-Type", "application/json")
			response := map[string]string{
				"accessToken": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJlbWFpbCI6Im16X3N5c3RlbSIsImV4cCI6MTcwMDAwMDAwMH0.c2lnbmF0dXJl",
			}
			json.NewEncoder(w).Encode(response)
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	// Use the server's URL as the endpoint
	endpoint := server.URL
	password := "mzp_" + strings.Repeat("a", 64)

	// Create the Frontegg client using the mocked server and context
	fronteggClient, err := NewFronteggClient(context.Background(), password, endpoint)
	require.NoError(t, err, "Error should be nil")
	require.NotNil(t, fronteggClient, "Frontegg client should not be nil")

	// The token should be set correctly in the Frontegg client
	require.Equal(t, "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJlbWFpbCI6Im16X3N5c3RlbSIsImV4cCI6MTcwMDAwMDAwMH0.c2lnbmF0dXJl", fronteggClient.Token, "Token should be set correctly in the Frontegg client")
}
