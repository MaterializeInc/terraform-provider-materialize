package materialize

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestResourceSourceKafkaCreateQuery(t *testing.T) {
	r := require.New(t)

	bs := NewSourceKafkaBuilder("source", "schema", "database")
	bs.Size("xsmall")
	bs.KafkaConnection(IdentifierSchemaStruct{Name: "kafka_connection", DatabaseName: "database", SchemaName: "schema"})
	bs.Topic("events")
	bs.Format(FormatSpecStruct{Text: true})
	r.Equal(`CREATE SOURCE "database"."schema"."source" FROM KAFKA CONNECTION "database"."schema"."kafka_connection" (TOPIC 'events') FORMAT TEXT WITH (SIZE = 'xsmall');`, bs.Create())

	bc := NewSourceKafkaBuilder("source", "schema", "database")
	bc.ClusterName("cluster")
	bc.KafkaConnection(IdentifierSchemaStruct{Name: "kafka_connection", DatabaseName: "database", SchemaName: "schema"})
	bc.Topic("events")
	bc.Format(FormatSpecStruct{Text: true})
	r.Equal(`CREATE SOURCE "database"."schema"."source" IN CLUSTER "cluster" FROM KAFKA CONNECTION "database"."schema"."kafka_connection" (TOPIC 'events') FORMAT TEXT;`, bc.Create())
}

func TestResourceSourceKafkaCreateParamsQuery(t *testing.T) {
	r := require.New(t)
	b := NewSourceKafkaBuilder("source", "schema", "database")
	b.Size("xsmall")
	b.KafkaConnection(IdentifierSchemaStruct{Name: "kafka_connection", DatabaseName: "database", SchemaName: "schema"})
	b.Topic("events")
	b.Format(FormatSpecStruct{Avro: &AvroFormatSpec{SchemaRegistryConnection: IdentifierSchemaStruct{Name: "csr_connection", DatabaseName: "database", SchemaName: "schema"}}})
	b.IncludeKey("KEY")
	b.IncludePartition("PARTITION")
	b.IncludeOffset("OFFSET")
	b.IncludeTimestamp("TIMESTAMP")
	b.Envelope(KafkaSourceEnvelopeStruct{Upsert: true})
	r.Equal(`CREATE SOURCE "database"."schema"."source" FROM KAFKA CONNECTION "database"."schema"."kafka_connection" (TOPIC 'events') FORMAT AVRO USING CONFLUENT SCHEMA REGISTRY CONNECTION "database"."schema"."csr_connection" INCLUDE KEY, PARTITION, OFFSET, TIMESTAMP ENVELOPE UPSERT WITH (SIZE = 'xsmall');`, b.Create())
}

func TestResourceSourceKafkaReadIdQuery(t *testing.T) {
	r := require.New(t)
	b := NewSourceKafkaBuilder("source", "schema", "database")
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
	b := NewSourceKafkaBuilder("source", "schema", "database")
	r.Equal(`ALTER SOURCE "database"."schema"."source" RENAME TO "database"."schema"."new_source";`, b.Rename("new_source"))
}

func TestResourceSourceKafkaResizeQuery(t *testing.T) {
	r := require.New(t)
	b := NewSourceKafkaBuilder("source", "schema", "database")
	r.Equal(`ALTER SOURCE "database"."schema"."source" SET (SIZE = 'xlarge');`, b.UpdateSize("xlarge"))
}

func TestResourceSourceKafkaDropQuery(t *testing.T) {
	r := require.New(t)
	b := NewSourceKafkaBuilder("source", "schema", "database")
	r.Equal(`DROP SOURCE "database"."schema"."source";`, b.Drop())
}

func TestResourceSourceKafkaReadParamsQuery(t *testing.T) {
	r := require.New(t)
	b := ReadSourceParams("u1")
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
