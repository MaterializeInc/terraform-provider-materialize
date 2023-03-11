package datasources

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMaterializedmaterializedViewQuery(t *testing.T) {
	r := require.New(t)
	b := materializedViewQuery("", "")
	r.Equal(`
		SELECT
			mz_materialized_views.id,
			mz_materialized_views.name,
			mz_schemas.name,
			mz_databases.name,
		SELECT mz_materialized_views.id
		FROM mz_materialized_views
		JOIN mz_schemas
			ON mz_materialized_views.schema_id = mz_schemas.id
		JOIN mz_databases
			ON mz_schemas.database_id = mz_databases.id;`, b)
}

func TestMaterializedViewDatabaseQuery(t *testing.T) {
	r := require.New(t)
	b := materializedViewQuery("database", "")
	r.Equal(`
		SELECT
			mz_materialized_views.id,
			mz_materialized_views.name,
			mz_schemas.name,
			mz_databases.name,
		SELECT mz_materialized_views.id
		FROM mz_materialized_views
		JOIN mz_schemas
			ON mz_materialized_views.schema_id = mz_schemas.id
		JOIN mz_databases
			ON mz_schemas.database_id = mz_databases.id
		WHERE mz_databases.name = 'database';`, b)
}

func TestMaterializedViewSchemaDatabaseQuery(t *testing.T) {
	r := require.New(t)
	b := materializedViewQuery("database", "schema")
	r.Equal(`
		SELECT
			mz_materialized_views.id,
			mz_materialized_views.name,
			mz_schemas.name,
			mz_databases.name,
		SELECT mz_materialized_views.id
		FROM mz_materialized_views
		JOIN mz_schemas
			ON mz_materialized_views.schema_id = mz_schemas.id
		JOIN mz_databases
			ON mz_schemas.database_id = mz_databases.id
		WHERE mz_databases.name = 'database' AND mz_schemas.name = 'schema';`, b)
}
