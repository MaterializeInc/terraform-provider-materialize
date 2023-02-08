package resources

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestResourceSourceCreate(t *testing.T) {
	r := require.New(t)

	bs := newSourceBuilder("source", "schema", "database")
	bs.Size("xsmall")
	r.Equal(`CREATE SOURCE database.schema.source WITH (SIZE = 'xsmall');`, bs.Create())

	bc := newSourceBuilder("source", "schema", "database")
	bc.ClusterName("cluster")
	r.Equal(`CREATE SOURCE database.schema.source IN CLUSTER cluster;`, bc.Create())
}

func TestResourceSourceCreateLoadGenerator(t *testing.T) {
	r := require.New(t)
	b := newSourceBuilder("source", "schema", "database")
	b.Size("xsmall")
	b.ConnectionType("LOAD GENERATOR")
	b.LoadGeneratorType("TPCH")
	b.TickInterval("1s")
	b.ScaleFactor(0.01)
	r.Equal(`CREATE SOURCE database.schema.source FROM LOAD GENERATOR TPCH (TICK INTERVAL '1s', SCALE FACTOR 0.01) WITH (SIZE = 'xsmall');`, b.Create())
}

func TestResourceSourceCreatePostgres(t *testing.T) {
	r := require.New(t)
	b := newSourceBuilder("source", "schema", "database")
	b.Size("xsmall")
	b.ConnectionType("POSTGRES")
	b.PostgresConnection("pg_connection")
	b.Publication("mz_source")
	r.Equal(`CREATE SOURCE database.schema.source FROM POSTGRES CONNECTION pg_connection (PUBLICATION 'mz_source') FOR ALL TABLES WITH (SIZE = 'xsmall');`, b.Create())
}

func TestResourceSourceCreatePostgresTables(t *testing.T) {
	r := require.New(t)
	b := newSourceBuilder("source", "schema", "database")
	b.Size("xsmall")
	b.ConnectionType("POSTGRES")
	b.PostgresConnection("pg_connection")
	b.Publication("mz_source")
	b.Tables(map[string]string{
		"schema1.table_1": "s1_table_1",
		"schema2_table_1": "s2_table_1",
	})
	r.Equal(`CREATE SOURCE database.schema.source FROM POSTGRES CONNECTION pg_connection (PUBLICATION 'mz_source') FOR TABLES (schema1.table_1 AS s1_table_1, schema2_table_1 AS s2_table_1) WITH (SIZE = 'xsmall');`, b.Create())
}

func TestResourceSourceCreateKafka(t *testing.T) {
	r := require.New(t)
	b := newSourceBuilder("source", "schema", "database")
	b.Size("xsmall")
	b.ConnectionType("KAFKA")
	b.KafkaConnection("kafka_connection")
	b.Topic("events")
	b.Format("AVRO")
	b.SchemaRegistryConnection("csr_connection")
	b.Envelope("UPSERT")
	r.Equal(`CREATE SOURCE database.schema.source FROM KAFKA CONNECTION kafka_connection (TOPIC 'events') FORMAT AVRO USING CONFLUENT SCHEMA REGISTRY CONNECTION csr_connection ENVELOPE UPSERT WITH (SIZE = 'xsmall');`, b.Create())
}

func TestResourceSourceRead(t *testing.T) {
	r := require.New(t)
	b := newSourceBuilder("source", "schema", "database")
	r.Equal(`
		SELECT
			mz_sources.id,
			mz_sources.name,
			mz_schemas.name,
			mz_databases.name,
			mz_sources.type,
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
		WHERE mz_sources.name = 'source'
		AND mz_schemas.name = 'schema'
		AND mz_databases.name = 'database';
	`, b.Read())
}

func TestResourceSourceRename(t *testing.T) {
	r := require.New(t)
	b := newSourceBuilder("source", "schema", "database")
	r.Equal(`ALTER SOURCE database.schema.source RENAME TO database.schema.new_source;`, b.Rename("new_source"))
}

func TestResourceSourceResize(t *testing.T) {
	r := require.New(t)
	b := newSourceBuilder("source", "schema", "database")
	r.Equal(`ALTER SOURCE database.schema.source SET (SIZE = 'xlarge');`, b.UpdateSize("xlarge"))
}

func TestResourceSourceDrop(t *testing.T) {
	r := require.New(t)
	b := newSourceBuilder("source", "schema", "database")
	r.Equal(`DROP SOURCE database.schema.source;`, b.Drop())
}
