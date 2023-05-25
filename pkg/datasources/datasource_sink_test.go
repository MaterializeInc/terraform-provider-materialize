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

func TestSinkDatasource(t *testing.T) {
	r := require.New(t)

	in := map[string]interface{}{
		"schema_name":   "schema",
		"database_name": "database",
	}
	d := schema.TestResourceDataRaw(t, Sink().Schema, in)
	r.NotNil(d)

	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		ir := mock.NewRows([]string{"id", "name", "schema_name", "database_name", "sink_type", "size", "envelope_type", "connection_name", "cluster_name"}).
			AddRow("u1", "sink", "schema", "database", "kafka", "small", "JSON", "conn", "cluster")
		mock.ExpectQuery(`
		SELECT
			mz_sinks.id,
			mz_sinks.name,
			mz_schemas.name AS schema_name,
			mz_databases.name AS database_name,
			mz_sinks.type AS sink_type,
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
		LEFT JOIN mz_clusters
			ON mz_sinks.cluster_id = mz_clusters.id
		WHERE mz_databases.name = 'database'
		AND mz_schemas.name = 'schema';`).WillReturnRows(ir)

		if err := sinkRead(context.TODO(), d, db); err != nil {
			t.Fatal(err)
		}
	})

}
