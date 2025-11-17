package datasources

import (
	"context"

	"github.com/MaterializeInc/terraform-provider-materialize/pkg/materialize"
	"github.com/MaterializeInc/terraform-provider-materialize/pkg/utils"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func SourceTable() *schema.Resource {
	return &schema.Resource{
		ReadContext: sourceTableRead,
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
				Description: "The source tables in the account",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The ID of the source table",
						},
						"name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The name of the source table",
						},
						"schema_name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The schema name of the source table",
						},
						"database_name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The database name of the source table",
						},
						"source": {
							Type:        schema.TypeList,
							Computed:    true,
							Description: "Information about the source",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"name": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "The name of the source",
									},
									"schema_name": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "The schema name of the source",
									},
									"database_name": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "The database name of the source",
									},
								},
							},
						},
						"source_type": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The type of the source",
						},
						"upstream_name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The name of the upstream table",
						},
						"upstream_schema_name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The schema name of the upstream table",
						},
						"comment": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The comment on the source table",
						},
						"owner_name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The name of the owner of the source table",
						},
					},
				},
			},
			"region": RegionSchema(),
		},
	}
}

func sourceTableRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	schemaName := d.Get("schema_name").(string)
	databaseName := d.Get("database_name").(string)

	var diags diag.Diagnostics

	metaDb, region, err := utils.GetDBClientFromMeta(meta, d)
	if err != nil {
		return diag.FromErr(err)
	}
	dataSource, err := materialize.ListSourceTables(metaDb, schemaName, databaseName)
	if err != nil {
		return diag.FromErr(err)
	}

	tableFormats := []map[string]interface{}{}
	for _, p := range dataSource {
		tableMap := map[string]interface{}{
			"id":                   p.TableId.String,
			"name":                 p.TableName.String,
			"schema_name":          p.SchemaName.String,
			"database_name":        p.DatabaseName.String,
			"source_type":          p.SourceType.String,
			"upstream_name":        p.UpstreamName.String,
			"upstream_schema_name": p.UpstreamSchemaName.String,
			"comment":              p.Comment.String,
			"owner_name":           p.OwnerName.String,
		}

		sourceMap := map[string]interface{}{
			"name":          p.SourceName.String,
			"schema_name":   p.SourceSchemaName.String,
			"database_name": p.SourceDatabaseName.String,
		}
		tableMap["source"] = []interface{}{sourceMap}

		tableFormats = append(tableFormats, tableMap)
	}

	if err := d.Set("tables", tableFormats); err != nil {
		return diag.FromErr(err)
	}

	SetId(string(region), "source_tables", databaseName, schemaName, d)

	return diags
}
