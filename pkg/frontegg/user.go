package frontegg

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/MaterializeInc/terraform-provider-materialize/pkg/clients"
)

const (
	UsersApiPathV1 = "/identity/resources/users/v1"
	UsersApiPathV2 = "/identity/resources/users/v2"
)

// UserRequest represents the request payload for creating or updating a user.
type UserRequest struct {
	Email           string   `json:"email"`
	RoleIDs         []string `json:"roleIds"`
	SkipInviteEmail bool     `json:"skipInviteEmail"`
}

type UserRole struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

// UserResponse represents the structure of a user response from Frontegg APIs.
type UserResponse struct {
	ID                string     `json:"id"`
	Email             string     `json:"email"`
	ProfilePictureURL string     `json:"profilePictureUrl"`
	Verified          bool       `json:"verified"`
	Metadata          string     `json:"metadata"`
	Provider          string     `json:"provider"`
	Roles             []UserRole `json:"roles"`
}

// CreateUser creates a new user in Frontegg.
func CreateUser(ctx context.Context, client *clients.FronteggClient, userRequest UserRequest) (UserResponse, error) {
	var userResponse UserResponse

	requestBody, err := jsonEncode(userRequest)
	if err != nil {
		return userResponse, err
	}

	endpoint := fmt.Sprintf("%s%s", client.Endpoint, UsersApiPathV2)
	resp, err := doRequest(ctx, client, "POST", endpoint, requestBody)
	if err != nil {
		return userResponse, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return userResponse, clients.HandleApiError(resp)
	}

	if err := json.NewDecoder(resp.Body).Decode(&userResponse); err != nil {
		return userResponse, fmt.Errorf("decoding response failed: %w", err)
	}

	return userResponse, nil
}

// ReadUser retrieves a user's details from Frontegg.
func ReadUser(ctx context.Context, client *clients.FronteggClient, userID string) (UserResponse, error) {
	var userResponse UserResponse

	endpoint := fmt.Sprintf("%s%s/%s", client.Endpoint, UsersApiPathV1, userID)
	resp, err := doRequest(ctx, client, "GET", endpoint, nil)
	if err != nil {
		return userResponse, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return userResponse, clients.HandleApiError(resp)
	}

	if err := json.NewDecoder(resp.Body).Decode(&userResponse); err != nil {
		return userResponse, fmt.Errorf("decoding response failed: %w", err)
	}

	return userResponse, nil
}

// DeleteUser deletes a user from Frontegg.
func DeleteUser(ctx context.Context, client *clients.FronteggClient, userID string) error {
	endpoint := fmt.Sprintf("%s%s/%s", client.Endpoint, UsersApiPathV1, userID)
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
