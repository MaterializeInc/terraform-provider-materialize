package datasources

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"strings"

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

func sourceQuery(databaseName, schemaName string) string {
	q := strings.Builder{}
	q.WriteString(`
		SELECT
			mz_sources.id,
			mz_sources.name,
			mz_schemas.name,
			mz_databases.name,
			mz_sources.type,
			mz_sources.size,
			mz_sources.envelope_type,
			mz_connections.name as connection_name,
			mz_clusters.name as cluster_name
		FROM mz_sources
		JOIN mz_schemas
			ON mz_sources.schema_id = mz_schemas.id
		JOIN mz_databases
			ON mz_schemas.database_id = mz_databases.id
		LEFT JOIN mz_connections
			ON mz_sources.connection_id = mz_connections.id
		LEFT JOIN mz_clusters
			ON mz_sources.cluster_id = mz_clusters.id`)

	if databaseName != "" {
		q.WriteString(fmt.Sprintf(`
		WHERE mz_databases.name = '%s'`, databaseName))

		if schemaName != "" {
			q.WriteString(fmt.Sprintf(` AND mz_schemas.name = '%s'`, schemaName))
		}
	}

	q.WriteString(`;`)
	return q.String()
}

func sourceRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*sql.DB)

	schemaName := d.Get("schema_name").(string)
	databaseName := d.Get("database_name").(string)
	q := sourceQuery(databaseName, schemaName)

	rows, err := conn.Query(q)
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
		var id, name, schema_name, database_name, source_type, size, envelope_type, connection_name, cluster_name string
		rows.Scan(&id, &name, &schema_name, &database_name, &source_type, &size, &envelope_type, &connection_name, &cluster_name)

		sourceMap := map[string]interface{}{}

		sourceMap["id"] = id
		sourceMap["name"] = name
		sourceMap["schema_name"] = schema_name
		sourceMap["database_name"] = database_name
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

	if databaseName != "" && schemaName != "" {
		id := fmt.Sprintf("%s|%s|sources", databaseName, schemaName)
		d.SetId(id)
	} else if databaseName != "" {
		id := fmt.Sprintf("%s|sources", databaseName)
		d.SetId(id)
	} else {
		d.SetId("sources")
	}

	return diags
}
