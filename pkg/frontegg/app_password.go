package frontegg

import (
	"context"
	"encoding/json"
	"fmt"
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
	endpoint := GetAppPasswordApiEndpoint(client, ApiTokenPath)

	resp, err := doRequest(ctx, client, "GET", endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("request to list app passwords failed: %w", err)
	}
	defer resp.Body.Close()

	if err := json.NewDecoder(resp.Body).Decode(&passwords); err != nil {
		return nil, fmt.Errorf("decoding app passwords failed: %w", err)
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
