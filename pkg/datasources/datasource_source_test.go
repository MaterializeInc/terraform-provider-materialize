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

func TestSourceDatasource(t *testing.T) {
	r := require.New(t)

	in := map[string]interface{}{
		"schema_name":   "schema",
		"database_name": "database",
	}
	d := schema.TestResourceDataRaw(t, Source().Schema, in)
	r.NotNil(d)

	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		ir := mock.NewRows([]string{"id", "name", "schema", "database"}).
			AddRow("u1", "view", "schema", "database")
		mock.ExpectQuery(`
		SELECT
			mz_sources.id,
			mz_sources.name,
			mz_schemas.name,
			mz_databases.name,
			mz_sources.type,
			mz_sources.size,
			mz_sources.envelope_type,
			mz_connections.name as connection_name,
			mz_clusters.name as cluster_name
		FROM mz_sources
		JOIN mz_schemas
			ON mz_sources.schema_id = mz_schemas.id
		JOIN mz_databases
			ON mz_schemas.database_id = mz_databases.id
		LEFT JOIN mz_connections
			ON mz_sources.connection_id = mz_connections.id
		LEFT JOIN mz_clusters
			ON mz_sources.cluster_id = mz_clusters.id
		WHERE mz_databases.name = 'database' AND mz_schemas.name = 'schema'`).WillReturnRows(ir)

		if err := sourceRead(context.TODO(), d, db); err != nil {
			t.Fatal(err)
		}
	})

}
