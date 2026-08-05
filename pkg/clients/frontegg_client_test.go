package clients

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
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

	// The token fetched at construction should be the one handed to callers.
	token, err := fronteggClient.Token(context.Background())
	require.NoError(t, err)
	require.Equal(t, "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJlbWFpbCI6Im16X3N5c3RlbSIsImV4cCI6MTcwMDAwMDAwMH0.c2lnbmF0dXJl", token, "Token should be set correctly in the Frontegg client")
}

func TestFronteggClient_AuthenticationError(t *testing.T) {
	// Start a local HTTP server to mock the Frontegg API with an error response
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized) // Simulate an authentication error
	}))
	defer server.Close()

	// Use the server's URL as the endpoint
	endpoint := server.URL
	password := "mzp_" + strings.Repeat("a", 64)

	// Create the Frontegg client using the mocked server and context
	_, err := NewFronteggClient(context.Background(), password, endpoint)
	require.Error(t, err, "Authentication error should result in an error")
}

func TestFronteggClient_TokenRefresh(t *testing.T) {
	// Create a mock HTTP server to simulate the Frontegg API
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/identity/resources/auth/v1/api-token" {
			// Generate a valid JWT token
			token := generateValidJWTToken()

			// Simulate a successful token refresh response with the generated JWT token
			w.Header().Set("Content-Type", "application/json")
			response := map[string]string{
				"accessToken": token,
			}
			json.NewEncoder(w).Encode(response)
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	// Use the mock server's URL as the endpoint
	endpoint := server.URL
	password := "mzp_" + strings.Repeat("a", 64)

	// Create the Frontegg client using the mock server and context
	fronteggClient, err := NewFronteggClient(context.Background(), password, endpoint)
	require.NoError(t, err, "Error should be nil")
	require.NotNil(t, fronteggClient, "Frontegg client should not be nil")

	// A freshly built client hands back its token without going back to the API.
	token, err := fronteggClient.Token(context.Background())
	require.NoError(t, err)
	require.NotEmpty(t, token, "Token should be available from the Frontegg client")
}

func generateValidJWTToken() string {
	// Create a JWT token with the correct format
	token := jwt.New(jwt.SigningMethodHS256)
	claims := token.Claims.(jwt.MapClaims)
	claims["email"] = "test@example.com"             // Add relevant claims
	claims["exp"] = time.Now().Add(time.Hour).Unix() // Set expiration time

	// Sign the token with a secret key (you can use a random key for testing)
	secretKey := []byte("your-secret-key")
	tokenString, _ := token.SignedString(secretKey)

	return tokenString
}

func TestFronteggClientTokenWithoutCredentials(t *testing.T) {
	// A client assembled without credentials cannot mint a token, and must say so
	// rather than panicking.
	fronteggClient := &FronteggClient{Endpoint: "http://mockedendpoint"}

	_, err := fronteggClient.Token(context.Background())
	require.Error(t, err, "A client with no credentials should report an error")
}

func TestParseAppPassword(t *testing.T) {
	validPassword := "mzp_" + strings.Repeat("a", 64)
	clientId, secretKey, err := parseAppPassword(validPassword)
	require.NoError(t, err, "Parsing valid password should not result in an error")
	require.Equal(t, "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa", clientId, "Client ID should match")
	require.Equal(t, "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa", secretKey, "Secret Key should match")

	invalidPassword := "invalid_password"
	_, _, err = parseAppPassword(invalidPassword)
	require.Error(t, err, "Parsing invalid password should result in an error")

	// Anything other than exactly 64 hex characters is rejected rather than
	// truncated into a malformed secret.
	_, _, err = parseAppPassword("mzp_" + strings.Repeat("a", 65))
	require.Error(t, err, "Parsing an over-long password should result in an error")
}

// trackedBody records whether the response body was closed.
type trackedBody struct {
	io.Reader
	closed bool
}

func (b *trackedBody) Close() error {
	b.closed = true
	return nil
}

type stubTransport struct {
	body *trackedBody
}

func (t *stubTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	return &http.Response{
		StatusCode: http.StatusInternalServerError,
		Body:       t.body,
		Header:     make(http.Header),
	}, nil
}

func TestFronteggRequestClosesBodyOnError(t *testing.T) {
	body := &trackedBody{Reader: strings.NewReader("boom")}
	client := &FronteggClient{HTTPClient: &http.Client{Transport: &stubTransport{body: body}}}

	resp, err := FronteggRequest(context.Background(), client, "GET", "https://example.invalid/x", nil)
	require.Error(t, err, "A 500 response should surface as an error")
	require.Nil(t, resp, "No response should be returned alongside the error")
	require.True(t, body.closed, "The response body should be closed when the caller cannot")
}
