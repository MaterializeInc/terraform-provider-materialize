package datasources

import (
	"context"
	"database/sql"
	"errors"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func Sink() *schema.Resource {
	return &schema.Resource{
		ReadContext: sinkRead,
		Schema: map[string]*schema.Schema{
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
						"type": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"size": {
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
		},
	}
}

func sinkRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*sql.DB)
	rows, err := conn.Query(`
		SELECT
			mz_sinks.id,
			mz_sinks.name,
			mz_sinks.type,
			mz_sinks.size,
			mz_sinks.envelope_type,
			mz_connections.name as connection_name,
			mz_clusters.name as cluster_name
		FROM mz_sinks
		JOIN mz_schemas
			ON mz_sinks.schema_id = mz_schemas.id
		LEFT JOIN mz_connections
			ON mz_sinks.connection_id = mz_connections.id
		LEFT JOIN mz_clusters
			ON mz_sinks.cluster_id = mz_clusters.id;
	`)
	if errors.Is(err, sql.ErrNoRows) {
		log.Printf("[DEBUG] no sinks found in account")
		d.SetId("")
		return diag.FromErr(err)
	} else if err != nil {
		log.Println("[DEBUG] failed to list sinks")
		d.SetId("")
		return diag.FromErr(err)
	}

	sinkFormats := []map[string]interface{}{}
	for rows.Next() {
		var id, name, sink_type, size, envelope_type, connection_name, cluster_name string
		rows.Scan(&id, &name, &sink_type, &size, &envelope_type, &connection_name, &cluster_name)

		sinkMap := map[string]interface{}{}

		sinkMap["id"] = id
		sinkMap["name"] = name
		sinkMap["sink_type"] = sink_type
		sinkMap["size"] = size
		sinkMap["envelope_type"] = envelope_type
		sinkMap["connection_name"] = connection_name
		sinkMap["cluster_name"] = cluster_name

		sinkFormats = append(sinkFormats, sinkMap)
	}

	if err := d.Set("sinks", sinkFormats); err != nil {
		return diag.FromErr(err)
	}

	d.SetId("sinks")
	return diags
}
