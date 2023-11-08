package datasources

import (
	"context"

	"github.com/MaterializeInc/terraform-provider-materialize/pkg/materialize"
	"github.com/MaterializeInc/terraform-provider-materialize/pkg/utils"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func Role() *schema.Resource {
	return &schema.Resource{
		ReadContext: roleRead,
		Schema: map[string]*schema.Schema{
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
			"region": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func roleRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	metaDb, err := utils.GetDBClientFromMeta(meta, d)
	if err != nil {
		return diag.FromErr(err)
	}
	dataSource, err := materialize.ListRoles(metaDb)
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

	d.SetId(utils.TransformIdWithRegion("roles"))
	return diags
}
