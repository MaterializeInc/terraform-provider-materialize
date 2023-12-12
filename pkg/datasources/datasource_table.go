package datasources

import (
	"context"

	"github.com/MaterializeInc/terraform-provider-materialize/pkg/materialize"
	"github.com/MaterializeInc/terraform-provider-materialize/pkg/utils"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func Table() *schema.Resource {
	return &schema.Resource{
		ReadContext: tableRead,
		Schema: map[string]*schema.Schema{
			"database_name": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Limit tables to a specific database",
			},
			"schema_name": {
				Type:         schema.TypeString,
				Optional:     true,
				Description:  "Limit tables to a specific schema within a specific database",
				RequiredWith: []string{"database_name"},
			},
			"tables": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "The tables in the account",
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
						"schema_name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"database_name": {
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

func tableRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	schemaName := d.Get("schema_name").(string)
	databaseName := d.Get("database_name").(string)

	var diags diag.Diagnostics

	metaDb, region, err := utils.GetDBClientFromMeta(meta, d)
	if err != nil {
		return diag.FromErr(err)
	}
	dataSource, err := materialize.ListTables(metaDb, schemaName, databaseName)
	if err != nil {
		return diag.FromErr(err)
	}

	tableFormats := []map[string]interface{}{}
	for _, p := range dataSource {
		tableMap := map[string]interface{}{}

		tableMap["id"] = p.TableId.String
		tableMap["name"] = p.TableName.String
		tableMap["schema_name"] = p.SchemaName.String
		tableMap["database_name"] = p.DatabaseName.String

		tableFormats = append(tableFormats, tableMap)
	}

	if err := d.Set("tables", tableFormats); err != nil {
		return diag.FromErr(err)
	}

	SetId(string(region), "tables", databaseName, schemaName, d)

	return diags
}
