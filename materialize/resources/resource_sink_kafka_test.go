package resources

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestResourceSinkKafkaCreate(t *testing.T) {
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

func TestResourceSinkKafkaCreateParams(t *testing.T) {
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

func TestResourceSinkKafkaReadId(t *testing.T) {
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

func TestResourceSinkKafkaRename(t *testing.T) {
	r := require.New(t)
	b := newSinkKafkaBuilder("sink", "schema", "database")
	r.Equal(`ALTER SINK database.schema.sink RENAME TO database.schema.new_sink;`, b.Rename("new_sink"))
}

func TestResourceSinkKafkaResize(t *testing.T) {
	r := require.New(t)
	b := newSinkKafkaBuilder("sink", "schema", "database")
	r.Equal(`ALTER SINK database.schema.sink SET (SIZE = 'xlarge');`, b.UpdateSize("xlarge"))
}

func TestResourceSinkKafkaDrop(t *testing.T) {
	r := require.New(t)
	b := newSinkKafkaBuilder("sink", "schema", "database")
	r.Equal(`DROP SINK database.schema.sink;`, b.Drop())
}

func TestResourceSinkKafkaReadParams(t *testing.T) {
	r := require.New(t)
	b := readSinkParams("u1")
	r.Equal(`
		SELECT
			mz_sinks.name,
			mz_schemas.name,
			mz_databases.name,
			mz_sinks.type,
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
		JOIN mz_clusters
			ON mz_sinks.cluster_id = mz_clusters.id
		WHERE mz_sinks.id = 'u1';`, b)
}
