package utils

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
func ListRoles(ctx context.Context, client *clients.FronteggClient) (map[string]string, error) {
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
