package materialize

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSourcePostgresCreateQuery(t *testing.T) {
	r := require.New(t)

	bs := NewSourcePostgresBuilder("source", "schema", "database")
	bs.Size("xsmall")
	bs.PostgresConnection(IdentifierSchemaStruct{Name: "pg_connection", SchemaName: "schema", DatabaseName: "database"})
	bs.Publication("mz_source")
	r.Equal(`CREATE SOURCE "database"."schema"."source" FROM POSTGRES CONNECTION "database"."schema"."pg_connection" (PUBLICATION 'mz_source') FOR ALL TABLES WITH (SIZE = 'xsmall');`, bs.Create())

	bc := NewSourcePostgresBuilder("source", "schema", "database")
	bc.ClusterName("cluster")
	bc.PostgresConnection(IdentifierSchemaStruct{Name: "pg_connection", SchemaName: "schema", DatabaseName: "database"})
	bc.Publication("mz_source")
	r.Equal(`CREATE SOURCE "database"."schema"."source" IN CLUSTER "cluster" FROM POSTGRES CONNECTION "database"."schema"."pg_connection" (PUBLICATION 'mz_source') FOR ALL TABLES;`, bc.Create())
}

func TestSourcePostgresCreateParamsQuery(t *testing.T) {
	r := require.New(t)
	b := NewSourcePostgresBuilder("source", "schema", "database")
	b.Size("xsmall")
	b.PostgresConnection(IdentifierSchemaStruct{Name: "pg_connection", SchemaName: "schema", DatabaseName: "database"})
	b.Publication("mz_source")
	b.TextColumns([]string{"table.unsupported_type_1", "table.unsupported_type_2"})
	b.Table([]TablePostgres{
		{
			Name:  "schema1.table_1",
			Alias: "s1_table_1",
		},
		{
			Name:  "schema2.table_1",
			Alias: "s2_table_1",
		},
	})
	r.Equal(`CREATE SOURCE "database"."schema"."source" FROM POSTGRES CONNECTION "database"."schema"."pg_connection" (PUBLICATION 'mz_source', TEXT COLUMNS (table.unsupported_type_1, table.unsupported_type_2)) FOR TABLES (schema1.table_1 AS s1_table_1, schema2.table_1 AS s2_table_1) WITH (SIZE = 'xsmall');`, b.Create())
}

func TestSourcePostgresReadIdQuery(t *testing.T) {
	r := require.New(t)
	b := NewSourcePostgresBuilder("source", "schema", "database")
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

func TestSourcePostgresRenameQuery(t *testing.T) {
	r := require.New(t)
	b := NewSourcePostgresBuilder("source", "schema", "database")
	r.Equal(`ALTER SOURCE "database"."schema"."source" RENAME TO "database"."schema"."new_source";`, b.Rename("new_source"))
}

func TestSourcePostgresResizeQuery(t *testing.T) {
	r := require.New(t)
	b := NewSourcePostgresBuilder("source", "schema", "database")
	r.Equal(`ALTER SOURCE "database"."schema"."source" SET (SIZE = 'xlarge');`, b.UpdateSize("xlarge"))
}

func TestSourcePostgresDropQuery(t *testing.T) {
	r := require.New(t)
	b := NewSourcePostgresBuilder("source", "schema", "database")
	r.Equal(`DROP SOURCE "database"."schema"."source";`, b.Drop())
}

func TestSourcePostgresReadParamsQuery(t *testing.T) {
	r := require.New(t)
	b := ReadSourceParams("u1")
	r.Equal(`
		SELECT
			mz_sources.name,
			mz_schemas.name,
			mz_databases.name,
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
