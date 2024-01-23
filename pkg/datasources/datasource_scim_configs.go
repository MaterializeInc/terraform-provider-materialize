package datasources

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/MaterializeInc/terraform-provider-materialize/pkg/utils"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

var dataSourceSCIM2ConfigurationsSchema = map[string]*schema.Schema{
	"configurations": {
		Type:     schema.TypeList,
		Computed: true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"id": {
					Type:        schema.TypeString,
					Computed:    true,
					Description: "The unique identifier of the SCIM 2.0 configuration.",
				},
				"source": {
					Type:        schema.TypeString,
					Computed:    true,
					Description: "The source of the SCIM 2.0 configuration.",
				},
				"tenant_id": {
					Type:        schema.TypeString,
					Computed:    true,
					Description: "The tenant ID related to the SCIM 2.0 configuration.",
				},
				"connection_name": {
					Type:        schema.TypeString,
					Computed:    true,
					Description: "The name of the SCIM 2.0 connection.",
				},
				"sync_to_user_management": {
					Type:        schema.TypeBool,
					Computed:    true,
					Description: "Indicates if the configuration is synced to user management.",
				},
				"created_at": {
					Type:        schema.TypeString,
					Computed:    true,
					Description: "The creation timestamp of the SCIM 2.0 configuration.",
				},
			},
		},
	},
}

type SCIM2Configuration struct {
	ID                   string `json:"id"`
	Source               string `json:"source"`
	TenantID             string `json:"tenantId"`
	ConnectionName       string `json:"connectionName"`
	SyncToUserManagement bool   `json:"syncToUserManagement"`
	CreatedAt            string `json:"createdAt"`
}

type SCIM2ConfigurationsResponse []SCIM2Configuration

func SCIMConfigs() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceSCIM2ConfigurationsRead,
		Schema:      dataSourceSCIM2ConfigurationsSchema,
	}
}

func dataSourceSCIM2ConfigurationsRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerMeta, err := utils.GetProviderMeta(meta)
	if err != nil {
		return diag.FromErr(err)
	}
	client := providerMeta.Frontegg

	endpoint := fmt.Sprintf("%s/frontegg/directory/resources/v1/configurations/scim2", client.Endpoint)
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
		var sb strings.Builder
		_, err = io.Copy(&sb, resp.Body)
		if err != nil {
			return diag.FromErr(err)
		}
		return diag.Errorf("error reading SCIM 2.0 configurations: status %d, response: %s", resp.StatusCode, sb.String())
	}

	var configurations SCIM2ConfigurationsResponse
	if err := json.NewDecoder(resp.Body).Decode(&configurations); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("configurations", flattenSCIM2Configurations(configurations)); err != nil {
		return diag.FromErr(err)
	}

	// Set the ID of the data source
	d.SetId("scim2_configs")

	return nil
}

// Helper function to flatten the SCIM 2.0 configurations data
func flattenSCIM2Configurations(configurations SCIM2ConfigurationsResponse) []interface{} {
	var flattenedConfigurations []interface{}
	for _, config := range configurations {
		flattenedConfig := map[string]interface{}{
			"id":                      config.ID,
			"source":                  config.Source,
			"tenant_id":               config.TenantID,
			"connection_name":         config.ConnectionName,
			"sync_to_user_management": config.SyncToUserManagement,
			"created_at":              config.CreatedAt,
		}
		flattenedConfigurations = append(flattenedConfigurations, flattenedConfig)
	}
	return flattenedConfigurations
}
