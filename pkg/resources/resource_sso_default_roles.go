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

var SSODefaultRolesSchema = map[string]*schema.Schema{
	"sso_config_id": {
		Type:        schema.TypeString,
		Required:    true,
		Description: "The ID of the associated SSO configuration.",
	},
	"roles": {
		Type:        schema.TypeSet,
		Elem:        &schema.Schema{Type: schema.TypeString},
		Required:    true,
		Description: "Set of default role names for the SSO configuration. These roles will be assigned by default to users who sign up via SSO.",
	},
}

type SSOConfigRolesResponse struct {
	RoleIds []string `json:"roleIds"`
}

func SSODefaultRoles() *schema.Resource {
	return &schema.Resource{
		CreateContext: ssoDefaultRolesCreateOrUpdate,
		ReadContext:   ssoDefaultRolesRead,
		UpdateContext: ssoDefaultRolesCreateOrUpdate,
		DeleteContext: ssoDefaultRolesDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: SSODefaultRolesSchema,

		Description: "The SSO default roles resource allows you to set the default roles for an SSO configuration. These roles will be assigned to users who sign in with SSO.",
	}
}

func ssoDefaultRolesCreateOrUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerMeta, err := utils.GetProviderMeta(meta)
	if err != nil {
		return diag.FromErr(err)
	}
	client := providerMeta.Frontegg

	ssoConfigID := d.Get("sso_config_id").(string)
	roleNames := convertToStringSlice(d.Get("roles").(*schema.Set).List())

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

	// Prepare payload for PUT request
	payload := map[string]interface{}{
		"roleIds": roleIDs,
	}

	requestBody, err := json.Marshal(payload)
	if err != nil {
		return diag.FromErr(err)
	}

	// Construct the PUT request
	endpoint := fmt.Sprintf("%s/frontegg/team/resources/sso/v1/configurations/%s/roles", client.Endpoint, ssoConfigID)
	req, err := http.NewRequestWithContext(ctx, "PUT", endpoint, bytes.NewBuffer(requestBody))
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
	if resp.StatusCode != http.StatusCreated {
		responseData, err := io.ReadAll(resp.Body)
		if err != nil {
			return diag.FromErr(err)
		}
		return diag.Errorf("error setting SSO default roles: status %d, response: %s", resp.StatusCode, string(responseData))
	}

	d.SetId(ssoConfigID)
	return ssoDefaultRolesRead(ctx, d, meta)
}

func ssoDefaultRolesRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerMeta, err := utils.GetProviderMeta(meta)
	if err != nil {
		return diag.FromErr(err)
	}
	client := providerMeta.Frontegg

	ssoConfigID := d.Id()

	if err := d.Set("sso_config_id", ssoConfigID); err != nil {
		return diag.FromErr(err)
	}

	// Construct the GET request
	endpoint := fmt.Sprintf("%s/frontegg/team/resources/sso/v1/configurations/%s/roles", client.Endpoint, ssoConfigID)
	req, err := http.NewRequestWithContext(ctx, "GET", endpoint, nil)
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
		return diag.Errorf("error reading SSO default roles: status %d, response: %s", resp.StatusCode, string(responseData))
	}

	var configRoles SSOConfigRolesResponse
	if err := json.NewDecoder(resp.Body).Decode(&configRoles); err != nil {
		return diag.FromErr(err)
	}

	// Map role IDs back to names
	roleMap, err := utils.ListRoles(ctx, client)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error fetching roles: %s", err))
	}

	var roleNames []string
	for _, roleID := range configRoles.RoleIds {
		for name, id := range roleMap {
			if id == roleID {
				roleNames = append(roleNames, name)
				break
			}
		}
	}

	if err := d.Set("roles", schema.NewSet(schema.HashString, convertToStringInterface(roleNames))); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func ssoDefaultRolesDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerMeta, err := utils.GetProviderMeta(meta)
	if err != nil {
		return diag.FromErr(err)
	}
	client := providerMeta.Frontegg

	ssoConfigID := d.Id()

	// Prepare an empty payload to clear the default roles
	payload := map[string]interface{}{
		// Empty array to signify no roles
		"roleIds": []string{},
	}

	requestBody, err := json.Marshal(payload)
	if err != nil {
		return diag.FromErr(err)
	}

	// Construct the PUT request
	endpoint := fmt.Sprintf("%s/frontegg/team/resources/sso/v1/configurations/%s/roles", client.Endpoint, ssoConfigID)
	req, err := http.NewRequestWithContext(ctx, "PUT", endpoint, bytes.NewBuffer(requestBody))
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
	if resp.StatusCode != http.StatusCreated {
		responseData, err := io.ReadAll(resp.Body)
		if err != nil {
			return diag.FromErr(err)
		}
		return diag.Errorf("error clearing SSO default roles: status %d, response: %s", resp.StatusCode, string(responseData))
	}

	d.SetId("")
	return nil
}

func convertToStringInterface(stringSlice []string) []interface{} {
	var interfaceSlice []interface{}
	for _, str := range stringSlice {
		interfaceSlice = append(interfaceSlice, str)
	}
	return interfaceSlice
}
