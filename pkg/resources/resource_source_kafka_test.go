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

var inSourceKafka = map[string]interface{}{
	"name":              "source",
	"schema_name":       "schema",
	"database_name":     "database",
	"cluster_name":      "cluster",
	"size":              "small",
	"item_name":         "item",
	"kafka_connection":  []interface{}{map[string]interface{}{"name": "kafka_conn"}},
	"topic":             "topic",
	"include_key":       true,
	"include_headers":   true,
	"include_partition": true,
	"include_offset":    true,
	"include_timestamp": true,
	"format":            []interface{}{map[string]interface{}{"avro": []interface{}{map[string]interface{}{"value_strategy": "avro_key_fullname", "schema_registry_connection": []interface{}{map[string]interface{}{"name": "csr_conn", "database_name": "database", "schema_name": "schema"}}}}}},
	"envelope":          []interface{}{map[string]interface{}{"upsert": true}},
	"primary_key":       []interface{}{"key_1", "key_2", "key_3"},
	"start_offset":      []interface{}{1, 2, 3},
	"start_timestamp":   -1000,
}

func TestResourceSourceKafkaCreate(t *testing.T) {
	r := require.New(t)
	d := schema.TestResourceDataRaw(t, SourceKafka().Schema, inSourceKafka)
	r.NotNil(d)

	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		// Create
		mock.ExpectExec(
			`CREATE SOURCE "database"."schema"."source" IN CLUSTER "cluster" FROM KAFKA CONNECTION "database"."schema"."kafka_conn" \(TOPIC 'topic', START TIMESTAMP -1000\) FORMAT AVRO USING CONFLUENT SCHEMA REGISTRY CONNECTION "database"."schema"."csr_conn" VALUE STRATEGY avro_key_fullname PRIMARY KEY \(key_1, key_2, key_3\) NOT ENFORCED START OFFSET \[1, 2, 3\] INCLUDE KEY, HEADERS, PARTITION, OFFSET, TIMESTAMP ENVELOPE UPSERT WITH \(SIZE = 'small'\);`,
		).WillReturnResult(sqlmock.NewResult(1, 1))

		// Query Id
		ir := mock.NewRows([]string{"id", "name", "schema_name", "database_name", "source_type", "size", "envelope_type", "connection_name", "cluster_name"}).
			AddRow("u1", "source", "schema", "database", "kafka", "small", "JSON", "conn", "cluster")
		mock.ExpectQuery(`
			SELECT
				mz_sources.id,
				mz_sources.name,
				mz_schemas.name AS schema_name,
				mz_databases.name AS database_name,
				mz_sources.type AS source_type,
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
			WHERE mz_databases.name = 'database'
			AND mz_schemas.name = 'schema'
			AND mz_sources.name = 'source';`).WillReturnRows(ir)

		// Query Params
		ip := mock.NewRows([]string{"id", "name", "schema_name", "database_name", "source_type", "size", "envelope_type", "connection_name", "cluster_name"}).
			AddRow("u1", "source", "schema", "database", "kafka", "small", "JSON", "conn", "cluster")
		mock.ExpectQuery(readSource).WillReturnRows(ip)

		if err := sourceKafkaCreate(context.TODO(), d, db); err != nil {
			t.Fatal(err)
		}
	})

}
