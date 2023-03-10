package datasources

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestViewQuery(t *testing.T) {
	r := require.New(t)
	b := viewQuery("", "")
	r.Equal(`
		SELECT
			mz_views.id,
			mz_views.name,
			mz_schemas.name,
			mz_databases.name,
		SELECT mz_views.id
		FROM mz_views
		JOIN mz_schemas
			ON mz_views.schema_id = mz_schemas.id
		JOIN mz_databases
			ON mz_schemas.database_id = mz_databases.id;`, b)
}

func TestViewDatabaseQuery(t *testing.T) {
	r := require.New(t)
	b := viewQuery("database", "")
	r.Equal(`
		SELECT
			mz_views.id,
			mz_views.name,
			mz_schemas.name,
			mz_databases.name,
		SELECT mz_views.id
		FROM mz_views
		JOIN mz_schemas
			ON mz_views.schema_id = mz_schemas.id
		JOIN mz_databases
			ON mz_schemas.database_id = mz_databases.id
		WHERE mz_databases.name = 'database';`, b)
}

func TestViewSchemaDatabaseQuery(t *testing.T) {
	r := require.New(t)
	b := viewQuery("database", "schema")
	r.Equal(`
		SELECT
			mz_views.id,
			mz_views.name,
			mz_schemas.name,
			mz_databases.name,
		SELECT mz_views.id
		FROM mz_views
		JOIN mz_schemas
			ON mz_views.schema_id = mz_schemas.id
		JOIN mz_databases
			ON mz_schemas.database_id = mz_databases.id
		WHERE mz_databases.name = 'database' AND mz_schemas.name = 'schema';`, b)
}
