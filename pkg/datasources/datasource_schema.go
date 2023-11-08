package datasources

import (
	"context"
	"fmt"

	"github.com/MaterializeInc/terraform-provider-materialize/pkg/materialize"
	"github.com/MaterializeInc/terraform-provider-materialize/pkg/utils"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func Schema() *schema.Resource {
	return &schema.Resource{
		ReadContext: schemaRead,
		Schema: map[string]*schema.Schema{
			"database_name": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Limit schemas to a specific database",
			},
			"schemas": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "The schemas in the account",
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

func schemaRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	databaseName := d.Get("database_name").(string)

	var diags diag.Diagnostics

	metaDb, err := utils.GetDBClientFromMeta(meta, d)
	if err != nil {
		return diag.FromErr(err)
	}
	dataSource, err := materialize.ListSchemas(metaDb, databaseName)
	if err != nil {
		return diag.FromErr(err)
	}

	schemasFormats := []map[string]interface{}{}
	for _, p := range dataSource {
		schemaMap := map[string]interface{}{}

		schemaMap["id"] = p.SchemaId.String
		schemaMap["name"] = p.SchemaName.String
		schemaMap["database_name"] = p.DatabaseName.String

		schemasFormats = append(schemasFormats, schemaMap)
	}

	if err := d.Set("schemas", schemasFormats); err != nil {
		return diag.FromErr(err)
	}

	if databaseName != "" {
		id := fmt.Sprintf("%s|schemas", databaseName)
		d.SetId(utils.TransformIdWithRegion(id))
	} else {
		d.SetId(utils.TransformIdWithRegion("schemas"))
	}

	return diags
}
