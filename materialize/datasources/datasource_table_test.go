package datasources

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestQueryQuery(t *testing.T) {
	r := require.New(t)
	b := tableQuery("", "")
	r.Equal(`
		SELECT
			mz_tables.id,
			mz_tables.name,
			mz_schemas.name,
			mz_databases.name,
		SELECT mz_tables.id
		FROM mz_tables
		JOIN mz_schemas
			ON mz_tables.schema_id = mz_schemas.id
		JOIN mz_databases
			ON mz_schemas.database_id = mz_databases.id;`, b)
}

func TestQueryDatabaseQuery(t *testing.T) {
	r := require.New(t)
	b := tableQuery("database", "")
	r.Equal(`
		SELECT
			mz_tables.id,
			mz_tables.name,
			mz_schemas.name,
			mz_databases.name,
		SELECT mz_tables.id
		FROM mz_tables
		JOIN mz_schemas
			ON mz_tables.schema_id = mz_schemas.id
		JOIN mz_databases
			ON mz_schemas.database_id = mz_databases.id
		WHERE mz_databases.name = 'database';`, b)
}

func TestQuerySchemaDatabaseQuery(t *testing.T) {
	r := require.New(t)
	b := tableQuery("database", "schema")
	r.Equal(`
		SELECT
			mz_tables.id,
			mz_tables.name,
			mz_schemas.name,
			mz_databases.name,
		SELECT mz_tables.id
		FROM mz_tables
		JOIN mz_schemas
			ON mz_tables.schema_id = mz_schemas.id
		JOIN mz_databases
			ON mz_schemas.database_id = mz_databases.id
		WHERE mz_databases.name = 'database' AND mz_schemas.name = 'schema';`, b)
}
