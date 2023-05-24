package materialize

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSourceLoadgenCreateQuery(t *testing.T) {
	r := require.New(t)

	bs := NewSourceLoadgenBuilder("source", "schema", "database")
	bs.Size("xsmall")
	bs.LoadGeneratorType("COUNTER")
	r.Equal(`CREATE SOURCE "database"."schema"."source" FROM LOAD GENERATOR COUNTER WITH (SIZE = 'xsmall');`, bs.Create())

	bc := NewSourceLoadgenBuilder("source", "schema", "database")
	bc.ClusterName("cluster")
	bc.LoadGeneratorType("COUNTER")
	r.Equal(`CREATE SOURCE "database"."schema"."source" IN CLUSTER "cluster" FROM LOAD GENERATOR COUNTER;`, bc.Create())
}

func TestSourceLoadgenCreateCounterParamsQuery(t *testing.T) {
	r := require.New(t)
	b := NewSourceLoadgenBuilder("source", "schema", "database")
	b.Size("xsmall")
	b.LoadGeneratorType("COUNTER")
	b.CounterOptions(CounterOptions{
		TickInterval:   "1s",
		MaxCardinality: 8,
	})
	r.Equal(`CREATE SOURCE "database"."schema"."source" FROM LOAD GENERATOR COUNTER (TICK INTERVAL '1s', MAX CARDINALITY 8) WITH (SIZE = 'xsmall');`, b.Create())
}

func TestSourceLoadgenCreateAuctionParamsQuery(t *testing.T) {
	r := require.New(t)
	b := NewSourceLoadgenBuilder("source", "schema", "database")
	b.Size("xsmall")
	b.LoadGeneratorType("AUCTION")
	b.AuctionOptions(AuctionOptions{
		TickInterval: "1s",
		ScaleFactor:  0.01,
	})
	r.Equal(`CREATE SOURCE "database"."schema"."source" FROM LOAD GENERATOR AUCTION (TICK INTERVAL '1s', SCALE FACTOR 0.01) FOR ALL TABLES WITH (SIZE = 'xsmall');`, b.Create())
}

func TestSourceLoadgenCreateTPCHParamsQuery(t *testing.T) {
	r := require.New(t)
	b := NewSourceLoadgenBuilder("source", "schema", "database")
	b.Size("xsmall")
	b.LoadGeneratorType("TPCH")
	b.TPCHOptions(TPCHOptions{
		TickInterval: "1s",
		ScaleFactor:  0.01,
		Table: []TableStruct{
			{
				Name:  "schema1.table_1",
				Alias: "s1_table_1",
			},
			{
				Name:  "schema2.table_1",
				Alias: "s2_table_1",
			},
		},
	})
	r.Equal(`CREATE SOURCE "database"."schema"."source" FROM LOAD GENERATOR TPCH (TICK INTERVAL '1s', SCALE FACTOR 0.01) FOR TABLES (schema1.table_1 AS s1_table_1, schema2.table_1 AS s2_table_1) WITH (SIZE = 'xsmall');`, b.Create())
}

func TestSourceLoadgenReadIdQuery(t *testing.T) {
	r := require.New(t)
	b := NewSourceLoadgenBuilder("source", "schema", "database")
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

func TestSourceLoadgenRenameQuery(t *testing.T) {
	r := require.New(t)
	b := NewSourceLoadgenBuilder("source", "schema", "database")
	r.Equal(`ALTER SOURCE "database"."schema"."source" RENAME TO "database"."schema"."new_source";`, b.Rename("new_source"))
}

func TestSourceLoadgenResizeQuery(t *testing.T) {
	r := require.New(t)
	b := NewSourceLoadgenBuilder("source", "schema", "database")
	r.Equal(`ALTER SOURCE "database"."schema"."source" SET (SIZE = 'xlarge');`, b.UpdateSize("xlarge"))
}

func TestSourceLoadgenDropQuery(t *testing.T) {
	r := require.New(t)
	b := NewSourceLoadgenBuilder("source", "schema", "database")
	r.Equal(`DROP SOURCE "database"."schema"."source";`, b.Drop())
}

func TestSourceLoadgenReadParamsQuery(t *testing.T) {
	r := require.New(t)
	b := ReadSourceParams("u1")
	r.Equal(`
		SELECT
			mz_sources.name AS source_name,
			mz_schemas.name AS schema_name,
			mz_databases.name AS database_name,
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
