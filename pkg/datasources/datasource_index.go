package datasources

import (
	"context"

	"github.com/MaterializeInc/terraform-provider-materialize/pkg/materialize"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/jmoiron/sqlx"
)

func Index() *schema.Resource {
	return &schema.Resource{
		ReadContext: indexRead,
		Schema: map[string]*schema.Schema{
			"database_name": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Limit indexes to a specific database",
			},
			"schema_name": {
				Type:         schema.TypeString,
				Optional:     true,
				Description:  "Limit indexes to a specific schema within a specific database",
				RequiredWith: []string{"database_name"},
			},
			"indexes": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "The indexes in the account",
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
						"obj_name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"obj_schema": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"obj_database": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func indexRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	schemaName := d.Get("schema_name").(string)
	databaseName := d.Get("database_name").(string)

	var diags diag.Diagnostics

	dataSource, err := materialize.ListIndexes(meta.(*sqlx.DB), schemaName, databaseName)
	if err != nil {
		return diag.FromErr(err)
	}

	indexFormats := []map[string]interface{}{}
	for _, p := range dataSource {
		indexMap := map[string]interface{}{}

		indexMap["id"] = p.IndexId.String
		indexMap["name"] = p.IndexName.String
		indexMap["obj_name"] = p.ObjectName.String
		indexMap["obj_schema"] = p.ObjectSchemaName.String
		indexMap["obj_database"] = p.ObjectDatabaseName.String

		indexFormats = append(indexFormats, indexMap)
	}

	if err := d.Set("indexes", indexFormats); err != nil {
		return diag.FromErr(err)
	}

	SetId("indexes", databaseName, schemaName, d)
	return diags
}
