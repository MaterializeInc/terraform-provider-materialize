package datasources

import (
	"context"

	"github.com/MaterializeInc/terraform-provider-materialize/pkg/frontegg"
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

func SCIMConfigs() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceSCIM2ConfigurationsRead,
		Schema:      dataSourceSCIM2ConfigurationsSchema,

		Description: "The SCIM 2.0 configurations data source allows you to fetch the SCIM 2.0 configurations.",
	}
}

func dataSourceSCIM2ConfigurationsRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerMeta, err := utils.GetProviderMeta(meta)
	if err != nil {
		return diag.FromErr(err)
	}
	client := providerMeta.Frontegg

	configurations, err := frontegg.FetchSCIM2Configurations(ctx, client)
	if err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("configurations", frontegg.FlattenSCIM2Configurations(configurations)); err != nil {
		return diag.FromErr(err)
	}

	d.SetId("scim2_configs")
	return nil
}
