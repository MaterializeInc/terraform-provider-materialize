package resources

import (
	"context"
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/require"
)

func TestResourceSinkKafkaCreate(t *testing.T) {
	r := require.New(t)

	in := map[string]interface{}{
		"name":             "sink",
		"schema_name":      "schema",
		"database_name":    "database",
		"cluster_name":     "cluster",
		"size":             "small",
		"item_name":        "item",
		"kafka_connection": "kafka_conn",
		"topic":            "topic",
		// "key":                        []interface{}{"key_1", "key_2"},
		"format":                     "AVRO",
		"envelope":                   "UPSERT",
		"schema_registry_connection": "csr_conn",
		"avro_key_fullname":          "avro_key_fullname",
		"avro_value_fullname":        "avro_value_fullname",
		"snapshot":                   false,
	}
	d := schema.TestResourceDataRaw(t, SinkKafka().Schema, in)
	r.NotNil(d)

	WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		// Create
		mock.ExpectExec(
			`CREATE SINK database.schema.sink IN CLUSTER cluster FROM item INTO KAFKA CONNECTION kafka_conn \(TOPIC 'topic'\) FORMAT AVRO USING CONFLUENT SCHEMA REGISTRY CONNECTION csr_conn WITH \(AVRO KEY FULLNAME avro_key_fullname AVRO VALUE FULLNAME avro_value_fullname\) ENVELOPE UPSERT WITH \( SIZE = 'small' SNAPSHOT = false\);`,
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
		ip := sqlmock.NewRows([]string{"name", "schema", "database", "sink_type", "size", "connection_name", "cluster_name"}).
			AddRow("conn", "schema", "database", "sink_type", "small", "conn", "cluster")
		mock.ExpectQuery(`
			SELECT
				mz_sinks.name,
				mz_schemas.name,
				mz_databases.name,
				mz_sinks.type,
				mz_sinks.size,
				mz_connections.name as connection_name,
				mz_clusters.name as cluster_name
			FROM mz_sinks
			JOIN mz_schemas
				ON mz_sinks.schema_id = mz_schemas.id
			JOIN mz_databases
				ON mz_schemas.database_id = mz_databases.id
			LEFT JOIN mz_connections
				ON mz_sinks.connection_id = mz_connections.id
			JOIN mz_clusters
				ON mz_sinks.cluster_id = mz_clusters.id
			WHERE mz_sinks.id = 'u1';`).WillReturnRows(ip)

		if err := sinkKafkaCreate(context.TODO(), d, db); err != nil {
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

	WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		mock.ExpectExec(`DROP SINK database.schema.sink;`).WillReturnResult(sqlmock.NewResult(1, 1))

		if err := sinkKafkaDelete(context.TODO(), d, db); err != nil {
			t.Fatal(err)
		}
	})

}

func TestSinkKafkaCreateQuery(t *testing.T) {
	r := require.New(t)

	bs := newSinkKafkaBuilder("sink", "schema", "database")
	bs.Size("xsmall")
	bs.ItemName("schema.table")
	r.Equal(`CREATE SINK database.schema.sink FROM schema.table WITH ( SIZE = 'xsmall' SNAPSHOT = false);`, bs.Create())

	bc := newSinkKafkaBuilder("sink", "schema", "database")
	bc.ClusterName("cluster")
	bc.ItemName("schema.table")
	bc.Snapshot(true)
	r.Equal(`CREATE SINK database.schema.sink IN CLUSTER cluster FROM schema.table;`, bc.Create())
}

func TestSinkKafkaCreateParamsQuery(t *testing.T) {
	r := require.New(t)
	b := newSinkKafkaBuilder("sink", "schema", "database")
	b.Size("xsmall")
	b.ItemName("schema.table")
	b.KafkaConnection("kafka_connection")
	b.Topic("test_avro_topic")
	b.Key([]string{"key_1", "key_2"})
	b.Format("AVRO")
	b.SchemaRegistryConnection("csr_connection")
	b.Envelope("UPSERT")
	b.Snapshot(false)
	r.Equal(`CREATE SINK database.schema.sink FROM schema.table INTO KAFKA CONNECTION kafka_connection KEY (key_1, key_2) (TOPIC 'test_avro_topic') FORMAT AVRO USING CONFLUENT SCHEMA REGISTRY CONNECTION csr_connection ENVELOPE UPSERT WITH ( SIZE = 'xsmall' SNAPSHOT = false);`, b.Create())
}

func TestSinkKafkaReadIdQuery(t *testing.T) {
	r := require.New(t)
	b := newSinkKafkaBuilder("sink", "schema", "database")
	r.Equal(`
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
	`, b.ReadId())
}

func TestSinkKafkaRenameQuery(t *testing.T) {
	r := require.New(t)
	b := newSinkKafkaBuilder("sink", "schema", "database")
	r.Equal(`ALTER SINK database.schema.sink RENAME TO database.schema.new_sink;`, b.Rename("new_sink"))
}

func TestSinkKafkaResizeQuery(t *testing.T) {
	r := require.New(t)
	b := newSinkKafkaBuilder("sink", "schema", "database")
	r.Equal(`ALTER SINK database.schema.sink SET (SIZE = 'xlarge');`, b.UpdateSize("xlarge"))
}

func TestSinkKafkaDropQuery(t *testing.T) {
	r := require.New(t)
	b := newSinkKafkaBuilder("sink", "schema", "database")
	r.Equal(`DROP SINK database.schema.sink;`, b.Drop())
}

func TestSinkKafkaReadParamsQuery(t *testing.T) {
	r := require.New(t)
	b := readSinkParams("u1")
	r.Equal(`
		SELECT
			mz_sinks.name,
			mz_schemas.name,
			mz_databases.name,
			mz_sinks.type,
			mz_sinks.size,
			mz_connections.name as connection_name,
			mz_clusters.name as cluster_name
		FROM mz_sinks
		JOIN mz_schemas
			ON mz_sinks.schema_id = mz_schemas.id
		JOIN mz_databases
			ON mz_schemas.database_id = mz_databases.id
		LEFT JOIN mz_connections
			ON mz_sinks.connection_id = mz_connections.id
		JOIN mz_clusters
			ON mz_sinks.cluster_id = mz_clusters.id
		WHERE mz_sinks.id = 'u1';`, b)
}
