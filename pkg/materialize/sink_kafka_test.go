package materialize

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSinkKafkaCreateQuery(t *testing.T) {
	r := require.New(t)

	bs := NewSinkKafkaBuilder("sink", "schema", "database")
	bs.Size("xsmall")
	bs.From(IdentifierSchemaStruct{Name: "table", SchemaName: "schema", DatabaseName: "database"})
	r.Equal(`CREATE SINK "database"."schema"."sink" FROM "database"."schema"."table" WITH ( SIZE = 'xsmall' SNAPSHOT = false);`, bs.Create())

	bc := NewSinkKafkaBuilder("sink", "schema", "database")
	bc.ClusterName("cluster")
	bc.From(IdentifierSchemaStruct{Name: "table", SchemaName: "schema", DatabaseName: "database"})
	bc.Snapshot(true)
	r.Equal(`CREATE SINK "database"."schema"."sink" IN CLUSTER "cluster" FROM "database"."schema"."table";`, bc.Create())
}

func TestSinkKafkaCreateParamsQuery(t *testing.T) {
	r := require.New(t)
	b := NewSinkKafkaBuilder("sink", "schema", "database")
	b.Size("xsmall")
	b.From(IdentifierSchemaStruct{Name: "table", SchemaName: "schema", DatabaseName: "database"})
	b.KafkaConnection(IdentifierSchemaStruct{Name: "kafka_connection", SchemaName: "schema", DatabaseName: "database"})
	b.Topic("test_avro_topic")
	b.Key([]string{"key_1", "key_2"})
	b.Format(SinkFormatSpecStruct{Avro: &SinkAvroFormatSpec{SchemaRegistryConnection: IdentifierSchemaStruct{Name: "csr_connection", DatabaseName: "database", SchemaName: "public"}}})
	b.Envelope("UPSERT")
	b.Snapshot(false)
	r.Equal(`CREATE SINK "database"."schema"."sink" FROM "database"."schema"."table" INTO KAFKA CONNECTION "database"."schema"."kafka_connection" KEY (key_1, key_2) (TOPIC 'test_avro_topic') FORMAT AVRO USING CONFLUENT SCHEMA REGISTRY CONNECTION "database"."public"."csr_connection" ENVELOPE UPSERT WITH ( SIZE = 'xsmall' SNAPSHOT = false);`, b.Create())
}

func TestSinkKafkaReadIdQuery(t *testing.T) {
	r := require.New(t)
	b := NewSinkKafkaBuilder("sink", "schema", "database")
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
	b := NewSinkKafkaBuilder("sink", "schema", "database")
	r.Equal(`ALTER SINK "database"."schema"."sink" RENAME TO "database"."schema"."new_sink";`, b.Rename("new_sink"))
}

func TestSinkKafkaResizeQuery(t *testing.T) {
	r := require.New(t)
	b := NewSinkKafkaBuilder("sink", "schema", "database")
	r.Equal(`ALTER SINK "database"."schema"."sink" SET (SIZE = 'xlarge');`, b.UpdateSize("xlarge"))
}

func TestSinkKafkaDropQuery(t *testing.T) {
	r := require.New(t)
	b := NewSinkKafkaBuilder("sink", "schema", "database")
	r.Equal(`DROP SINK "database"."schema"."sink";`, b.Drop())
}

func TestSinkKafkaReadParamsQuery(t *testing.T) {
	r := require.New(t)
	b := ReadSinkParams("u1")
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
