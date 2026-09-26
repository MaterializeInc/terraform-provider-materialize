package materialize

import (
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/MaterializeInc/terraform-provider-materialize/pkg/testhelpers"
	"github.com/jmoiron/sqlx"
)

// https://github.com/MaterializeInc/materialize/blob/main/test/testdrive/tables.td
// https://materialize.com/docs/sql/create-table/

func TestTableCreate(t *testing.T) {
	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		mock.ExpectExec(
			`CREATE TABLE "database"."schema"."table"
			\("a" int, "b" text, "c" text NOT NULL, "d" int DEFAULT \(1\), "e" text NOT NULL DEFAULT NULL\);`,
		).WillReturnResult(sqlmock.NewResult(1, 1))

		o := MaterializeObject{Name: "table", SchemaName: "schema", DatabaseName: "database"}
		b := NewTableBuilder(db, o)
		b.Column([]TableColumn{
			{
				ColName: "a",
				ColType: "int",
			},
			{
				ColName: "b",
				ColType: "text",
			},
			{
				ColName: "c",
				ColType: "text",
				NotNull: true,
			},
			{
				ColName: "d",
				ColType: "int",
				Default: "(1)",
			},
			{
				ColName: "e",
				ColType: "text",
				NotNull: true,
				Default: "NULL",
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
			`ALTER TABLE "database"."schema"."table" RENAME TO "new_table";`,
		).WillReturnResult(sqlmock.NewResult(1, 1))

		o := MaterializeObject{Name: "table", SchemaName: "schema", DatabaseName: "database"}
		if err := NewTableBuilder(db, o).Rename("new_table"); err != nil {
			t.Fatal(err)
		}
	})
}

func TestTableDrop(t *testing.T) {
	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		mock.ExpectExec(
			`DROP TABLE "database"."schema"."table";`,
		).WillReturnResult(sqlmock.NewResult(1, 1))

		o := MaterializeObject{Name: "table", SchemaName: "schema", DatabaseName: "database"}
		if err := NewTableBuilder(db, o).Drop(); err != nil {
			t.Fatal(err)
		}
	})
}

// Column names are quoted so Materialize keeps them exactly as written; unquoted
// names folded to lower case and could not contain spaces or other characters.
func TestTableCreateQuotesColumnNames(t *testing.T) {
	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		mock.ExpectExec(
			`CREATE TABLE "database"."schema"."table"
			\("TenantId" int, "first name" text, "order" text, "we""ird" text, "Quoted" text\);`,
		).WillReturnResult(sqlmock.NewResult(1, 1))

		o := MaterializeObject{Name: "table", SchemaName: "schema", DatabaseName: "database"}
		b := NewTableBuilder(db, o)
		b.Column([]TableColumn{
			{ColName: "TenantId", ColType: "int"},
			{ColName: "first name", ColType: "text"},
			{ColName: "order", ColType: "text"},
			{ColName: `we"ird`, ColType: "text"},
			// already quoted in the value: passed through as before
			{ColName: `"Quoted"`, ColType: "text"},
		})

		if err := b.Create(); err != nil {
			t.Fatal(err)
		}
	})
}
