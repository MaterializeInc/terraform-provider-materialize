package resources

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/MaterializeInc/terraform-provider-materialize/pkg/utils"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

var SSORoleGroupMappingSchema = map[string]*schema.Schema{
	"sso_config_id": {
		Type:        schema.TypeString,
		Required:    true,
		Description: "The ID of the associated SSO configuration.",
	},
	"group": {
		Type:        schema.TypeString,
		Required:    true,
		Description: "The name of the SSO group.",
	},
	"roles": {
		Type:        schema.TypeList,
		Elem:        &schema.Schema{Type: schema.TypeString},
		Required:    true,
		Description: "List of role names associated with the group.",
	},
}

// GroupMapping represents the structure for SSO group role mapping.
type GroupMapping struct {
	ID          string   `json:"id"`
	Group       string   `json:"group"`
	Enabled     bool     `json:"enabled"`
	RoleIds     []string `json:"roleIds"`
	SsoConfigId string   `json:"-"`
}

func SSORoleGroupMapping() *schema.Resource {
	return &schema.Resource{
		CreateContext: ssoGroupMappingCreate,
		ReadContext:   ssoGroupMappingRead,
		UpdateContext: ssoGroupMappingUpdate,
		DeleteContext: ssoGroupMappingDelete,

		Schema: SSORoleGroupMappingSchema,
	}
}

func ssoGroupMappingCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerMeta, err := utils.GetProviderMeta(meta)
	if err != nil {
		return diag.FromErr(err)
	}
	client := providerMeta.Frontegg

	ssoConfigID := d.Get("sso_config_id").(string)
	group := d.Get("group").(string)
	roleNames := convertToStringSlice(d.Get("roles").([]interface{}))

	// Fetch role IDs based on role names.
	roleMap, err := utils.ListRoles(ctx, client)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error fetching roles: %s", err))
	}

	var roleIDs []string
	for _, roleName := range roleNames {
		if roleID, ok := roleMap[roleName]; ok {
			roleIDs = append(roleIDs, roleID)
		} else {
			// TODO: Fail the process if the role is not found
			return diag.Errorf("role not found: %s", roleName)
		}
	}

	payload := map[string]interface{}{
		"group":   group,
		"roleIds": roleIDs,
	}

	requestBody, err := json.Marshal(payload)
	if err != nil {
		return diag.FromErr(err)
	}

	endpoint := fmt.Sprintf("%s/frontegg/team/resources/sso/v1/configurations/%s/groups", client.Endpoint, ssoConfigID)
	req, err := http.NewRequestWithContext(ctx, "POST", endpoint, bytes.NewBuffer(requestBody))
	if err != nil {
		return diag.FromErr(err)
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", "Bearer "+client.Token)

	resp, err := client.HTTPClient.Do(req)
	if err != nil {
		return diag.FromErr(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		responseData, err := io.ReadAll(resp.Body)
		if err != nil {
			return diag.FromErr(err)
		}
		return diag.Errorf("error creating SSO group mapping: status %d, response: %s", resp.StatusCode, string(responseData))
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return diag.FromErr(err)
	}

	groupID, ok := result["id"].(string)
	if !ok {
		return diag.Errorf("error retrieving ID from SSO group mapping creation response")
	}

	d.SetId(groupID)
	return ssoGroupMappingRead(ctx, d, meta)
}

func ssoGroupMappingRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerMeta, err := utils.GetProviderMeta(meta)
	if err != nil {
		return diag.FromErr(err)
	}
	client := providerMeta.Frontegg

	ssoConfigID := d.Get("sso_config_id").(string)
	groupID := d.Id()

	endpoint := fmt.Sprintf("%s/frontegg/team/resources/sso/v1/configurations/%s/groups", client.Endpoint, ssoConfigID)
	req, err := http.NewRequestWithContext(ctx, "GET", endpoint, nil)
	if err != nil {
		return diag.FromErr(err)
	}
	req.Header.Add("Authorization", "Bearer "+client.Token)

	resp, err := client.HTTPClient.Do(req)
	if err != nil {
		return diag.FromErr(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		responseData, err := io.ReadAll(resp.Body)
		if err != nil {
			return diag.FromErr(err)
		}
		return diag.Errorf("error reading SSO group mappings: status %d, response: %s", resp.StatusCode, string(responseData))
	}

	var groups []GroupMapping
	if err := json.NewDecoder(resp.Body).Decode(&groups); err != nil {
		return diag.FromErr(err)
	}

	for _, group := range groups {
		if group.ID == groupID {
			d.Set("group", group.Group)
			d.Set("role_ids", group.RoleIds)
			d.Set("enabled", group.Enabled)
			return nil
		}
	}

	d.SetId("")
	return nil
}

func ssoGroupMappingUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerMeta, err := utils.GetProviderMeta(meta)
	if err != nil {
		return diag.FromErr(err)
	}
	client := providerMeta.Frontegg

	ssoConfigID := d.Get("sso_config_id").(string)
	group := d.Get("group").(string)
	roleNames := convertToStringSlice(d.Get("roles").([]interface{}))

	// Fetch role IDs based on role names
	roleMap, err := utils.ListRoles(ctx, client)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error fetching roles: %s", err))
	}

	var roleIDs []string
	for _, roleName := range roleNames {
		if roleID, ok := roleMap[roleName]; ok {
			roleIDs = append(roleIDs, roleID)
		} else {
			return diag.Errorf("role not found: %s", roleName)
		}
	}

	// Prepare payload for the PATCH request
	payload := map[string]interface{}{
		"group":   group,
		"roleIds": roleIDs,
	}

	// Serialize the payload
	requestBody, err := json.Marshal(payload)
	if err != nil {
		return diag.FromErr(err)
	}

	// Construct the PATCH request
	endpoint := fmt.Sprintf("%s/frontegg/team/resources/sso/v1/configurations/%s/groups/%s", client.Endpoint, ssoConfigID, d.Id())
	req, err := http.NewRequestWithContext(ctx, "PATCH", endpoint, bytes.NewBuffer(requestBody))
	if err != nil {
		return diag.FromErr(err)
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", "Bearer "+client.Token)

	// Send the request
	resp, err := client.HTTPClient.Do(req)
	if err != nil {
		return diag.FromErr(err)
	}
	defer resp.Body.Close()

	// Check for a successful response
	if resp.StatusCode != http.StatusOK {
		responseData, err := io.ReadAll(resp.Body)
		if err != nil {
			return diag.FromErr(err)
		}
		return diag.Errorf("error updating SSO group mapping: status %d, response: %s", resp.StatusCode, string(responseData))
	}

	return ssoGroupMappingRead(ctx, d, meta)
}

func ssoGroupMappingDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerMeta, err := utils.GetProviderMeta(meta)
	if err != nil {
		return diag.FromErr(err)
	}
	client := providerMeta.Frontegg

	ssoConfigID := d.Get("sso_config_id").(string)

	// Construct the DELETE request
	endpoint := fmt.Sprintf("%s/frontegg/team/resources/sso/v1/configurations/%s/groups/%s", client.Endpoint, ssoConfigID, d.Id())
	req, err := http.NewRequestWithContext(ctx, "DELETE", endpoint, nil)
	if err != nil {
		return diag.FromErr(err)
	}
	req.Header.Add("Authorization", "Bearer "+client.Token)

	// Send the request
	resp, err := client.HTTPClient.Do(req)
	if err != nil {
		return diag.FromErr(err)
	}
	defer resp.Body.Close()

	// Check for a successful response
	if resp.StatusCode != http.StatusOK {
		responseData, err := io.ReadAll(resp.Body)
		if err != nil {
			return diag.FromErr(err)
		}
		return diag.Errorf("error deleting SSO group mapping: status %d, response: %s", resp.StatusCode, string(responseData))
	}

	d.SetId("")
	return nil
}