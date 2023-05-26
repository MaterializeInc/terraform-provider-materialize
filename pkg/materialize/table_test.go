package materialize

import (
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/MaterializeInc/terraform-provider-materialize/pkg/testhelpers"
	"github.com/jmoiron/sqlx"
)

func TestTableCreate(t *testing.T) {
	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		mock.ExpectExec(
			`CREATE TABLE "database"."schema"."table" \(column_1 int, column_2 text NOT NULL\);`,
		).WillReturnResult(sqlmock.NewResult(1, 1))

		b := NewTableBuilder(db, "table", "schema", "database")
		b.Column([]TableColumn{
			{
				ColName: "column_1",
				ColType: "int",
			},
			{
				ColName: "column_2",
				ColType: "text",
				NotNull: true,
			},
		})

		if err := b.Create(); err != nil {
			t.Fatal(err)
		}
	})
}

func TestTableRename(t *testing.T) {
	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		mock.ExpectExec(
			`ALTER TABLE "database"."schema"."table" RENAME TO "database"."schema"."new_table";`,
		).WillReturnResult(sqlmock.NewResult(1, 1))

		if err := NewTableBuilder(db, "table", "schema", "database").Rename("new_table"); err != nil {
			t.Fatal(err)
		}
	})
}

func TestTableDrop(t *testing.T) {
	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		mock.ExpectExec(
			`DROP TABLE "database"."schema"."table";`,
		).WillReturnResult(sqlmock.NewResult(1, 1))

		if err := NewTableBuilder(db, "table", "schema", "database").Drop(); err != nil {
			t.Fatal(err)
		}
	})
}
