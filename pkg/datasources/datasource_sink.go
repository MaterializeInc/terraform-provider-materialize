package datasources

import (
	"context"

	"github.com/MaterializeInc/terraform-provider-materialize/pkg/materialize"
	"github.com/MaterializeInc/terraform-provider-materialize/pkg/utils"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func Sink() *schema.Resource {
	return &schema.Resource{
		ReadContext: sinkRead,
		Schema: map[string]*schema.Schema{
			"database_name": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Limit sinks to a specific database",
			},
			"schema_name": {
				Type:         schema.TypeString,
				Optional:     true,
				Description:  "Limit sinks to a specific schema within a specific database",
				RequiredWith: []string{"database_name"},
			},
			"sinks": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "The sinks in the account",
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

func sinkRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	schemaName := d.Get("schema_name").(string)
	databaseName := d.Get("database_name").(string)

	var diags diag.Diagnostics

	metaDb, region, err := utils.GetDBClientFromMeta(meta, d)
	if err != nil {
		return diag.FromErr(err)
	}
	dataSource, err := materialize.ListSinks(metaDb, schemaName, databaseName)
	if err != nil {
		return diag.FromErr(err)
	}

	sinkFormats := []map[string]interface{}{}
	for _, p := range dataSource {
		sinkMap := map[string]interface{}{}

		sinkMap["id"] = p.SinkId.String
		sinkMap["name"] = p.SinkName.String
		sinkMap["schema_name"] = p.SchemaName.String
		sinkMap["database_name"] = p.DatabaseName.String
		sinkMap["type"] = p.SinkType.String
		sinkMap["envelope_type"] = p.EnvelopeType.String
		sinkMap["connection_name"] = p.ConnectionName.String
		sinkMap["cluster_name"] = p.ClusterName.String

		sinkFormats = append(sinkFormats, sinkMap)
	}

	if err := d.Set("sinks", sinkFormats); err != nil {
		return diag.FromErr(err)
	}

	SetId(string(region), "sinks", databaseName, schemaName, d)

	return diags
}
