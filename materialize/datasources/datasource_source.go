package datasources

import (
	"context"
	"database/sql"
	"errors"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func Source() *schema.Resource {
	return &schema.Resource{
		ReadContext: sourceRead,
		Schema: map[string]*schema.Schema{
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

func sourceRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*sql.DB)
	rows, err := conn.Query(`
		SELECT
			mz_sources.id,
			mz_sources.name,
			mz_sources.type,
			mz_sources.size,
			mz_sources.envelope_type,
			mz_connections.name as connection_name,
			mz_clusters.name as cluster_name
		FROM mz_sources
		JOIN mz_schemas
			ON mz_sources.schema_id = mz_schemas.id
		LEFT JOIN mz_connections
			ON mz_sources.connection_id = mz_connections.id
		LEFT JOIN mz_clusters
			ON mz_sources.cluster_id = mz_clusters.id;
	`)
	if errors.Is(err, sql.ErrNoRows) {
		log.Printf("[DEBUG] no sources found in account")
		d.SetId("")
		return diag.FromErr(err)
	} else if err != nil {
		log.Println("[DEBUG] failed to list sources")
		d.SetId("")
		return diag.FromErr(err)
	}

	sourceFormats := []map[string]interface{}{}
	for rows.Next() {
		var id, name, source_type, size, envelope_type, connection_name, cluster_name string
		rows.Scan(&id, &name, &source_type, &size, &envelope_type, &connection_name, &cluster_name)

		sourceMap := map[string]interface{}{}

		sourceMap["id"] = id
		sourceMap["name"] = name
		sourceMap["type"] = source_type
		sourceMap["size"] = size
		sourceMap["envelope_type"] = envelope_type
		sourceMap["connection_name"] = connection_name
		sourceMap["cluster_name"] = cluster_name

		sourceFormats = append(sourceFormats, sourceMap)
	}

	if err := d.Set("sources", sourceFormats); err != nil {
		return diag.FromErr(err)
	}

	d.SetId("sources")
	return diags
}
