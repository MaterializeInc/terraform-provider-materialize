package datasources

import (
	"context"
	"testing"

	"github.com/MaterializeInc/terraform-provider-materialize/pkg/testhelpers"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/require"
)

func TestMaterializedViewDatasource(t *testing.T) {
	r := require.New(t)

	in := map[string]interface{}{
		"schema_name":   "schema",
		"database_name": "database",
	}
	d := schema.TestResourceDataRaw(t, MaterializedView().Schema, in)
	r.NotNil(d)

	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		ir := mock.NewRows([]string{"id", "materialized_view_name", "schema_name", "database_name", "cluster_name"}).
			AddRow("u1", "view", "schema", "database", "cluster")
		mock.ExpectQuery(`
			SELECT
				mz_materialized_views.id,
				mz_materialized_views.name AS materialized_view_name,
				mz_schemas.name AS schema_name,
				mz_databases.name AS database_name,
				mz_clusters.name AS cluster_name
			FROM mz_materialized_views
			JOIN mz_schemas
				ON mz_materialized_views.schema_id = mz_schemas.id
			JOIN mz_databases
				ON mz_schemas.database_id = mz_databases.id
			LEFT JOIN mz_clusters
				ON mz_materialized_views.cluster_id = mz_clusters.id
			WHERE mz_databases.name = 'database'
			AND mz_schemas.name = 'schema'`).WillReturnRows(ir)

		if err := materializedViewRead(context.TODO(), d, db); err != nil {
			t.Fatal(err)
		}
	})

}
