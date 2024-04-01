package frontegg

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/MaterializeInc/terraform-provider-materialize/pkg/clients"
)

const (
	SSOGroupMappingApiPathV1 = "/frontegg/team/resources/sso/v1/configurations"
)

// GroupMapping represents the structure for SSO group role mapping.
type GroupMapping struct {
	ID          string   `json:"id"`
	Group       string   `json:"group"`
	Enabled     bool     `json:"enabled"`
	RoleIds     []string `json:"roleIds"`
	SsoConfigId string   `json:"-"`
}

// CreateSSOGroupMapping creates a new SSO group role mapping.
func CreateSSOGroupMapping(ctx context.Context, client *clients.FronteggClient, ssoConfigID, group string, roleIDs []string) (*GroupMapping, error) {
	payload := map[string]interface{}{
		"group":   group,
		"roleIds": roleIDs,
	}

	requestBody, err := jsonEncode(payload)
	if err != nil {
		return nil, err
	}

	endpoint := fmt.Sprintf("%s%s/%s/groups", client.Endpoint, SSOGroupMappingApiPathV1, ssoConfigID)
	resp, err := doRequest(ctx, client, "POST", endpoint, requestBody)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result GroupMapping
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return &result, nil
}

// GetSSOGroupMappings retrieves all SSO group role mappings for a specific SSO configuration.
func GetSSOGroupMappings(ctx context.Context, client *clients.FronteggClient, ssoConfigID string) (*[]GroupMapping, error) {
	endpoint := fmt.Sprintf("%s%s/%s/groups", client.Endpoint, SSOGroupMappingApiPathV1, ssoConfigID)
	resp, err := doRequest(ctx, client, "GET", endpoint, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var groups []GroupMapping
	if err := json.NewDecoder(resp.Body).Decode(&groups); err != nil {
		return nil, err
	}

	if groups == nil {
		groups = []GroupMapping{}
	}

	return &groups, nil
}

// FetchSSOGroupMapping retrieves a specific SSO group role mapping.
func FetchSSOGroupMapping(ctx context.Context, client *clients.FronteggClient, ssoConfigID, groupID string) (*GroupMapping, error) {
	// Call the FetchSSOGroupMappings function to get all group mappings.
	groups, err := GetSSOGroupMappings(ctx, client, ssoConfigID)
	if err != nil {
		return nil, err
	}

	// Find the group mapping with the specified group ID.
	for _, group := range *groups {
		log.Printf("group.ID: %s, groupID: %s", group.ID, groupID)
		if group.ID == groupID {
			return &group, nil
		}
	}

	return nil, fmt.Errorf("group mapping not found: %s", groupID)
}

// UpdateSSOGroupMapping updates an existing SSO group role mapping.
func UpdateSSOGroupMapping(ctx context.Context, client *clients.FronteggClient, ssoConfigID, groupID, group string, roleIDs []string) (*GroupMapping, error) {
	payload := map[string]interface{}{
		"group":   group,
		"roleIds": roleIDs,
	}

	requestBody, err := jsonEncode(payload)
	if err != nil {
		return nil, err
	}

	endpoint := fmt.Sprintf("%s%s/%s/groups/%s", client.Endpoint, SSOGroupMappingApiPathV1, ssoConfigID, groupID)
	resp, err := doRequest(ctx, client, "PATCH", endpoint, requestBody)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result GroupMapping
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return &result, nil
}

// DeleteSSOGroupMapping deletes an existing SSO group role mapping.
func DeleteSSOGroupMapping(ctx context.Context, client *clients.FronteggClient, ssoConfigID, groupID string) error {
	endpoint := fmt.Sprintf("%s%s/%s/groups/%s", client.Endpoint, SSOGroupMappingApiPathV1, ssoConfigID, groupID)
	resp, err := doRequest(ctx, client, "DELETE", endpoint, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("error deleting SSO group mapping: status %d", resp.StatusCode)
	}

	return nil
}
