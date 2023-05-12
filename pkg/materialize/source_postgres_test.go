package materialize

// import (
// 	"testing"

// 	"github.com/stretchr/testify/require"
// )

// func TestSourcePostgresCreate(t *testing.T) {
// 	r := require.New(t)
// 	b := NewSourcePostgresBuilder("source", "schema", "database")
// 	b.Size("xsmall")
// 	b.PostgresConnection(IdentifierSchemaStruct{Name: "pg_connection", SchemaName: "schema", DatabaseName: "database"})
// 	b.Publication("mz_source")
// 	b.TextColumns([]string{"table.unsupported_type_1", "table.unsupported_type_2"})
// 	b.Table([]Table{
// 		{
// 			Name:  "schema1.table_1",
// 			Alias: "s1_table_1",
// 		},
// 		{
// 			Name:  "schema2.table_1",
// 			Alias: "s2_table_1",
// 		},
// 	})
// 	r.Equal(`CREATE SOURCE "database"."schema"."source" FROM POSTGRES CONNECTION "database"."schema"."pg_connection" (PUBLICATION 'mz_source', TEXT COLUMNS (table.unsupported_type_1, table.unsupported_type_2)) FOR TABLES (schema1.table_1 AS s1_table_1, schema2.table_1 AS s2_table_1) WITH (SIZE = 'xsmall');`, b.Create())
// }
