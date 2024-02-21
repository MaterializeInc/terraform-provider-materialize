package frontegg

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/MaterializeInc/terraform-provider-materialize/pkg/clients"
)

type AppPasswordResponse struct {
	ClientID    string    `json:"clientId"`
	Description string    `json:"description"`
	Owner       string    `json:"owner"`
	CreatedAt   time.Time `json:"created_at"`
	Secret      string    `json:"secret"`
}

const (
	ApiTokenPath = "/identity/resources/users/api-tokens/v1"
)

// ListAppPasswords fetches a list of app passwords from the API.
func ListAppPasswords(ctx context.Context, client *clients.FronteggClient) ([]AppPasswordResponse, error) {
	var passwords []AppPasswordResponse

	// Construct the request URL
	url := GetAppPasswordApiEndpoint(client, ApiTokenPath)

	// Create and send the HTTP request
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request failed: %w", err)
	}
	req.Header.Add("Authorization", "Bearer "+client.Token)

	// Execute the request
	resp, err := client.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("executing request failed: %w", err)
	}
	defer resp.Body.Close()

	// Check the response code
	if resp.StatusCode != http.StatusOK {
		return nil, clients.HandleApiError(resp)
	}

	// Decode the response body
	responseBody, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		return nil, fmt.Errorf("reading response body failed: %w", readErr)
	}

	if err := json.Unmarshal(responseBody, &passwords); err != nil {
		return nil, fmt.Errorf("decoding response failed: %w", err)
	}

	return passwords, nil
}

// Helper function to construct the full API endpoint for app passwords
func GetAppPasswordApiEndpoint(client *clients.FronteggClient, resourcePath string, resourceID ...string) string {
	if len(resourceID) > 0 {
		return fmt.Sprintf("%s%s/%s", client.Endpoint, resourcePath, resourceID[0])
	}
	return fmt.Sprintf("%s%s", client.Endpoint, resourcePath)
}
