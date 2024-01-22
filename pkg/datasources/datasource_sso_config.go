package datasources

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/MaterializeInc/terraform-provider-materialize/pkg/utils"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

var dataSourceSSOConfigSchema = map[string]*schema.Schema{
	"sso_configs": {
		Type:     schema.TypeList,
		Computed: true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"id": {
					Type:        schema.TypeString,
					Computed:    true,
					Description: "The ID of the SSO configuration.",
				},
				"enabled": {
					Type:        schema.TypeBool,
					Computed:    true,
					Description: "Whether SSO is enabled or not.",
				},
				"sso_endpoint": {
					Type:        schema.TypeString,
					Computed:    true,
					Description: "The URL endpoint for the SSO service.",
				},
				"public_certificate": {
					Type:        schema.TypeString,
					Computed:    true,
					Description: "The public certificate of the SSO service in PEM format.",
				},
				"sign_request": {
					Type:        schema.TypeBool,
					Computed:    true,
					Description: "Indicates whether the SSO request needs to be digitally signed.",
				},
				"type": {
					Type:        schema.TypeString,
					Computed:    true,
					Description: "The type of SSO protocol being used (e.g., SAML, OIDC).",
				},
				"oidc_client_id": {
					Type:        schema.TypeString,
					Computed:    true,
					Description: "The client ID of the OIDC application.",
				},
				"oidc_secret": {
					Type:        schema.TypeString,
					Computed:    true,
					Description: "The client secret of the OIDC application.",
				},
				"domains": {
					Type:        schema.TypeList,
					Computed:    true,
					Description: "List of domains associated with the SSO configuration.",
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"id": {
								Type:     schema.TypeString,
								Computed: true,
							},
							"domain": {
								Type:     schema.TypeString,
								Computed: true,
							},
							"validated": {
								Type:     schema.TypeBool,
								Computed: true,
							},
						},
					},
				},
				"role_ids": {
					Type:        schema.TypeList,
					Computed:    true,
					Description: "List of role IDs associated with the SSO configuration.",
					Elem:        &schema.Schema{Type: schema.TypeString},
				},
				"groups": {
					Type:        schema.TypeList,
					Computed:    true,
					Description: "List of groups associated with the SSO configuration.",
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"group": {
								Type:     schema.TypeString,
								Computed: true,
							},
							"enabled": {
								Type:     schema.TypeBool,
								Computed: true,
							},
							"id": {
								Type:     schema.TypeString,
								Computed: true,
							},
							"role_ids": {
								Type:     schema.TypeList,
								Computed: true,
								Elem:     &schema.Schema{Type: schema.TypeString},
							},
						},
					},
				},
			},
		},
	},
}

func SSOConfig() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceSSOConfigRead,
		Schema:      dataSourceSSOConfigSchema,
	}
}

func dataSourceSSOConfigRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerMeta, err := utils.GetProviderMeta(meta)
	if err != nil {
		return diag.FromErr(err)
	}
	client := providerMeta.Frontegg

	endpoint := fmt.Sprintf("%s/frontegg/team/resources/sso/v1/configurations", client.Endpoint)
	req, err := http.NewRequestWithContext(ctx, "GET", endpoint, nil)
	if err != nil {
		return diag.FromErr(err)
	}

	req.Header.Add("Authorization", "Bearer "+client.Token)

	log.Printf("[DEBUG] Sending request to %s", endpoint)

	resp, err := client.HTTPClient.Do(req)
	if err != nil {
		return diag.FromErr(err)
	}
	defer resp.Body.Close()

	log.Printf("[DEBUG] Received status code: %d", resp.StatusCode)

	if resp.StatusCode != http.StatusOK {
		responseData, err := io.ReadAll(resp.Body)
		if err != nil {
			return diag.FromErr(err)
		}
		return diag.Errorf("error reading SSO configurations: status %d, response: %s", resp.StatusCode, string(responseData))
	}

	var ssoConfigurations []map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&ssoConfigurations); err != nil {
		return diag.FromErr(err)
	}

	var configurations []map[string]interface{}
	for _, ssoConfig := range ssoConfigurations {
		configuration := make(map[string]interface{})
		configuration["id"] = ssoConfig["id"]
		configuration["enabled"] = ssoConfig["enabled"]
		configuration["sso_endpoint"] = ssoConfig["ssoEndpoint"]
		configuration["public_certificate"] = ssoConfig["publicCertificate"]
		configuration["sign_request"] = ssoConfig["signRequest"]
		configuration["type"] = ssoConfig["type"]
		configuration["oidc_client_id"] = ssoConfig["oidcClientId"]
		configuration["oidc_secret"] = ssoConfig["oidcSecret"]

		domainsRaw, ok := ssoConfig["domains"].([]interface{})
		if !ok {
			continue
		}
		var domains []map[string]interface{}
		for _, domain := range domainsRaw {
			domainMap, ok := domain.(map[string]interface{})
			if !ok {
				continue
			}
			domainData := map[string]interface{}{
				"id":        domainMap["id"],
				"domain":    domainMap["domain"],
				"validated": domainMap["validated"],
			}
			domains = append(domains, domainData)
		}
		configuration["domains"] = domains

		roleIDsRaw, ok := ssoConfig["roleIds"].([]interface{})
		if !ok {
			continue
		}

		// Convert role IDs to []string
		var roleIDs []string
		for _, roleID := range roleIDsRaw {
			if roleIDStr, ok := roleID.(string); ok {
				roleIDs = append(roleIDs, roleIDStr)
			}
		}

		configuration["role_ids"] = roleIDs

		groupsRaw, ok := ssoConfig["groups"].([]interface{})
		if !ok {
			continue
		}
		var groups []map[string]interface{}
		for _, group := range groupsRaw {
			groupMap, ok := group.(map[string]interface{})
			if !ok {
				continue
			}
			groupData := map[string]interface{}{
				"group":    groupMap["group"],
				"enabled":  groupMap["enabled"],
				"id":       groupMap["id"],
				"role_ids": groupMap["roleIds"],
			}
			groups = append(groups, groupData)
		}
		configuration["groups"] = groups
		configurations = append(configurations, configuration)
	}

	if err := d.Set("sso_configs", configurations); err != nil {
		return diag.FromErr(err)
	}

	if len(configurations) > 0 {
		d.SetId(configurations[0]["id"].(string))
	} else {
		d.SetId("no_sso_configs")
	}

	return nil
}
