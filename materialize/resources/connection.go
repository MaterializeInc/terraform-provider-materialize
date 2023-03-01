package resources

import (
	"database/sql"
	"fmt"
)

func readConnectionId(name, schema, database string) string {
	return fmt.Sprintf(`
		SELECT mz_connections.id
		FROM mz_connections
		JOIN mz_schemas
			ON mz_connections.schema_id = mz_schemas.id
		JOIN mz_databases
			ON mz_schemas.database_id = mz_databases.id
		WHERE mz_connections.name = '%s'
		AND mz_schemas.name = '%s'
		AND mz_databases.name = '%s';
	`, name, schema, database)
}

func readConnectionParams(id string) string {
	return fmt.Sprintf(`
		SELECT
			mz_connections.name,
			mz_schemas.name,
			mz_databases.name,
			mz_connections.type
		FROM mz_connections
		JOIN mz_schemas
			ON mz_connections.schema_id = mz_schemas.id
		JOIN mz_databases
			ON mz_schemas.database_id = mz_databases.id
		WHERE mz_connections.id = '%s';`, id)
}

//lint:ignore U1000 Ignore unused function temporarily for debugging
type _connection struct {
	name            sql.NullString `db:"name"`
	schema_name     sql.NullString `db:"schema_name"`
	database_name   sql.NullString `db:"database_name"`
	connection_type sql.NullString `db:"connection_type"`
}
