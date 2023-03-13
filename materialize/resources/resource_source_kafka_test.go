package resources

import (
	"context"
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/require"
)

func TestResourceSourceKafkaCreate(t *testing.T) {
	r := require.New(t)

	in := map[string]interface{}{
		"name":                       "source",
		"schema_name":                "schema",
		"database_name":              "database",
		"cluster_name":               "cluster",
		"size":                       "small",
		"item_name":                  "item",
		"kafka_connection":           "kafka_conn",
		"topic":                      "topic",
		"include_key":                "key",
		"include_headers":            true,
		"include_partition":          "parition",
		"include_offset":             "offset",
		"include_timestamp":          "timestamp",
		"format":                     "AVRO",
		"key_format":                 "AVRO",
		"envelope":                   "UPSERT",
		"schema_registry_connection": "csr_conn",
		"value_strategy":             "avro_key_fullname",
		// "primary_key":                []interface{}{"key_1", "key_2", "key_3"},
		// "start_offset":               []interface{}{1, 2, 3},
		"start_timestamp": -1000,
	}
	d := schema.TestResourceDataRaw(t, SourceKafka().Schema, in)
	r.NotNil(d)

	WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		// Create
		mock.ExpectExec(
			`CREATE SOURCE "database"."schema"."source" IN CLUSTER cluster FROM KAFKA CONNECTION kafka_conn \(TOPIC 'topic'\) KEY FORMAT AVRO VALUE FORMAT AVRO USING CONFLUENT SCHEMA REGISTRY CONNECTION csr_conn START TIMESTAMP -1000 VALUE STRATEGY avro_key_fullname INCLUDE key, HEADERS, parition, offset, timestamp ENVELOPE UPSERT WITH \(SIZE = 'small'\);`,
		).WillReturnResult(sqlmock.NewResult(1, 1))

		// Query Id
		ir := mock.NewRows([]string{"id"}).AddRow("u1")
		mock.ExpectQuery(`
			SELECT mz_sources.id
			FROM mz_sources
			JOIN mz_schemas
				ON mz_sources.schema_id = mz_schemas.id
			JOIN mz_databases
				ON mz_schemas.database_id = mz_databases.id
			LEFT JOIN mz_connections
				ON mz_sources.connection_id = mz_connections.id
			JOIN mz_clusters
				ON mz_sources.cluster_id = mz_clusters.id
			WHERE mz_sources.name = 'source'
			AND mz_schemas.name = 'schema'
			AND mz_databases.name = 'database';
		`).WillReturnRows(ir)

		// Query Params
		ip := sqlmock.NewRows([]string{"name", "schema", "database", "source_type", "size", "connection_name", "cluster_name"}).
			AddRow("conn", "schema", "database", "source_type", "small", "conn", "cluster")
		mock.ExpectQuery(`
			SELECT
				mz_sources.name,
				mz_schemas.name,
				mz_databases.name,
				mz_sources.type,
				mz_sources.size,
				mz_connections.name as connection_name,
				mz_clusters.name as cluster_name
			FROM mz_sources
			JOIN mz_schemas
				ON mz_sources.schema_id = mz_schemas.id
			JOIN mz_databases
				ON mz_schemas.database_id = mz_databases.id
			LEFT JOIN mz_connections
				ON mz_sources.connection_id = mz_connections.id
			JOIN mz_clusters
				ON mz_sources.cluster_id = mz_clusters.id
			WHERE mz_sources.id = 'u1';`).WillReturnRows(ip)

		if err := sourceKafkaCreate(context.TODO(), d, db); err != nil {
			t.Fatal(err)
		}
	})

}

func TestResourceSourceKafkaDelete(t *testing.T) {
	r := require.New(t)

	in := map[string]interface{}{
		"name":          "source",
		"schema_name":   "schema",
		"database_name": "database",
	}
	d := schema.TestResourceDataRaw(t, SourceKafka().Schema, in)
	r.NotNil(d)

	WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		mock.ExpectExec(`DROP SOURCE "database"."schema"."source";`).WillReturnResult(sqlmock.NewResult(1, 1))

		if err := sourceKafkaDelete(context.TODO(), d, db); err != nil {
			t.Fatal(err)
		}
	})

}

