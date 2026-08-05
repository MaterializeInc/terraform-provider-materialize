package clients

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// FronteggClient talks to the Frontegg admin API and doubles as the TokenSource
// for every other authorized client.
//
// HTTPClient, Endpoint and Email are fixed once the client is built. The token
// state behind them is mutable and guarded by mu.
type FronteggClient struct {
	HTTPClient *http.Client
	Endpoint   string
	// Email identifies the authenticated principal and is also used as the
	// database user. It is derived from the credentials, so it never changes for
	// the lifetime of the client.
	Email string

	mu          sync.Mutex
	token       string
	tokenExpiry time.Time
	password    string
}

// NewFronteggClient function for initializing a new Frontegg client with an auth token
func NewFronteggClient(ctx context.Context, password, endpoint string) (*FronteggClient, error) {
	// Fetch a token up front: it validates the credentials before Terraform gets
	// going and yields the email used as the database user.
	token, email, tokenExpiry, err := getToken(ctx, password, endpoint)
	if err != nil {
		return nil, fmt.Errorf("failed to get token: %v", err)
	}

	c := &FronteggClient{
		Endpoint:    endpoint,
		Email:       email,
		token:       token,
		tokenExpiry: refreshDeadline(tokenExpiry),
		password:    password,
	}
	c.HTTPClient = newAuthorizedClient(c)

	return c, nil
}

// Token implements TokenSource. It returns the cached token while it is still
// good and otherwise fetches a new one. The lock is deliberately held across the
// fetch so that concurrent callers arriving on an expired token produce a single
// request instead of one each.
func (c *FronteggClient) Token(ctx context.Context) (string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.token != "" && time.Now().Before(c.tokenExpiry) {
		return c.token, nil
	}

	if c.password == "" {
		return "", errors.New("no credentials available to obtain a Frontegg token")
	}

	token, _, tokenExpiry, err := getToken(ctx, c.password, c.Endpoint)
	if err != nil {
		return "", fmt.Errorf("failed to refresh token: %w", err)
	}

	c.token = token
	c.tokenExpiry = refreshDeadline(tokenExpiry)

	return c.token, nil
}

// refreshDeadline returns the point at which a token should be replaced: halfway
// between now and its actual expiry, so a request is never sent with a token
// about to lapse.
func refreshDeadline(expiry time.Time) time.Time {
	return time.Now().Add(time.Until(expiry) / 2)
}

// GetToken function to authenticate with the Frontegg API and retrieve a token
func getToken(ctx context.Context, password string, endpoint string) (string, string, time.Time, error) {
	clientId, secretKey, err := parseAppPassword(password)
	if err != nil {
		return "", "", time.Time{}, err
	}

	adminEndpoint := fmt.Sprintf("%s/identity/resources/auth/v1/api-token", endpoint)

	payload := map[string]string{
		"clientId": clientId,
		"secret":   secretKey,
	}
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return "", "", time.Time{}, err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", adminEndpoint, bytes.NewBuffer(payloadBytes))
	if err != nil {
		return "", "", time.Time{}, err
	}
	req.Header.Add("Content-Type", "application/json")

	// Not the authorized client: this call is what obtains the token.
	resp, err := authClient.Do(req)
	if err != nil {
		return "", "", time.Time{}, err
	}
	defer resp.Body.Close()

	// Read the response body into the 'body' variable
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", "", time.Time{}, err
	}

	if resp.StatusCode != http.StatusOK {
		return "", "", time.Time{}, fmt.Errorf("authentication failed: %s", string(body))
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return "", "", time.Time{}, err
	}

	tokenString, ok := result["accessToken"].(string)
	if !ok {
		return "", "", time.Time{}, errors.New("access token not found in the response")
	}

	// Parse the token without verifying the signature.
	token, _, err := new(jwt.Parser).ParseUnverified(tokenString, jwt.MapClaims{})
	if err != nil {
		return "", "", time.Time{}, fmt.Errorf("error parsing token: %v", err)
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return "", "", time.Time{}, errors.New("invalid token claims")
	}

	email, ok := claims["email"].(string)
	if !ok {
		email = ""
		// If email is not present (service account case), use metadata.user or sub as identifier
		if metadata, hasMetadata := claims["metadata"].(map[string]interface{}); hasMetadata {
			if user, hasUser := metadata["user"].(string); hasUser {
				email = user
			}
		}
		if email == "" {
			if sub, hasSub := claims["sub"].(string); hasSub {
				email = sub
			} else {
				return "", "", time.Time{}, errors.New("neither email nor subject found in token")
			}
		}
	}

	var tokenExpiry time.Time
	if expiresIn, ok := result["expiresIn"].(float64); ok {
		tokenExpiry = time.Now().Add(time.Duration(expiresIn) * time.Second)
	} else {
		// Default expiry time if not provided in the response
		tokenExpiry = time.Now().Add(1 * time.Hour)
	}

	return tokenString, email, tokenExpiry, nil
}

