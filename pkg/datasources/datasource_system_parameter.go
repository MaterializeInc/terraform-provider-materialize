package datasources

import (
	"context"

	"github.com/MaterializeInc/terraform-provider-materialize/pkg/utils"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func SystemParameter() *schema.Resource {
	return &schema.Resource{
		ReadContext: systemParameterRead,
		Schema: map[string]*schema.Schema{
			"parameters": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The name of the system parameter.",
						},
						"setting": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The value of the system parameter.",
						},
						"description": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The description of the system parameter.",
						},
					},
				},
			},
			"region": RegionSchema(),
		},
	}
}

func systemParameterRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	metaDb, region, err := utils.GetDBClientFromMeta(meta, d)
	if err != nil {
		return diag.FromErr(err)
	}
	conn := metaDb

	rows, err := conn.Query("SHOW ALL;")
	if err != nil {
		return diag.FromErr(err)
	}
	defer rows.Close()

	var parameters []map[string]interface{}

	for rows.Next() {
		var name, setting, description string
		if err := rows.Scan(&name, &setting, &description); err != nil {
			return diag.FromErr(err)
		}

		param := make(map[string]interface{})
		param["name"] = name
		param["setting"] = setting
		param["description"] = description

		parameters = append(parameters, param)
	}

	if err := rows.Err(); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("parameters", parameters); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("region", string(region)); err != nil {
		return diag.FromErr(err)
	}

	d.SetId(utils.TransformIdWithRegion(string(region), "system_parameter"))

	return diags
}
