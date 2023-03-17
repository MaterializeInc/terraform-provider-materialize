package resources

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/jmoiron/sqlx"
)

func readSourceId(name, schema, database string) string {
	return fmt.Sprintf(`
		SELECT mz_sources.id
		FROM mz_sources
		JOIN mz_schemas
			ON mz_sources.schema_id = mz_schemas.id
		JOIN mz_databases
			ON mz_schemas.database_id = mz_databases.id
		LEFT JOIN mz_connections
			ON mz_sources.connection_id = mz_connections.id
		JOIN mz_clusters
			ON mz_sources.cluster_id = mz_clusters.id
		WHERE mz_sources.name = %s
		AND mz_schemas.name = %s
		AND mz_databases.name = %s;
	`, QuoteString(name), QuoteString(schema), QuoteString(database))
}

func readSourceParams(id string) string {
	return fmt.Sprintf(`
		SELECT
			mz_sources.name,
			mz_schemas.name,
			mz_databases.name,
			mz_sources.type,
			mz_sources.size,
			mz_connections.name as connection_name,
			mz_clusters.name as cluster_name
		FROM mz_sources
		JOIN mz_schemas
			ON mz_sources.schema_id = mz_schemas.id
		JOIN mz_databases
			ON mz_schemas.database_id = mz_databases.id
		LEFT JOIN mz_connections
			ON mz_sources.connection_id = mz_connections.id
		JOIN mz_clusters
			ON mz_sources.cluster_id = mz_clusters.id
		WHERE mz_sources.id = %s;`, QuoteString(id))
}

func SourceRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*sqlx.DB)
	i := d.Id()
	q := readSourceParams(i)

	var name, schema, database, source_type, size, connection_name, cluster_name *string
	if err := conn.QueryRowx(q).Scan(&name, &schema, &database, &source_type, &size, &connection_name, &cluster_name); err != nil {
		return diag.FromErr(err)
	}

	d.SetId(i)

	if err := d.Set("name", name); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("schema_name", schema); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("database_name", database); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("source_type", source_type); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("size", size); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("cluster_name", cluster_name); err != nil {
		return diag.FromErr(err)
	}

	qn := fmt.Sprintf("%s.%s.%s", *database, *schema, *name)
	if err := d.Set("qualified_name", qn); err != nil {
		return diag.FromErr(err)
	}

	return nil
}
