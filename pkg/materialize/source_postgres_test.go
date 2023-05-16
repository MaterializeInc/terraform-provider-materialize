package materialize

import (
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/MaterializeInc/terraform-provider-materialize/pkg/testhelpers"
	"github.com/jmoiron/sqlx"
)

func TestSourcePostgresCreate(t *testing.T) {
	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		mock.ExpectExec(
			`CREATE SOURCE "database"."schema"."source" FROM POSTGRES CONNECTION "database"."schema"."pg_connection" \(PUBLICATION 'mz_source', TEXT COLUMNS \(table.unsupported_type_1, table.unsupported_type_2\)\) FOR TABLES \(schema1.table_1 AS s1_table_1, schema2.table_1 AS s2_table_1\) WITH \(SIZE = 'xsmall'\);`,
		).WillReturnResult(sqlmock.NewResult(1, 1))

		b := NewSourcePostgresBuilder(db, "source", "schema", "database")
		b.Size("xsmall")
		b.PostgresConnection(IdentifierSchemaStruct{Name: "pg_connection", SchemaName: "schema", DatabaseName: "database"})
		b.Publication("mz_source")
		b.TextColumns([]string{"table.unsupported_type_1", "table.unsupported_type_2"})
		b.Table([]Table{
			{
				Name:  "schema1.table_1",
				Alias: "s1_table_1",
			},
			{
				Name:  "schema2.table_1",
				Alias: "s2_table_1",
			},
		})

		if err := b.Create(); err != nil {
			t.Fatal(err)
		}
	})
}
