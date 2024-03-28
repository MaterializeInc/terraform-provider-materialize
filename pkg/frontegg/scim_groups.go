package frontegg

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/MaterializeInc/terraform-provider-materialize/pkg/clients"
)

const (
	SCIMGroupsApiPathV1 = "/frontegg/identity/resources/groups/v1"
)

// GroupCreateParams represents the parameters for creating a new group.
type GroupCreateParams struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Color       string `json:"color,omitempty"`
	Metadata    string `json:"metadata,omitempty"`
}

// GroupUpdateParams represents the parameters for updating an existing group.
type GroupUpdateParams struct {
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
	Color       string `json:"color,omitempty"`
	Metadata    string `json:"metadata,omitempty"`
}

// ScimGroup represents the structure of a group in the response.
type ScimGroup struct {
	ID          string     `json:"id"`
	Name        string     `json:"name"`
	Description string     `json:"description"`
	Metadata    string     `json:"metadata"`
	Roles       []ScimRole `json:"roles"`
	Users       []ScimUser `json:"users"`
	ManagedBy   string     `json:"managedBy"`
}

// ScimRole represents the structure of a role within a group.
type ScimRole struct {
	ID          string `json:"id"`
	Key         string `json:"key"`
	Name        string `json:"name"`
	Description string `json:"description"`
	IsDefault   bool   `json:"is_default"`
}

// ScimUser represents the structure of a user within a group.
type ScimUser struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

// SCIMGroupsResponse represents the overall structure of the response from the SCIM groups API.
type SCIMGroupsResponse struct {
	Groups []ScimGroup `json:"groups"`
}

// AddRolesToGroupParams represents the parameters for adding roles to a group.
type AddRolesToGroupParams struct {
	RoleIds []string `json:"roleIds"`
}

// CreateSCIMGroup creates a new group in Frontegg.
func CreateSCIMGroup(ctx context.Context, client *clients.FronteggClient, params GroupCreateParams) (*ScimGroup, error) {
	endpoint := fmt.Sprintf("%s%s", client.Endpoint, SCIMGroupsApiPathV1)
	payloadBytes, err := json.Marshal(params)
	if err != nil {
		return nil, err
	}
	body := bytes.NewReader(payloadBytes)

	req, err := http.NewRequestWithContext(ctx, "POST", endpoint, body)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Authorization", "Bearer "+client.Token)
	req.Header.Add("Content-Type", "application/json")

	resp, err := client.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		var sb strings.Builder
		_, err = io.Copy(&sb, resp.Body)
		if err != nil {
			return nil, err
		}
		return nil, fmt.Errorf("error creating SCIM group: status %d, response: %s", resp.StatusCode, sb.String())
	}

	var group ScimGroup
	if err := json.NewDecoder(resp.Body).Decode(&group); err != nil {
		return nil, err
	}

	return &group, nil
}

// UpdateSCIMGroup updates an existing group in Frontegg.
func UpdateSCIMGroup(ctx context.Context, client *clients.FronteggClient, groupID string, params GroupUpdateParams) (*ScimGroup, error) {
	endpoint := fmt.Sprintf("%s%s/%s", client.Endpoint, SCIMGroupsApiPathV1, groupID)
	payloadBytes, err := json.Marshal(params)
	if err != nil {
		return nil, err
	}
	body := bytes.NewReader(payloadBytes)

	req, err := http.NewRequestWithContext(ctx, "PATCH", endpoint, body)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Authorization", "Bearer "+client.Token)
	req.Header.Add("Content-Type", "application/json")

	resp, err := client.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var sb strings.Builder
		_, err = io.Copy(&sb, resp.Body)
		if err != nil {
			return nil, err
		}
		return nil, fmt.Errorf("error updating SCIM group: status %d, response: %s", resp.StatusCode, sb.String())
	}

	var group ScimGroup
	if err := json.NewDecoder(resp.Body).Decode(&group); err != nil {
		return nil, err
	}

	return &group, nil
}

// DeleteSCIMGroup deletes an existing group in Frontegg.
func DeleteSCIMGroup(ctx context.Context, client *clients.FronteggClient, groupID string) error {
	endpoint := fmt.Sprintf("%s%s/%s", client.Endpoint, SCIMGroupsApiPathV1, groupID)

	req, err := http.NewRequestWithContext(ctx, "DELETE", endpoint, nil)
	if err != nil {
		return err
	}

	req.Header.Add("Authorization", "Bearer "+client.Token)

	resp, err := client.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var sb strings.Builder
		_, err = io.Copy(&sb, resp.Body)
		if err != nil {
			return err
		}
		return fmt.Errorf("error deleting SCIM group: status %d, response: %s", resp.StatusCode, sb.String())
	}

	return nil
}

// GetSCIMGroupByID fetches a single SCIM group by its ID.
func GetSCIMGroupByID(ctx context.Context, client *clients.FronteggClient, groupID string) (*ScimGroup, error) {
	endpoint := fmt.Sprintf("%s%s/%s", client.Endpoint, SCIMGroupsApiPathV1, groupID)

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
		var sb strings.Builder
		_, err = io.Copy(&sb, resp.Body)
		if err != nil {
			return nil, err
		}
		return nil, fmt.Errorf("error fetching SCIM group: status %d, response: %s", resp.StatusCode, sb.String())
	}

	var group ScimGroup
	if err := json.NewDecoder(resp.Body).Decode(&group); err != nil {
		return nil, err
	}

	return &group, nil
}

