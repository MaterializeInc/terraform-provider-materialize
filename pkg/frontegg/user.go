package frontegg

import (
	"bytes"
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
	Email   string   `json:"email"`
	RoleIDs []string `json:"roleIds"`
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

	requestBody, err := json.Marshal(userRequest)
	if err != nil {
		return userResponse, err
	}

	url := fmt.Sprintf("%s%s", client.Endpoint, UsersApiPathV2)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(requestBody))
	if err != nil {
		return userResponse, fmt.Errorf("creating request failed: %w", err)
	}
	req.Header.Add("Authorization", "Bearer "+client.Token)
	req.Header.Add("Content-Type", "application/json")

	resp, err := client.HTTPClient.Do(req)
	if err != nil {
		return userResponse, fmt.Errorf("executing request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return userResponse, clients.HandleApiError(resp)
	}

	if err = json.NewDecoder(resp.Body).Decode(&userResponse); err != nil {
		return userResponse, fmt.Errorf("decoding response failed: %w", err)
	}

	return userResponse, nil
}

// ReadUser retrieves a user's details from Frontegg.
func ReadUser(ctx context.Context, client *clients.FronteggClient, userID string) (UserResponse, error) {
	var userResponse UserResponse

	url := fmt.Sprintf("%s%s/%s", client.Endpoint, UsersApiPathV1, userID)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return userResponse, fmt.Errorf("creating request failed: %w", err)
	}
	req.Header.Add("Authorization", "Bearer "+client.Token)

	resp, err := client.HTTPClient.Do(req)
	if err != nil {
		return userResponse, fmt.Errorf("executing request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return userResponse, clients.HandleApiError(resp)
	}

	if err = json.NewDecoder(resp.Body).Decode(&userResponse); err != nil {
		return userResponse, fmt.Errorf("decoding response failed: %w", err)
	}

	return userResponse, nil
}

// DeleteUser deletes a user from Frontegg.
func DeleteUser(ctx context.Context, client *clients.FronteggClient, userID string) error {
	url := fmt.Sprintf("%s%s/%s", client.Endpoint, UsersApiPathV1, userID)
	req, err := http.NewRequestWithContext(ctx, "DELETE", url, nil)
	if err != nil {
		return fmt.Errorf("creating request failed: %w", err)
	}
	req.Header.Add("Authorization", "Bearer "+client.Token)

	resp, err := client.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("executing request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return clients.HandleApiError(resp)
	}

	return nil
}
