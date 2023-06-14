package resources

import (
	"context"
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/MaterializeInc/terraform-provider-materialize/pkg/testhelpers"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/require"
)

var readSink = `
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
WHERE mz_sinks.id = 'u1';`

func TestResourceSinkUpdate(t *testing.T) {
	r := require.New(t)
	d := schema.TestResourceDataRaw(t, SinkKafka().Schema, inSinkKafka)

	// Set current state
	d.SetId("u1")
	d.Set("name", "old_sink")
	d.Set("size", "medium")
	r.NotNil(d)

	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		mock.ExpectExec(`ALTER SINK "database"."schema"."old_sink" SET \(SIZE = 'small'\);`).WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectExec(`ALTER SINK "database"."schema"."" RENAME TO "sink";`).WillReturnResult(sqlmock.NewResult(1, 1))

		// Query Params
		ip := mock.NewRows([]string{"id", "name", "schema_name", "database_name", "sink_type", "size", "envelope_type", "connection_name", "cluster_name"}).
			AddRow("u1", "sink", "schema", "database", "kafka", "small", "JSON", "conn", "cluster")
		mock.ExpectQuery(readSink).WillReturnRows(ip)

		if err := sinkUpdate(context.TODO(), d, db); err != nil {
			t.Fatal(err)
		}
	})

}

func TestResourceSinkDelete(t *testing.T) {
	r := require.New(t)

	in := map[string]interface{}{
		"name":          "sink",
		"schema_name":   "schema",
		"database_name": "database",
	}
	d := schema.TestResourceDataRaw(t, SinkKafka().Schema, in)
	r.NotNil(d)

	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		mock.ExpectExec(`DROP SINK "database"."schema"."sink";`).WillReturnResult(sqlmock.NewResult(1, 1))

		if err := sinkDelete(context.TODO(), d, db); err != nil {
			t.Fatal(err)
		}
	})
}
