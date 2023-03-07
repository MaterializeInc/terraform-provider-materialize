package resources

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestResourceSourceKafkaCreate(t *testing.T) {
	r := require.New(t)

	bs := newSourceKafkaBuilder("source", "schema", "database")
	bs.Size("xsmall")
	bs.KafkaConnection("kafka_connection")
	bs.Topic("events")
	bs.Format("TEXT")
	r.Equal(`CREATE SOURCE database.schema.source FROM KAFKA CONNECTION kafka_connection (TOPIC 'events') FORMAT TEXT WITH (SIZE = 'xsmall');`, bs.Create())

	bc := newSourceKafkaBuilder("source", "schema", "database")
	bc.ClusterName("cluster")
	bc.KafkaConnection("kafka_connection")
	bc.Topic("events")
	bc.Format("TEXT")
	r.Equal(`CREATE SOURCE database.schema.source IN CLUSTER cluster FROM KAFKA CONNECTION kafka_connection (TOPIC 'events') FORMAT TEXT;`, bc.Create())
}

func TestResourceSourceKafkaCreateParams(t *testing.T) {
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
	r.Equal(`CREATE SOURCE database.schema.source FROM KAFKA CONNECTION kafka_connection (TOPIC 'events') FORMAT AVRO USING CONFLUENT SCHEMA REGISTRY CONNECTION csr_connection INCLUDE KEY, PARTITION, OFFSET, TIMESTAMP ENVELOPE UPSERT WITH (SIZE = 'xsmall');`, b.Create())
}

func TestResourceSourceKafkaReadId(t *testing.T) {
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

func TestResourceSourceKafkaRename(t *testing.T) {
	r := require.New(t)
	b := newSourceKafkaBuilder("source", "schema", "database")
	r.Equal(`ALTER SOURCE database.schema.source RENAME TO database.schema.new_source;`, b.Rename("new_source"))
}

func TestResourceSourceKafkaResize(t *testing.T) {
	r := require.New(t)
	b := newSourceKafkaBuilder("source", "schema", "database")
	r.Equal(`ALTER SOURCE database.schema.source SET (SIZE = 'xlarge');`, b.UpdateSize("xlarge"))
}

func TestResourceSourceKafkaDrop(t *testing.T) {
	r := require.New(t)
	b := newSourceKafkaBuilder("source", "schema", "database")
	r.Equal(`DROP SOURCE database.schema.source;`, b.Drop())
}

func TestResourceSourceKafkaReadParams(t *testing.T) {
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