// FetchSCIMGroups fetches all SCIM groups.
func FetchSCIMGroups(ctx context.Context, client *clients.FronteggClient) ([]ScimGroup, error) {
	endpoint := fmt.Sprintf("%s%s?_groupsRelations=rolesAndUsers", client.Endpoint, SCIMGroupsApiPathV1)
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
		var sb strings.Builder
		_, err = io.Copy(&sb, resp.Body)
		if err != nil {
			return nil, err
		}
		return nil, fmt.Errorf("error reading SCIM groups: status %d, response: %s", resp.StatusCode, sb.String())
	}

	var groupsResponse SCIMGroupsResponse
	if err := json.NewDecoder(resp.Body).Decode(&groupsResponse); err != nil {
		return nil, err
	}

	return groupsResponse.Groups, nil
}

// AddRolesToGroup adds roles to an existing group in Frontegg.
func AddRolesToGroup(ctx context.Context, client *clients.FronteggClient, groupId string, roleIds []string) error {
	endpoint := fmt.Sprintf("%s%s/%s/roles", client.Endpoint, SCIMGroupsApiPathV1, groupId)
	params := AddRolesToGroupParams{
		RoleIds: roleIds,
	}
	payloadBytes, err := json.Marshal(params)
	if err != nil {
		return err
	}
	body := bytes.NewReader(payloadBytes)

	req, err := http.NewRequestWithContext(ctx, "POST", endpoint, body)
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
		var sb strings.Builder
		_, err = io.Copy(&sb, resp.Body)
		if err != nil {
			return err
		}
		return fmt.Errorf("error adding roles to SCIM group: status %d, response: %s", resp.StatusCode, sb.String())
	}

	return nil
}

// RemoveRolesFromGroup removes roles from an existing group in Frontegg.
func RemoveRolesFromGroup(ctx context.Context, client *clients.FronteggClient, groupId string, roleIds []string) error {
	endpoint := fmt.Sprintf("%s%s/%s/roles", client.Endpoint, SCIMGroupsApiPathV1, groupId)
	params := AddRolesToGroupParams{
		RoleIds: roleIds,
	}
	payloadBytes, err := json.Marshal(params)
	if err != nil {
		return err
	}
	body := bytes.NewReader(payloadBytes)

	req, err := http.NewRequestWithContext(ctx, "DELETE", endpoint, body)
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

	if resp.StatusCode != http.StatusOK {
		var sb strings.Builder
		_, err = io.Copy(&sb, resp.Body)
		if err != nil {
			return err
		}
		return fmt.Errorf("error removing roles from SCIM group: status %d, response: %s", resp.StatusCode, sb.String())
	}

	return nil
}

// AddUsersToGroup adds users to an existing group in Frontegg.
func AddUsersToGroup(ctx context.Context, client *clients.FronteggClient, groupId string, userIds []string) error {
	endpoint := fmt.Sprintf("%s%s/%s/users", client.Endpoint, SCIMGroupsApiPathV1, groupId)
	params := struct {
		UserIds []string `json:"userIds"`
	}{
		UserIds: userIds,
	}
	payloadBytes, err := json.Marshal(params)
	if err != nil {
		return err
	}
	body := bytes.NewReader(payloadBytes)

	req, err := http.NewRequestWithContext(ctx, "POST", endpoint, body)
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
		var sb strings.Builder
		_, err = io.Copy(&sb, resp.Body)
		if err != nil {
			return err
		}
		return fmt.Errorf("error adding users to SCIM group: status %d, response: %s", resp.StatusCode, sb.String())
	}

	return nil
}

// RemoveUsersFromGroup removes users from an existing group in Frontegg.
func RemoveUsersFromGroup(ctx context.Context, client *clients.FronteggClient, groupId string, userIds []string) error {
	endpoint := fmt.Sprintf("%s%s/%s/users", client.Endpoint, SCIMGroupsApiPathV1, groupId)
	params := struct {
		UserIds []string `json:"userIds"`
	}{
		UserIds: userIds,
	}
	payloadBytes, err := json.Marshal(params)
	if err != nil {
		return err
	}
	body := bytes.NewReader(payloadBytes)

	req, err := http.NewRequestWithContext(ctx, "DELETE", endpoint, body)
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

	if resp.StatusCode != http.StatusOK {
		var sb strings.Builder
		_, err = io.Copy(&sb, resp.Body)
		if err != nil {
			return err
		}
		return fmt.Errorf("error removing users from SCIM group: status %d, response: %s", resp.StatusCode, sb.String())
	}

	return nil
}

// Helper function to flatten the roles data
func FlattenRoles(roles []ScimRole) []interface{} {
	var flattenedRoles []interface{}
	for _, role := range roles {
		flattenedRole := map[string]interface{}{
			"id":          role.ID,
			"key":         role.Key,
			"name":        role.Name,
			"description": role.Description,
			"is_default":  role.IsDefault,
		}
		flattenedRoles = append(flattenedRoles, flattenedRole)
	}
	return flattenedRoles
}

// Helper function to flatten the users data
func FlattenUsers(users []ScimUser) []interface{} {
	var flattenedUsers []interface{}
	for _, user := range users {
		flattenedUser := map[string]interface{}{
			"id":    user.ID,
			"name":  user.Name,
			"email": user.Email,
		}
		flattenedUsers = append(flattenedUsers, flattenedUser)
	}
	return flattenedUsers
}

// Helper function to flatten the groups data
func FlattenScimGroups(groups []ScimGroup) []interface{} {
	var flattenedGroups []interface{}
	for _, group := range groups {
		flattenedGroup := map[string]interface{}{
			"id":          group.ID,
			"name":        group.Name,
			"description": group.Description,
			"metadata":    group.Metadata,
			"roles":       FlattenRoles(group.Roles),
			"users":       FlattenUsers(group.Users),
			"managed_by":  group.ManagedBy,
		}
		flattenedGroups = append(flattenedGroups, flattenedGroup)
	}
	return flattenedGroups
}
