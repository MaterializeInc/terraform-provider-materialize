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
	// "key":                        []interface{}{"key_1", "key_2"},
	"format":   []interface{}{map[string]interface{}{"avro": []interface{}{map[string]interface{}{"avro_key_fullname": "avro_key_fullname", "avro_value_fullname": "avro_value_fullname", "schema_registry_connection": []interface{}{map[string]interface{}{"name": "csr_conn", "database_name": "database", "schema_name": "schema"}}}}}},
	"envelope": []interface{}{map[string]interface{}{"upsert": true}},
	"snapshot": false,
}

func TestResourceSinkKafkaCreate(t *testing.T) {
	r := require.New(t)
	d := schema.TestResourceDataRaw(t, SinkKafka().Schema, inSinkKafka)
	r.NotNil(d)

	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		// Create
		mock.ExpectExec(
			`CREATE SINK "database"."schema"."sink" IN CLUSTER "cluster" FROM "database"."public"."item" INTO KAFKA CONNECTION "database"."schema"."kafka_conn" \(TOPIC 'topic'\) FORMAT AVRO USING CONFLUENT SCHEMA REGISTRY CONNECTION "database"."schema"."csr_conn" WITH \(AVRO KEY FULLNAME 'avro_key_fullname' AVRO VALUE FULLNAME 'avro_value_fullname'\) ENVELOPE UPSERT WITH \( SIZE = 'small' SNAPSHOT = false\);`,
		).WillReturnResult(sqlmock.NewResult(1, 1))

		// Query Id
		ir := mock.NewRows([]string{"id"}).AddRow("u1")
		mock.ExpectQuery(`
			SELECT mz_sinks.id
			FROM mz_sinks
			JOIN mz_schemas
				ON mz_sinks.schema_id = mz_schemas.id
			JOIN mz_databases
				ON mz_schemas.database_id = mz_databases.id
			LEFT JOIN mz_connections
				ON mz_sinks.connection_id = mz_connections.id
			JOIN mz_clusters
				ON mz_sinks.cluster_id = mz_clusters.id
			WHERE mz_sinks.name = 'sink'
			AND mz_schemas.name = 'schema'
			AND mz_databases.name = 'database';
		`).WillReturnRows(ir)

		// Query Params
		ip := sqlmock.NewRows([]string{"name", "schema", "database", "size", "connection_name", "cluster_name"}).
			AddRow("sink", "schema", "database", "small", "conn", "cluster")
		mock.ExpectQuery(readSink).WillReturnRows(ip)

		if err := sinkKafkaCreate(context.TODO(), d, db); err != nil {
			t.Fatal(err)
		}
	})

}

func TestResourceSinkKafkaUpdate(t *testing.T) {
	r := require.New(t)
	d := schema.TestResourceDataRaw(t, SinkKafka().Schema, inSinkKafka)

	// Set current state
	d.SetId("u1")
	d.Set("name", "old_sink")
	d.Set("size", "medium")
	r.NotNil(d)

	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		mock.ExpectExec(`ALTER SINK "database"."schema"."old_sink" SET \(SIZE = 'small'\);`).WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectExec(`ALTER SINK "database"."schema"."old_sink" RENAME TO "database"."schema"."sink";`).WillReturnResult(sqlmock.NewResult(1, 1))

		// Query Params
		ip := sqlmock.NewRows([]string{"name", "schema", "database", "size", "connection_name", "cluster_name"}).
			AddRow("sink", "schema", "database", "small", "conn", "cluster")
		mock.ExpectQuery(readSink).WillReturnRows(ip)

		if err := sinkKafkaUpdate(context.TODO(), d, db); err != nil {
			t.Fatal(err)
		}
	})

}

func TestResourceSinkKafkaDelete(t *testing.T) {
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

		if err := sinkKafkaDelete(context.TODO(), d, db); err != nil {
			t.Fatal(err)
		}
	})

}
