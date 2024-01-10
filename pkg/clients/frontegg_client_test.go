package clients

import (
	"context"
	"encoding/json"
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

	// The token should be set correctly in the Frontegg client
	require.Equal(t, "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJlbWFpbCI6Im16X3N5c3RlbSIsImV4cCI6MTcwMDAwMDAwMH0.c2lnbmF0dXJl", fronteggClient.Token, "Token should be set correctly in the Frontegg client")
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

	// The token should be initially set correctly in the Frontegg client
	require.NotEmpty(t, fronteggClient.Token, "Token should be set correctly in the Frontegg client")

	// Verify that the client does not detect the need for token refresh immediately
	require.NoError(t, fronteggClient.NeedsTokenRefresh(), "Token should not be considered expired")
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

func TestFronteggClient_NeedsTokenRefresh(t *testing.T) {
	// Create a Frontegg client with an expired token
	fronteggClient := &FronteggClient{
		Token:       "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJlbWFpbCI6Im16X3N5c3RlbSIsImV4cCI6MTYwMDAwMDAwMH0.c2lnbmF0dXJl",
		Email:       "test@example.com",
		Endpoint:    "http://mockedendpoint",
		TokenExpiry: time.Now().Add(-time.Hour), // Expired token
		Password:    "mzp_" + strings.Repeat("a", 64),
	}

	// Verify that the client correctly detects the need for token refresh
	require.Error(t, fronteggClient.NeedsTokenRefresh(), "Token should be considered expired and require refresh")
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
}
