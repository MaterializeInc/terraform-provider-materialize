package materialize

import (
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/MaterializeInc/terraform-provider-materialize/pkg/testhelpers"
	"github.com/jmoiron/sqlx"
)

var sourceTable = MaterializeObject{Name: "table", SchemaName: "schema", DatabaseName: "database"}

func TestSourceTableCreate(t *testing.T) {
	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		mock.ExpectExec(
			`CREATE TABLE "database"."schema"."table"
			FROM SOURCE "materialize"."public"."source"
			\(REFERENCE "upstream_schema"."upstream_table"\)
			WITH \(TEXT COLUMNS \(column1, column2\)\);`,
		).WillReturnResult(sqlmock.NewResult(1, 1))

		b := NewSourceTableBuilder(db, sourceTable)
		b.Source(IdentifierSchemaStruct{Name: "source", SchemaName: "public", DatabaseName: "materialize"})
		b.UpstreamName("upstream_table")
		b.UpstreamSchemaName("upstream_schema")
		b.TextColumns([]string{"column1", "column2"})

		if err := b.Create(); err != nil {
			t.Fatal(err)
		}
	})
}

func TestSourceTableRename(t *testing.T) {
	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		mock.ExpectExec(
			`ALTER TABLE "database"."schema"."table" RENAME TO "database"."schema"."new_table";`,
		).WillReturnResult(sqlmock.NewResult(1, 1))

		b := NewSourceTableBuilder(db, sourceTable)
		if err := b.Rename("new_table"); err != nil {
			t.Fatal(err)
		}
	})
}

func TestSourceTableDrop(t *testing.T) {
	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		mock.ExpectExec(
			`DROP TABLE "database"."schema"."table";`,
		).WillReturnResult(sqlmock.NewResult(1, 1))

		b := NewSourceTableBuilder(db, sourceTable)
		if err := b.Drop(); err != nil {
			t.Fatal(err)
		}
	})
}