func parseAppPassword(password string) (string, string, error) {
	strippedPassword := strings.TrimPrefix(password, "mzp_")

	re := regexp.MustCompile("[^0-9a-fA-F]")
	filteredChars := re.ReplaceAllString(strippedPassword, "")

	// An app password is two dashless UUIDs, so exactly 64 hex characters.
	// Anything longer used to be truncated into a malformed secret, which
	// surfaced later as an opaque authentication failure.
	if len(filteredChars) != 64 {
		return "", "", fmt.Errorf("invalid app password: expected 64 hexadecimal characters, got %d", len(filteredChars))
	}

	clientId := formatDashlessUuid(filteredChars[0:32])
	secretKey := formatDashlessUuid(filteredChars[32:64])

	return clientId, secretKey, nil
}

func formatDashlessUuid(dashlessUuid string) string {
	parts := []string{
		dashlessUuid[0:8],
		dashlessUuid[8:12],
		dashlessUuid[12:16],
		dashlessUuid[16:20],
		dashlessUuid[20:],
	}
	return strings.Join(parts, "-")
}

// Helper function to handle API errors
func HandleApiError(resp *http.Response) error {
	if resp.StatusCode == http.StatusNotFound {
		return NewFronteggAPIError(resp, "Resource not found")
	}
	return HandleAPIError(resp)
}

// Helper function to perform HTTP requests
func FronteggRequest(ctx context.Context, client *FronteggClient, method, url string, body []byte) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, method, url, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	req.Header.Add("Content-Type", "application/json")

	resp, err := client.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode >= 400 {
		// The caller never sees this response, so read the error out of the body
		// and close it here rather than leaking the connection.
		apiErr := HandleApiError(resp)
		_ = resp.Body.Close()
		return nil, apiErr
	}

	return resp, nil
}

// FronteggAPIError represents a standardized error structure for Frontegg API calls
type FronteggAPIError struct {
	StatusCode int
	Message    string
}

func (e *FronteggAPIError) Error() string {
	return fmt.Sprintf("Frontegg API error (HTTP %d): %s", e.StatusCode, e.Message)
}

// NewFronteggAPIError creates a new FronteggAPIError instance
func NewFronteggAPIError(resp *http.Response, message string) *FronteggAPIError {
	return &FronteggAPIError{
		StatusCode: resp.StatusCode,
		Message:    message,
	}
}

// HandleAPIError processes an HTTP response and returns a FronteggAPIError if applicable
func HandleAPIError(resp *http.Response) error {
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return nil
	}

	responseBody, _ := io.ReadAll(resp.Body)
	message := fmt.Sprintf("unexpected status code: %d - %s", resp.StatusCode, string(responseBody))
	return NewFronteggAPIError(resp, message)
}

// IsNotFoundError checks if the error is a 404 Not Found error
func IsNotFoundError(err error) bool {
	if fronteggErr, ok := err.(*FronteggAPIError); ok {
		return fronteggErr.StatusCode == http.StatusNotFound
	}
	return false
}
