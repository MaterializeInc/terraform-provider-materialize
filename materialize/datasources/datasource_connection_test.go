package datasources

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestConnectionQuery(t *testing.T) {
	r := require.New(t)
	b := connectionQuery("", "")
	r.Equal(`
		SELECT
			mz_connections.id,
			mz_connections.name,
			mz_schemas.name AS schema_name,
			mz_databases.name AS database_name,
			mz_connections.type
		FROM mz_connections
		JOIN mz_schemas
			ON mz_connections.schema_id = mz_schemas.id
		JOIN mz_databases
			ON mz_schemas.database_id = mz_databases.id;`, b)
}

func TestConnectionDatabaseQuery(t *testing.T) {
	r := require.New(t)
	b := connectionQuery("database", "")
	r.Equal(`
		SELECT
			mz_connections.id,
			mz_connections.name,
			mz_schemas.name AS schema_name,
			mz_databases.name AS database_name,
			mz_connections.type
		FROM mz_connections
		JOIN mz_schemas
			ON mz_connections.schema_id = mz_schemas.id
		JOIN mz_databases
			ON mz_schemas.database_id = mz_databases.id
		WHERE mz_databases.name = 'database';`, b)
}

func TestConnectionSchemaDatabaseQuery(t *testing.T) {
	r := require.New(t)
	b := connectionQuery("database", "schema")
	r.Equal(`
		SELECT
			mz_connections.id,
			mz_connections.name,
			mz_schemas.name AS schema_name,
			mz_databases.name AS database_name,
			mz_connections.type
		FROM mz_connections
		JOIN mz_schemas
			ON mz_connections.schema_id = mz_schemas.id
		JOIN mz_databases
			ON mz_schemas.database_id = mz_databases.id
		WHERE mz_databases.name = 'database' AND mz_schemas.name = 'schema';`, b)
}
