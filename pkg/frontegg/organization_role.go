package frontegg

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/MaterializeInc/terraform-provider-materialize/pkg/clients"
)

type OrganizationRoleCreateParams struct {
	Name          string   `json:"name"`
	Key           string   `json:"key"`
	Description   string   `json:"description"`
	BaseRoleID    string   `json:"baseRoleId"`
	PermissionIDs []string `json:"permissionIds"`
}

func CreateOrganizationRole(ctx context.Context, client *clients.FronteggClient, params OrganizationRoleCreateParams) (*FronteggRole, error) {
	body, err := jsonEncode(params)
	if err != nil {
		return nil, err
	}
	resp, err := doRequest(ctx, client, "POST", client.Endpoint+SSORolesApiPathV2, body)
	if err != nil {
		return nil, err
	}
	var role FronteggRole
	decodeErr := json.NewDecoder(resp.Body).Decode(&role)
	closeErr := resp.Body.Close()
	if decodeErr != nil {
		return nil, fmt.Errorf("error decoding created role: %w", decodeErr)
	}
	if closeErr != nil {
		return nil, fmt.Errorf("error closing created role response: %w", closeErr)
	}
	if role.ID == "" {
		return nil, fmt.Errorf("role creation returned no ID")
	}
	return &role, nil
}

func UpdateOrganizationRole(ctx context.Context, client *clients.FronteggClient, id, description string) error {
	body, err := jsonEncode(map[string]string{"description": description})
	if err != nil {
		return err
	}
	resp, err := doRequest(ctx, client, "PATCH", organizationRoleURL(client, id), body)
	if err != nil {
		return err
	}
	return resp.Body.Close()
}

func SetOrganizationRolePermissions(ctx context.Context, client *clients.FronteggClient, id string, permissions []string) error {
	body, err := jsonEncode(map[string][]string{"permissionIds": permissions})
	if err != nil {
		return err
	}
	resp, err := doRequest(ctx, client, "PUT", organizationRoleURL(client, id)+"/permissions", body)
	if err != nil {
		return err
	}
	return resp.Body.Close()
}

func DeleteOrganizationRole(ctx context.Context, client *clients.FronteggClient, id string) error {
	resp, err := doRequest(ctx, client, "DELETE", organizationRoleURL(client, id), nil)
	if err != nil {
		return err
	}
	return resp.Body.Close()
}

func organizationRoleURL(client *clients.FronteggClient, id string) string {
	return client.Endpoint + "/identity/resources/roles/v1/" + url.PathEscape(id)
}
