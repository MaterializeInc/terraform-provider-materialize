package resources

import (
	"database/sql"
	"fmt"
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
			mz_sinks.envelope_type,
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

//lint:ignore U1000 Ignore unused function temporarily for debugging
type _sink struct {
	name            sql.NullString `db:"name"`
	schema_name     sql.NullString `db:"schema_name"`
	database_name   sql.NullString `db:"database_name"`
	source_type     sql.NullString `db:"source_type"`
	size            sql.NullString `db:"size"`
	envelope_type   sql.NullString `db:"envelope_type"`
	connection_name sql.NullString `db:"connection_name"`
	cluster_name    sql.NullString `db:"cluster_name"`
}
