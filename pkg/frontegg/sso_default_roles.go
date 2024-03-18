package frontegg

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/MaterializeInc/terraform-provider-materialize/pkg/clients"
)

const (
	SSORolesApiPathV1 = "/frontegg/team/resources/sso/v1/configurations/%s/roles"
	SSORolesApiPathV2 = "/identity/resources/roles/v2"
)

type RoleIDs struct {
	RoleIds []string `json:"roleIds"`
}

type FronteggRolesResponse struct {
	Items    []FronteggRole `json:"items"`
	Metadata struct {
		TotalItems int `json:"totalItems"`
		TotalPages int `json:"totalPages"`
	} `json:"_metadata"`
}

type FronteggRole struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// ListRoles fetches roles from the Frontegg API and returns a map of role names to their IDs.
func ListSSORoles(ctx context.Context, client *clients.FronteggClient) (map[string]string, error) {
	rolesURL := fmt.Sprintf("%s/identity/resources/roles/v2", client.Endpoint)

	req, err := http.NewRequestWithContext(ctx, "GET", rolesURL, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %v", err)
	}
	req.Header.Add("Authorization", "Bearer "+client.Token)

	resp, err := client.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error executing request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("error fetching roles, status code: %d", resp.StatusCode)
	}

	// Read and reset the response body
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %v", err)
	}
	resp.Body = io.NopCloser(bytes.NewBuffer(responseBody))

	// Decode the JSON response
	var rolesResponse FronteggRolesResponse
	if err := json.NewDecoder(resp.Body).Decode(&rolesResponse); err != nil {
		return nil, fmt.Errorf("error decoding response: %v", err)
	}

	// Create a map of role names to their IDs
	roleMap := make(map[string]string)
	for _, role := range rolesResponse.Items {
		log.Printf("[DEBUG] Role found: %s - %s\n", role.Name, role.ID)
		if role.Name == "Organization Admin" {
			roleMap["Admin"] = role.ID
		} else if role.Name == "Organization Member" {
			roleMap["Member"] = role.ID
		}
	}

	return roleMap, nil
}

// SetSSODefaultRoles sets the default roles for an SSO configuration.
func SetSSODefaultRoles(ctx context.Context, client *clients.FronteggClient, configID string, roleIDs []string) error {
	payload := RoleIDs{RoleIds: roleIDs}
	requestBody, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	endpoint := fmt.Sprintf(client.Endpoint+SSORolesApiPathV1, configID)
	req, err := http.NewRequestWithContext(ctx, "PUT", endpoint, bytes.NewBuffer(requestBody))
	if err != nil {
		return err
	}

	req.Header.Add("Authorization", "Bearer "+client.Token)
	req.Header.Add("Content-Type", "application/json")

	resp, err := client.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		responseData, err := io.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		return fmt.Errorf("error setting SSO default roles: status %d, response: %s", resp.StatusCode, string(responseData))
	}

	return nil
}

// GetSSODefaultRoles retrieves the default roles for an SSO configuration.
func GetSSODefaultRoles(ctx context.Context, client *clients.FronteggClient, configID string) ([]string, error) {
	endpoint := fmt.Sprintf(client.Endpoint+SSORolesApiPathV1, configID)
	req, err := http.NewRequestWithContext(ctx, "GET", endpoint, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Authorization", "Bearer "+client.Token)

	resp, err := client.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		responseData, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		return nil, fmt.Errorf("error reading SSO default roles: status %d, response: %s", resp.StatusCode, string(responseData))
	}

	var rolesResponse RoleIDs
	if err := json.NewDecoder(resp.Body).Decode(&rolesResponse); err != nil {
		return nil, err
	}

	return rolesResponse.RoleIds, nil
}

// ClearSSODefaultRoles clears the default roles for an SSO configuration.
func ClearSSODefaultRoles(ctx context.Context, client *clients.FronteggClient, configID string) error {
	return SetSSODefaultRoles(ctx, client, configID, []string{})
}
