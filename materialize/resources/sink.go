package resources

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/jmoiron/sqlx"
)

func readSinkId(name, schema, database string) string {
	return fmt.Sprintf(`
		SELECT mz_sinks.id
		FROM mz_sinks
		JOIN mz_schemas
			ON mz_sinks.schema_id = mz_schemas.id
		JOIN mz_databases
			ON mz_schemas.database_id = mz_databases.id
		LEFT JOIN mz_connections
			ON mz_sinks.connection_id = mz_connections.id
		JOIN mz_clusters
			ON mz_sinks.cluster_id = mz_clusters.id
		WHERE mz_sinks.name = '%s'
		AND mz_schemas.name = '%s'
		AND mz_databases.name = '%s';
	`, name, schema, database)
}

func readSinkParams(id string) string {
	return fmt.Sprintf(`
		SELECT
			mz_sinks.name,
			mz_schemas.name,
			mz_databases.name,
			mz_sinks.type,
			mz_sinks.size,
			mz_connections.name as connection_name,
			mz_clusters.name as cluster_name
		FROM mz_sinks
		JOIN mz_schemas
			ON mz_sinks.schema_id = mz_schemas.id
		JOIN mz_databases
			ON mz_schemas.database_id = mz_databases.id
		LEFT JOIN mz_connections
			ON mz_sinks.connection_id = mz_connections.id
		JOIN mz_clusters
			ON mz_sinks.cluster_id = mz_clusters.id
		WHERE mz_sinks.id = '%s';`, id)
}

func SinkRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*sqlx.DB)
	i := d.Id()
	q := readSinkParams(i)

	var name, schema, database, sink_type, size, connection_name, cluster_name *string
	if err := conn.QueryRowx(q).Scan(&name, &schema, &database, &sink_type, &size, &connection_name, &cluster_name); err != nil {
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

	if err := d.Set("sink_type", sink_type); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("size", size); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("cluster_name", cluster_name); err != nil {
		return diag.FromErr(err)
	}

	setQualifiedName(d)
	return nil
}
