package clients

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// FronteggClient struct to encapsulate the http.Client with additional properties
type FronteggClient struct {
	HTTPClient  *http.Client
	Token       string
	Email       string
	Endpoint    string
	TokenExpiry time.Time
	Password    string
}

// NewFronteggClient function for initializing a new Frontegg client with an auth token
func NewFronteggClient(ctx context.Context, password, endpoint string) (*FronteggClient, error) {
	token, email, tokenExpiry, err := getToken(ctx, password, endpoint)
	if err != nil {
		return nil, fmt.Errorf("failed to get token: %v", err)
	}

	transport := &tokenTransport{
		Token:     token,
		Transport: http.DefaultTransport,
	}

	client := &http.Client{Transport: transport}

	return &FronteggClient{
		HTTPClient:  client,
		Token:       token,
		Email:       email,
		Endpoint:    endpoint,
		TokenExpiry: tokenExpiry.Add(-time.Duration(0.5*float64(time.Until(tokenExpiry).Nanoseconds())) * time.Nanosecond),
		Password:    password,
	}, nil
}

// tokenTransport struct to add the Authorization header to each request
type tokenTransport struct {
	Token     string
	Transport http.RoundTripper
}

// RoundTrip method to execute the request with the token
func (t *tokenTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	// Deep copy the request to ensure it's safe to modify
	req2 := cloneRequest(req)
	req2.Header.Set("Authorization", "Bearer "+t.Token)
	return t.Transport.RoundTrip(req2)
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

	resp, err := http.DefaultClient.Do(req)
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
		return "", "", time.Time{}, errors.New("email claim not found in token")
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

// Get the token from the FronteggClient
func (c *FronteggClient) GetToken() (string, error) {
	return c.Token, nil
}

// Get the email from the FronteggClient
func (c *FronteggClient) GetEmail() (string, error) {
	return c.Email, nil
}

// Get the endpoint from the FronteggClient
func (c *FronteggClient) GetEndpoint() (string, error) {
	return c.Endpoint, nil
}

// Get the token expiry from the FronteggClient
func (c *FronteggClient) GetTokenExpiry() (time.Time, error) {
	return c.TokenExpiry, nil
}

// Get the password from the FronteggClient
func (c *FronteggClient) GetPassword() (string, error) {
	return c.Password, nil
}

func cloneRequest(r *http.Request) *http.Request {
	// Deep copy the request
	r2 := new(http.Request)
	*r2 = *r
	// Deep copy the URL
	r2.URL = new(url.URL)
	*r2.URL = *r.URL
	// Deep copy the Header
	r2.Header = make(http.Header, len(r.Header))
	for k, s := range r.Header {
		r2.Header[k] = append([]string(nil), s...)
	}
	return r2
}

func parseAppPassword(password string) (string, string, error) {
	strippedPassword := strings.TrimPrefix(password, "mzp_")
	var clientId, secretKey string

	re := regexp.MustCompile("[^0-9a-fA-F]")
	filteredChars := re.ReplaceAllString(strippedPassword, "")

	if len(filteredChars) < 64 {
		return "", "", fmt.Errorf("invalid app password length: %d", len(filteredChars))
	}

	clientId = formatDashlessUuid(filteredChars[0:32])
	secretKey = formatDashlessUuid(filteredChars[32:])

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

func (c *FronteggClient) NeedsTokenRefresh() bool {
	return time.Now().After(c.TokenExpiry)
}

func (c *FronteggClient) RefreshToken() error {
	log.Printf("[DEBUG] Refreshing Frontegg: %v\n", c)

	token, email, tokenExpiry, err := getToken(context.Background(), c.Password, c.Endpoint)
	if err != nil {
		return fmt.Errorf("failed to get token: %v", err)
	}

	transport := &tokenTransport{
		Token:     token,
		Transport: http.DefaultTransport,
	}

	client := &http.Client{Transport: transport}

	c.HTTPClient = client
	c.Token = token
	c.Email = email
	c.TokenExpiry = tokenExpiry.Add(-time.Duration(0.5*float64(time.Until(tokenExpiry).Nanoseconds())) * time.Nanosecond)

	return nil
}