func TestResourceSourceKafkaCreateQuery(t *testing.T) {
	r := require.New(t)

	bs := newSourceKafkaBuilder("source", "schema", "database")
	bs.Size("xsmall")
	bs.KafkaConnection("kafka_connection")
	bs.Topic("events")
	bs.Format("TEXT")
	r.Equal(`CREATE SOURCE "database"."schema"."source" FROM KAFKA CONNECTION kafka_connection (TOPIC 'events') FORMAT TEXT WITH (SIZE = 'xsmall');`, bs.Create())

	bc := newSourceKafkaBuilder("source", "schema", "database")
	bc.ClusterName("cluster")
	bc.KafkaConnection("kafka_connection")
	bc.Topic("events")
	bc.Format("TEXT")
	r.Equal(`CREATE SOURCE "database"."schema"."source" IN CLUSTER cluster FROM KAFKA CONNECTION kafka_connection (TOPIC 'events') FORMAT TEXT;`, bc.Create())
}

func TestResourceSourceKafkaCreateParamsQuery(t *testing.T) {
	r := require.New(t)
	b := newSourceKafkaBuilder("source", "schema", "database")
	b.Size("xsmall")
	b.KafkaConnection("kafka_connection")
	b.Topic("events")
	b.Format("AVRO")
	b.IncludeKey("KEY")
	b.IncludePartition("PARTITION")
	b.IncludeOffset("OFFSET")
	b.IncludeTimestamp("TIMESTAMP")
	b.SchemaRegistryConnection("csr_connection")
	b.Envelope("UPSERT")
	r.Equal(`CREATE SOURCE "database"."schema"."source" FROM KAFKA CONNECTION kafka_connection (TOPIC 'events') FORMAT AVRO USING CONFLUENT SCHEMA REGISTRY CONNECTION csr_connection INCLUDE KEY, PARTITION, OFFSET, TIMESTAMP ENVELOPE UPSERT WITH (SIZE = 'xsmall');`, b.Create())
}

func TestResourceSourceKafkaReadIdQuery(t *testing.T) {
	r := require.New(t)
	b := newSourceKafkaBuilder("source", "schema", "database")
	r.Equal(`
		SELECT mz_sources.id
		FROM mz_sources
		JOIN mz_schemas
			ON mz_sources.schema_id = mz_schemas.id
		JOIN mz_databases
			ON mz_schemas.database_id = mz_databases.id
		LEFT JOIN mz_connections
			ON mz_sources.connection_id = mz_connections.id
		JOIN mz_clusters
			ON mz_sources.cluster_id = mz_clusters.id
		WHERE mz_sources.name = 'source'
		AND mz_schemas.name = 'schema'
		AND mz_databases.name = 'database';
	`, b.ReadId())
}

func TestResourceSourceKafkaRenameQuery(t *testing.T) {
	r := require.New(t)
	b := newSourceKafkaBuilder("source", "schema", "database")
	r.Equal(`ALTER SOURCE "database"."schema"."source" RENAME TO "database"."schema"."new_source";`, b.Rename("new_source"))
}

func TestResourceSourceKafkaResizeQuery(t *testing.T) {
	r := require.New(t)
	b := newSourceKafkaBuilder("source", "schema", "database")
	r.Equal(`ALTER SOURCE "database"."schema"."source" SET (SIZE = 'xlarge');`, b.UpdateSize("xlarge"))
}

func TestResourceSourceKafkaDropQuery(t *testing.T) {
	r := require.New(t)
	b := newSourceKafkaBuilder("source", "schema", "database")
	r.Equal(`DROP SOURCE "database"."schema"."source";`, b.Drop())
}

func TestResourceSourceKafkaReadParamsQuery(t *testing.T) {
	r := require.New(t)
	b := readSourceParams("u1")
	r.Equal(`
		SELECT
			mz_sources.name,
			mz_schemas.name,
			mz_databases.name,
			mz_sources.type,
			mz_sources.size,
			mz_connections.name as connection_name,
			mz_clusters.name as cluster_name
		FROM mz_sources
		JOIN mz_schemas
			ON mz_sources.schema_id = mz_schemas.id
		JOIN mz_databases
			ON mz_schemas.database_id = mz_databases.id
		LEFT JOIN mz_connections
			ON mz_sources.connection_id = mz_connections.id
		JOIN mz_clusters
			ON mz_sources.cluster_id = mz_clusters.id
		WHERE mz_sources.id = 'u1';`, b)
}
