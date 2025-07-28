package datasources

import (
	"context"

	"github.com/MaterializeInc/terraform-provider-materialize/pkg/frontegg"
	"github.com/MaterializeInc/terraform-provider-materialize/pkg/utils"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

var dataSourceSCIMGroupsSchema = map[string]*schema.Schema{
	"groups": {
		Type:     schema.TypeList,
		Computed: true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"id": {
					Type:        schema.TypeString,
					Computed:    true,
					Description: "The ID of the group. This is a unique identifier for the group. ",
				},
				"name": {
					Type:        schema.TypeString,
					Computed:    true,
					Description: "The name of the group.",
				},
				"description": {
					Type:        schema.TypeString,
					Computed:    true,
					Description: "The description of the group.",
				},
				"metadata": {
					Type:        schema.TypeString,
					Computed:    true,
					Optional:    true,
					Description: "The metadata of the group.",
				},
				"roles": {
					Type:     schema.TypeList,
					Computed: true,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"id": {
								Type:        schema.TypeString,
								Computed:    true,
								Description: "The ID of the role. This is a unique identifier for the role.",
							},
							"key": {
								Type:        schema.TypeString,
								Computed:    true,
								Description: "The key of the role.",
							},
							"name": {
								Type:        schema.TypeString,
								Computed:    true,
								Description: "The name of the role.",
							},
							"description": {
								Type:        schema.TypeString,
								Computed:    true,
								Description: "The description of the role.",
							},
							"is_default": {
								Type:        schema.TypeBool,
								Computed:    true,
								Description: "Indicates whether the role is the default role.",
							},
						},
					},
				},
				"users": {
					Type:     schema.TypeList,
					Computed: true,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"id": {
								Type:        schema.TypeString,
								Computed:    true,
								Description: "The ID of the user. This is a unique identifier for the user.",
							},
							"name": {
								Type:        schema.TypeString,
								Computed:    true,
								Description: "The name of the user.",
							},
							"email": {
								Type:        schema.TypeString,
								Computed:    true,
								Description: "The email of the user.",
							},
						},
					},
				},
				"managed_by": {
					Type:        schema.TypeString,
					Computed:    true,
					Description: "The ID of the user who manages the group.",
				},
			},
		},
	},
}

// SCIMGroups data source function
func SCIMGroups() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceSCIMGroupsRead,
		Schema:      dataSourceSCIMGroupsSchema,

		Description: "The SCIM groups data source allows you to fetch the available groups.",
	}
}

// Read function for SCIM groups data source
func dataSourceSCIMGroupsRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerMeta, err := utils.GetProviderMeta(meta)
	if err != nil {
		return diag.FromErr(err)
	}

	// Validate that SCIM groups data source is only used in SaaS mode
	if diags := providerMeta.ValidateSaaSOnly("materialize_scim_groups data source"); diags.HasError() {
		return diags
	}

	client := providerMeta.Frontegg

	groups, err := frontegg.FetchSCIMGroups(ctx, client)
	if err != nil {
		return diag.FromErr(err)
	}

	// Map the response to the schema
	if err := d.Set("groups", frontegg.FlattenScimGroups(groups)); err != nil {
		return diag.FromErr(err)
	}

	// Set the ID of the data source
	d.SetId("scim_groups")

	return nil
}
