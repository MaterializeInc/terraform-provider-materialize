package datasources

import (
	"context"
	"fmt"

	"github.com/MaterializeInc/terraform-provider-materialize/pkg/materialize"
	"github.com/MaterializeInc/terraform-provider-materialize/pkg/utils"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func Role() *schema.Resource {
	return &schema.Resource{
		ReadContext: roleRead,
		Schema: map[string]*schema.Schema{
			"like_pattern": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Filter roles by name using SQL LIKE pattern (e.g., 'prod_%', '%_admin')",
			},
			"roles": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "The roles in the account",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"name": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"region": RegionSchema(),
		},
	}
}

func roleRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	likePattern := d.Get("like_pattern").(string)

	var diags diag.Diagnostics

	metaDb, region, err := utils.GetDBClientFromMeta(meta, d)
	if err != nil {
		return diag.FromErr(err)
	}
	dataSource, err := materialize.ListRoles(metaDb, likePattern)
	if err != nil {
		return diag.FromErr(err)
	}

	roleFormats := []map[string]interface{}{}
	for _, p := range dataSource {
		roleMap := map[string]interface{}{}

		roleMap["id"] = p.RoleId.String
		roleMap["name"] = p.RoleName.String

		roleFormats = append(roleFormats, roleMap)
	}

	if err := d.Set("roles", roleFormats); err != nil {
		return diag.FromErr(err)
	}

	idSuffix := "roles"
	if likePattern != "" {
		idSuffix = fmt.Sprintf("roles|%s", likePattern)
	}
	d.SetId(utils.TransformIdWithRegion(string(region), idSuffix))
	return diags
}
