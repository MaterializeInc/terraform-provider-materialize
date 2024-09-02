package frontegg

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/MaterializeInc/terraform-provider-materialize/pkg/clients"
)

const (
	UsersApiPathV1       = "/identity/resources/users/v1"
	UsersApiPathV2       = "/identity/resources/users/v2"
	UsersApiPathV3       = "/identity/resources/users/v3"
	TeamMembersApiPathV1 = "/frontegg/team/resources/members/v1"
)

// UserRequest represents the request payload for creating or updating a user.
type UserRequest struct {
	Email           string   `json:"email"`
	Password        string   `json:"password,omitempty"`
	RoleIDs         []string `json:"roleIds"`
	SkipInviteEmail bool     `json:"skipInviteEmail"`
}

type UserRole struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

type QueryUsersParams struct {
	IncludeSubTenants bool   `url:"_includeSubTenants,omitempty"`
	Limit             int    `url:"_limit,omitempty"`
	Offset            int    `url:"_offset,omitempty"`
	Email             string `url:"_email,omitempty"`
	TenantID          string `url:"_tenantId,omitempty"`
	IDs               string `url:"ids,omitempty"`
	SortBy            string `url:"_sortBy,omitempty"`
	Order             string `url:"_order,omitempty"`
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

	if err := json.NewDecoder(resp.Body).Decode(&userResponse); err != nil {
		return userResponse, fmt.Errorf("decoding response failed: %w", err)
	}

	return userResponse, nil
}

func GetUsers(ctx context.Context, client *clients.FronteggClient, params QueryUsersParams) ([]UserResponse, error) {
	var response struct {
		Items    []UserResponse `json:"items"`
		Metadata struct {
			TotalItems int `json:"totalItems"`
		} `json:"_metadata"`
	}

	// Construct the query string
	values := url.Values{}
	if params.IncludeSubTenants {
		values.Set("_includeSubTenants", "true")
	}
	if params.Limit > 0 {
		values.Set("_limit", fmt.Sprintf("%d", params.Limit))
	}
	if params.Offset > 0 {
		values.Set("_offset", fmt.Sprintf("%d", params.Offset))
	}
	if params.Email != "" {
		values.Set("_email", params.Email)
	}
	if params.IDs != "" {
		values.Set("ids", params.IDs)
	}
	if params.SortBy != "" {
		values.Set("_sortBy", params.SortBy)
	}
	if params.Order != "" {
		values.Set("_order", params.Order)
	}

	endpoint := fmt.Sprintf("%s%s?%s", client.Endpoint, UsersApiPathV3, values.Encode())

	resp, err := doRequest(ctx, client, "GET", endpoint, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response body failed: %w", err)
	}

	resp.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

	if resp.StatusCode != http.StatusOK {
		return nil, clients.HandleApiError(resp)
	}

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("decoding response failed: %w", err)
	}

	if len(response.Items) == 0 {
		return nil, fmt.Errorf("no user found with email: %s", params.Email)
	}

	return response.Items, nil
}

func UpdateUserRoles(ctx context.Context, client *clients.FronteggClient, userID string, email string, roleIDs []string) error {
	payload := struct {
		ID      string   `json:"id"`
		Email   string   `json:"email"`
		RoleIDs []string `json:"roleIds"`
	}{
		ID:      userID,
		Email:   email,
		RoleIDs: roleIDs,
	}

	requestBody, err := jsonEncode(payload)
	if err != nil {
		return err
	}

	endpoint := fmt.Sprintf("%s%s", client.Endpoint, TeamMembersApiPathV1)
	resp, err := doRequest(ctx, client, "PUT", endpoint, requestBody)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return clients.HandleApiError(resp)
	}

	return nil
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
