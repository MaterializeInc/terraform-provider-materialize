package resources

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestResourceSourceLoadgenCreate(t *testing.T) {
	r := require.New(t)

	bs := newSourceLoadgenBuilder("source", "schema", "database")
	bs.Size("xsmall")
	bs.LoadGeneratorType("COUNTER")
	r.Equal(`CREATE SOURCE database.schema.source FROM LOAD GENERATOR COUNTER WITH (SIZE = 'xsmall');`, bs.Create())

	bc := newSourceLoadgenBuilder("source", "schema", "database")
	bc.ClusterName("cluster")
	bc.LoadGeneratorType("COUNTER")
	r.Equal(`CREATE SOURCE database.schema.source IN CLUSTER cluster FROM LOAD GENERATOR COUNTER;`, bc.Create())
}

func TestResourceSourceLoadgenCreateParams(t *testing.T) {
	r := require.New(t)
	b := newSourceLoadgenBuilder("source", "schema", "database")
	b.Size("xsmall")
	b.LoadGeneratorType("TPCH")
	b.TickInterval("1s")
	b.ScaleFactor(0.01)
	r.Equal(`CREATE SOURCE database.schema.source FROM LOAD GENERATOR TPCH (TICK INTERVAL '1s', SCALE FACTOR 0.01) FOR ALL TABLES WITH (SIZE = 'xsmall');`, b.Create())
}

func TestResourceSourceLoadgenReadId(t *testing.T) {
	r := require.New(t)
	b := newSourceLoadgenBuilder("source", "schema", "database")
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

func TestResourceSourceLoadgenRename(t *testing.T) {
	r := require.New(t)
	b := newSourceLoadgenBuilder("source", "schema", "database")
	r.Equal(`ALTER SOURCE database.schema.source RENAME TO database.schema.new_source;`, b.Rename("new_source"))
}

func TestResourceSourceLoadgenResize(t *testing.T) {
	r := require.New(t)
	b := newSourceLoadgenBuilder("source", "schema", "database")
	r.Equal(`ALTER SOURCE database.schema.source SET (SIZE = 'xlarge');`, b.UpdateSize("xlarge"))
}

func TestResourceSourceLoadgenDrop(t *testing.T) {
	r := require.New(t)
	b := newSourceLoadgenBuilder("source", "schema", "database")
	r.Equal(`DROP SOURCE database.schema.source;`, b.Drop())
}

func TestResourceSourceLoadgenReadParams(t *testing.T) {
	r := require.New(t)
	b := readSourceParams("u1")
	r.Equal(`
		SELECT
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
		JOIN mz_clusters
			ON mz_sources.cluster_id = mz_clusters.id
		WHERE mz_sources.id = 'u1';`, b)
}
