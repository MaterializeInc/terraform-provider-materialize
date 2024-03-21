package frontegg

import (
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
