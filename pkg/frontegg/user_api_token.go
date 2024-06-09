package frontegg

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/MaterializeInc/terraform-provider-materialize/pkg/clients"
)

type UserApiTokenRequest struct {
	Description string `json:"description"`
}

type UserApiTokenResponse struct {
	ClientID    string    `json:"clientId"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	Secret      string    `json:"secret"`
}

const (
	UserApiTokenPath = "/identity/resources/users/api-tokens/v1"
)

func CreateUserApiToken(ctx context.Context, client *clients.FronteggClient, request UserApiTokenRequest) (UserApiTokenResponse, error) {
	var response UserApiTokenResponse

	requestBody, err := jsonEncode(request)
	if err != nil {
		return response, err
	}

	endpoint := GetUserApiTokenApiEndpoint(client, UserApiTokenPath)
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

func DeleteUserApiToken(ctx context.Context, client *clients.FronteggClient, id string) error {
	endpoint := GetUserApiTokenApiEndpoint(client, UserApiTokenPath, id)
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

// ListUserApiTokens fetches a list of user API tokens from the API.
func ListUserApiTokens(ctx context.Context, client *clients.FronteggClient) ([]UserApiTokenResponse, error) {
	var tokens []UserApiTokenResponse
	endpoint := GetUserApiTokenApiEndpoint(client, UserApiTokenPath)

	resp, err := doRequest(ctx, client, "GET", endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("request to list user API tokens failed: %w", err)
	}
	defer resp.Body.Close()

	if err := json.NewDecoder(resp.Body).Decode(&tokens); err != nil {
		return nil, fmt.Errorf("decoding user API tokens failed: %w", err)
	}

	return tokens, nil
}

// Helper function to construct the full API endpoint for a user API token.
func GetUserApiTokenApiEndpoint(client *clients.FronteggClient, resourcePath string, resourceID ...string) string {
	if len(resourceID) > 0 {
		return fmt.Sprintf("%s%s/%s", client.Endpoint, resourcePath, resourceID[0])
	}
	return fmt.Sprintf("%s%s", client.Endpoint, resourcePath)
}
