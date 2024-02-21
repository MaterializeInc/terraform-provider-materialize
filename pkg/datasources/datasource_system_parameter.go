package datasources

import (
	"context"
	"fmt"

	"github.com/MaterializeInc/terraform-provider-materialize/pkg/materialize"
	"github.com/MaterializeInc/terraform-provider-materialize/pkg/utils"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func SystemParameter() *schema.Resource {
	return &schema.Resource{
		ReadContext: systemParameterRead,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The name of the specific system parameter to fetch.",
			},
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

	paramName, nameExists := d.GetOk("name")
	var query string
	var parameters []map[string]interface{}

	if nameExists {
		query = fmt.Sprintf("SHOW %s;", materialize.QuoteIdentifier(paramName.(string)))

		row := conn.QueryRow(query)
		var setting string
		if err := row.Scan(&setting); err != nil {
			return diag.FromErr(err)
		}

		// Since we're querying a specific parameter, construct the parameter slice manually
		param := make(map[string]interface{})
		param["name"] = paramName
		param["setting"] = setting
		param["description"] = paramName

		parameters = append(parameters, param)
	} else {
		query = "SHOW ALL;"

		rows, err := conn.Query(query)
		if err != nil {
			return diag.FromErr(err)
		}
		defer rows.Close()

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
	}

	// Set the parameters and region regardless of query type
	if err := d.Set("parameters", parameters); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("region", string(region)); err != nil {
		return diag.FromErr(err)
	}

	idSuffix := "all"
	if nameExists {
		idSuffix = paramName.(string)
	}
	d.SetId(utils.TransformIdWithRegion(string(region), "system_parameter_"+idSuffix))

	return diags
}
