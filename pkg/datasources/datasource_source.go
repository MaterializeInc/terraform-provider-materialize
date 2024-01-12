package datasources

import (
	"context"

	"github.com/MaterializeInc/terraform-provider-materialize/pkg/materialize"
	"github.com/MaterializeInc/terraform-provider-materialize/pkg/utils"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func Source() *schema.Resource {
	return &schema.Resource{
		ReadContext: sourceRead,
		Schema: map[string]*schema.Schema{
			"database_name": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Limit sources to a specific database",
			},
			"schema_name": {
				Type:         schema.TypeString,
				Optional:     true,
				Description:  "Limit sources to a specific schema within a specific database",
				RequiredWith: []string{"database_name"},
			},
			"sources": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "The sources in the account",
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
						"type": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"envelope_type": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"connection_name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"cluster_name": {
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

func sourceRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	schemaName := d.Get("schema_name").(string)
	databaseName := d.Get("database_name").(string)

	var diags diag.Diagnostics

	metaDb, region, err := utils.GetDBClientFromMeta(meta, d)
	if err != nil {
		return diag.FromErr(err)
	}
	dataSource, err := materialize.ListSources(metaDb, schemaName, databaseName)
	if err != nil {
		return diag.FromErr(err)
	}

	sourceFormats := []map[string]interface{}{}
	for _, p := range dataSource {
		sourceMap := map[string]interface{}{}

		sourceMap["id"] = p.SourceId.String
		sourceMap["name"] = p.SourceName.String
		sourceMap["schema_name"] = p.SchemaName.String
		sourceMap["database_name"] = p.DatabaseName.String
		sourceMap["type"] = p.SourceType.String
		sourceMap["envelope_type"] = p.EnvelopeType.String
		sourceMap["connection_name"] = p.ConnectionName.String
		sourceMap["cluster_name"] = p.ClusterName.String

		sourceFormats = append(sourceFormats, sourceMap)
	}

	if err := d.Set("sources", sourceFormats); err != nil {
		return diag.FromErr(err)
	}

	SetId(string(region), "sources", databaseName, schemaName, d)

	return diags
}
