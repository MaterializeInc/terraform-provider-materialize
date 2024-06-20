package frontegg

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/MaterializeInc/terraform-provider-materialize/pkg/clients"
)

type TenantApiTokenRequest struct {
	Description string            `json:"description"`
	Metadata    map[string]string `json:"metadata"`
	RoleIDs     []string          `json:"roleIds"`
}

type TenantApiTokenResponse struct {
	ClientID        string            `json:"clientId"`
	Description     string            `json:"description"`
	Secret          string            `json:"secret"`
	CreatedByUserId string            `json:"createdByUserId"`
	Metadata        map[string]string `json:"metadata"`
	CreatedAt       time.Time         `json:"createdAt"`
	RoleIDs         []string          `json:"roleIds"`
}

const (
	TenantApiTokenPath = "/identity/resources/tenants/api-tokens/v1"
)

func CreateTenantApiToken(ctx context.Context, client *clients.FronteggClient, request TenantApiTokenRequest) (TenantApiTokenResponse, error) {
	var response TenantApiTokenResponse

	requestBody, err := jsonEncode(request)
	if err != nil {
		return response, err
	}

	endpoint := GetTenantApiTokenApiEndpoint(client, TenantApiTokenPath)
	resp, err := doRequest(ctx, client, "POST", endpoint, requestBody)
	if err != nil {
		return response, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return response, clients.HandleApiError(resp)
	}

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return response, fmt.Errorf("decoding response failed: %w", err)
	}

	return response, nil
}

func DeleteTenantApiToken(ctx context.Context, client *clients.FronteggClient, id string) error {
	endpoint := GetTenantApiTokenApiEndpoint(client, TenantApiTokenPath, id)
	resp, err := doRequest(ctx, client, "DELETE", endpoint, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return clients.HandleApiError(resp)
	}

	return nil
}

// ListTenantApiTokens fetches a list of tenant API tokens from the API.
func ListTenantApiTokens(ctx context.Context, client *clients.FronteggClient) ([]TenantApiTokenResponse, error) {
	var tokens []TenantApiTokenResponse
	endpoint := GetTenantApiTokenApiEndpoint(client, TenantApiTokenPath)

	resp, err := doRequest(ctx, client, "GET", endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("request to list tenant API tokens failed: %w", err)
	}
	defer resp.Body.Close()

	if err := json.NewDecoder(resp.Body).Decode(&tokens); err != nil {
		return nil, fmt.Errorf("decoding tenant API tokens failed: %w", err)
	}

	return tokens, nil
}

// Helper function to construct the full API endpoint for a tenant API token.
func GetTenantApiTokenApiEndpoint(client *clients.FronteggClient, resourcePath string, resourceID ...string) string {
	if len(resourceID) > 0 {
		return fmt.Sprintf("%s%s/%s", client.Endpoint, resourcePath, resourceID[0])
	}
	return fmt.Sprintf("%s%s", client.Endpoint, resourcePath)
}
