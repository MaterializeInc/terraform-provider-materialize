package materialize

import (
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/MaterializeInc/terraform-provider-materialize/pkg/testhelpers"
	"github.com/jmoiron/sqlx"
)

func TestTableCreateQuery(t *testing.T) {
	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		mock.ExpectExec(
			`CREATE TABLE "database"."schema"."table" (column_1 int, column_2 text NOT NULL);`,
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

		b.Create()
	})
}

func TestTableRenameQuery(t *testing.T) {
	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		mock.ExpectExec(
			`ALTER TABLE "database"."schema"."table" RENAME TO "database"."schema"."new_table";`,
		).WillReturnResult(sqlmock.NewResult(1, 1))

		b := NewTableBuilder(db, "table", "schema", "database")

		b.Create()
	})
}

func TestTableDropQuery(t *testing.T) {
	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		mock.ExpectExec(
			`DROP TABLE "database"."schema"."table";`,
		).WillReturnResult(sqlmock.NewResult(1, 1))

		b := NewTableBuilder(db, "table", "schema", "database")

		b.Create()
	})
}
