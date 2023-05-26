package resources

import (
	"context"
	"testing"

	"github.com/MaterializeInc/terraform-provider-materialize/pkg/testhelpers"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/require"
)

var inSinkKafka = map[string]interface{}{
	"name":             "sink",
	"schema_name":      "schema",
	"database_name":    "database",
	"cluster_name":     "cluster",
	"size":             "small",
	"from":             []interface{}{map[string]interface{}{"name": "item", "schema_name": "public", "database_name": "database"}},
	"kafka_connection": []interface{}{map[string]interface{}{"name": "kafka_conn"}},
	"topic":            "topic",
	"key":              []interface{}{"key_1", "key_2"},
	"format":           []interface{}{map[string]interface{}{"avro": []interface{}{map[string]interface{}{"avro_key_fullname": "avro_key_fullname", "avro_value_fullname": "avro_value_fullname", "schema_registry_connection": []interface{}{map[string]interface{}{"name": "csr_conn", "database_name": "database", "schema_name": "schema"}}}}}},
	"envelope":         []interface{}{map[string]interface{}{"upsert": true}},
	"snapshot":         false,
}

func TestResourceSinkKafkaCreate(t *testing.T) {
	r := require.New(t)
	d := schema.TestResourceDataRaw(t, SinkKafka().Schema, inSinkKafka)
	r.NotNil(d)

	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		// Create
		mock.ExpectExec(
			`CREATE SINK "database"."schema"."sink" IN CLUSTER "cluster" FROM "database"."public"."item" INTO KAFKA CONNECTION "database"."schema"."kafka_conn" KEY \(key_1, key_2\) \(TOPIC 'topic'\) FORMAT AVRO USING CONFLUENT SCHEMA REGISTRY CONNECTION "database"."schema"."csr_conn" WITH \(AVRO KEY FULLNAME 'avro_key_fullname' AVRO VALUE FULLNAME 'avro_value_fullname'\) ENVELOPE UPSERT WITH \( SIZE = 'small' SNAPSHOT = false\);`,
		).WillReturnResult(sqlmock.NewResult(1, 1))

		// Query Id
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
			AND mz_schemas.name = 'schema'
			AND mz_sinks.name = 'sink';`).WillReturnRows(ir)

		// Query Params
		ip := mock.NewRows([]string{"id", "name", "schema_name", "database_name", "sink_type", "size", "envelope_type", "connection_name", "cluster_name"}).
			AddRow("u1", "sink", "schema", "database", "kafka", "small", "JSON", "conn", "cluster")
		mock.ExpectQuery(readSink).WillReturnRows(ip)

		if err := sinkKafkaCreate(context.TODO(), d, db); err != nil {
			t.Fatal(err)
		}
	})

}
