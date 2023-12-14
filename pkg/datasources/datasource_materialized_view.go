package datasources

import (
	"context"

	"github.com/MaterializeInc/terraform-provider-materialize/pkg/materialize"
	"github.com/MaterializeInc/terraform-provider-materialize/pkg/utils"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func MaterializedView() *schema.Resource {
	return &schema.Resource{
		ReadContext: materializedViewRead,
		Schema: map[string]*schema.Schema{
			"database_name": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Limit materialized views to a specific database",
			},
			"schema_name": {
				Type:         schema.TypeString,
				Optional:     true,
				Description:  "Limit materialized views to a specific schema within a specific database",
				RequiredWith: []string{"database_name"},
			},
			"materialized_views": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "The materialized views in the account",
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
			"region": RegionSchema(),
		},
	}
}

func materializedViewRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	schemaName := d.Get("schema_name").(string)
	databaseName := d.Get("database_name").(string)

	var diags diag.Diagnostics

	metaDb, region, err := utils.GetDBClientFromMeta(meta, d)
	if err != nil {
		return diag.FromErr(err)
	}
	dataSource, err := materialize.ListMaterializedViews(metaDb, schemaName, databaseName)
	if err != nil {
		return diag.FromErr(err)
	}

	materizliedViewFormats := []map[string]interface{}{}
	for _, p := range dataSource {
		materizliedViewMap := map[string]interface{}{}

		materizliedViewMap["id"] = p.MaterializedViewId.String
		materizliedViewMap["name"] = p.MaterializedViewName.String
		materizliedViewMap["schema_name"] = p.SchemaName.String
		materizliedViewMap["database_name"] = p.DatabaseName.String

		materizliedViewFormats = append(materizliedViewFormats, materizliedViewMap)
	}

	if err := d.Set("materialized_views", materizliedViewFormats); err != nil {
		return diag.FromErr(err)
	}

	SetId(string(region), "materialized_views", databaseName, schemaName, d)
	return diags
}
