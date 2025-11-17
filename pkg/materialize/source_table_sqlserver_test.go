package materialize

import (
	"testing"

	"github.com/MaterializeInc/terraform-provider-materialize/pkg/testhelpers"
	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
)

func TestSourceTableSQLServerBuilder(t *testing.T) {
	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		b := NewSourceTableSQLServerBuilder(db, MaterializeObject{Name: "table", SchemaName: "schema", DatabaseName: "database"})
		b.Source(IdentifierSchemaStruct{Name: "source", SchemaName: "public", DatabaseName: "materialize"})
		b.UpstreamName("upstream_table")
		b.UpstreamSchemaName("upstream_schema")
		b.TextColumns([]ColumnReferenceStruct{
			{ColumnName: "column1"},
			{ColumnName: "column2"},
		})
		b.ExcludeColumns([]ColumnReferenceStruct{
			{ColumnName: "column3"},
			{ColumnName: "column4"},
		})

		mock.ExpectExec(
			`CREATE TABLE "database"."schema"."table" FROM SOURCE "materialize"."public"."source" \(REFERENCE "upstream_schema"."upstream_table"\) WITH \(TEXT COLUMNS \("column1", "column2"\), EXCLUDE COLUMNS \("column3", "column4"\)\);`,
		).WillReturnResult(sqlmock.NewResult(1, 1))

		if err := b.Create(); err != nil {
			t.Fatal(err)
		}
	})
}
