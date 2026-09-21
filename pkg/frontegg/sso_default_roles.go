package frontegg

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
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
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Key         string   `json:"key"`
	Description string   `json:"description"`
	TenantID    string   `json:"tenantId"`
	Permissions []string `json:"permissions"`
}

// RoleName preserves custom names and the provider's built-in role aliases.
func RoleName(name string) string {
	switch name {
	case "Organization Admin":
		return "Admin"
	case "Organization Member":
		return "Member"
	default:
		return name
	}
}

func FetchFronteggRoles(ctx context.Context, client *clients.FronteggClient) ([]FronteggRole, error) {
	var roles []FronteggRole
	for page := 0; ; page++ {
		endpoint := fmt.Sprintf("%s%s?_sortBy=key&_order=ASC&_limit=2000&_offset=%d", client.Endpoint, SSORolesApiPathV2, page)
		resp, err := doRequest(ctx, client, "GET", endpoint, nil)
		if err != nil {
			return nil, err
		}
		var result FronteggRolesResponse
		err = json.NewDecoder(resp.Body).Decode(&result)
		resp.Body.Close()
		if err != nil {
			return nil, fmt.Errorf("error decoding roles: %w", err)
		}
		roles = append(roles, result.Items...)
		if page+1 >= result.Metadata.TotalPages {
			return roles, nil
		}
	}
}

// ListFronteggRoles includes tenant roles as well as built-in organization roles.
func ListFronteggRoles(ctx context.Context, client *clients.FronteggClient) (map[string]string, error) {
	roles, err := FetchFronteggRoles(ctx, client)
	if err != nil {
		return nil, err
	}
	roleMap := make(map[string]string)
	for _, role := range roles {
		name := RoleName(role.Name)
		if id, ok := roleMap[name]; ok && id != role.ID {
			return nil, fmt.Errorf("ambiguous organization role name: %s", name)
		}
		roleMap[name] = role.ID
	}
	return roleMap, nil
}

// SetSSODefaultRoles sets the default roles for an SSO configuration.
func SetSSODefaultRoles(ctx context.Context, client *clients.FronteggClient, configID string, roleIDs []string) error {
	payload := RoleIDs{RoleIds: roleIDs}
	requestBody, err := jsonEncode(payload)
	if err != nil {
		return err
	}

	endpoint := fmt.Sprintf(client.Endpoint+SSORolesApiPathV1, configID)
	resp, err := doRequest(ctx, client, "PUT", endpoint, requestBody)
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

	resp, err := doRequest(ctx, client, "GET", endpoint, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

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
