package datasources

import (
	"context"
	"testing"

	"github.com/MaterializeInc/terraform-materialize-provider/pkg/testhelpers"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/require"
)

func TestIndexDatasource(t *testing.T) {
	r := require.New(t)

	in := map[string]interface{}{
		"schema_name":   "schema",
		"database_name": "database",
	}
	d := schema.TestResourceDataRaw(t, Index().Schema, in)
	r.NotNil(d)

	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		ir := mock.NewRows([]string{"id", "name", "obj", "schema", "database"}).
			AddRow("u1", "index", "obj", "schema", "database")
		mock.ExpectQuery(`
			SELECT DISTINCT
				mz_indexes.id,
				mz_indexes.name,
				COALESCE\(sources.o_name, views.o_name, mviews.o_name\),
				COALESCE\(sources.s_name, views.s_name, mviews.s_name\),
				COALESCE\(sources.d_name, views.d_name, mviews.d_name\)
			FROM mz_indexes
			LEFT JOIN \(
				SELECT
					mz_sources.id,
					mz_sources.name AS o_name,
					mz_schemas.name AS s_name,
					mz_databases.name AS d_name
				FROM mz_sources
				JOIN mz_schemas
					ON mz_sources.schema_id = mz_schemas.id
				JOIN mz_databases
					ON mz_schemas.database_id = mz_databases.id
			\) sources
				ON mz_indexes.on_id = sources.id
			LEFT JOIN \(
				SELECT
					mz_views.id,
					mz_views.name AS o_name,
					mz_schemas.name AS s_name,
					mz_databases.name AS d_name
				FROM mz_views
				JOIN mz_schemas
					ON mz_views.schema_id = mz_schemas.id
				JOIN mz_databases
					ON mz_schemas.database_id = mz_databases.id
			\) views
				ON mz_indexes.on_id = sources.id
			LEFT JOIN \(
				SELECT
					mz_materialized_views.id,
					mz_materialized_views.name AS o_name,
					mz_schemas.name AS s_name,
					mz_databases.name AS d_name
				FROM mz_materialized_views
				JOIN mz_schemas
					ON mz_materialized_views.schema_id = mz_schemas.id
				JOIN mz_databases
					ON mz_schemas.database_id = mz_databases.id
			\) mviews
				ON mz_indexes.on_id = sources.id
			WHERE COALESCE\(sources.o_name, views.o_name, mviews.o_name\) IS NOT NULL
			AND mz_databases.name = 'database' AND mz_schemas.name = 'schema'`).WillReturnRows(ir)

		if err := indexRead(context.TODO(), d, db); err != nil {
			t.Fatal(err)
		}
	})

}
