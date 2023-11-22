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
			`CREATE TABLE "database"."schema"."table" \(a int, b text\);`,
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
		})

		if err := b.Create(); err != nil {
			t.Fatal(err)
		}
	})
}

func TestTableNotNullCreate(t *testing.T) {
	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		mock.ExpectExec(
			`CREATE TABLE "database"."schema"."table" \(a int, b text NOT NULL\);`,
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
				NotNull: true,
			},
		})

		if err := b.Create(); err != nil {
			t.Fatal(err)
		}
	})
}

func TestTableDefaultCreate(t *testing.T) {
	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		mock.ExpectExec(
			`CREATE TABLE "database"."schema"."table" \(a int DEFAULT 1\);`,
		).WillReturnResult(sqlmock.NewResult(1, 1))

		o := MaterializeObject{Name: "table", SchemaName: "schema", DatabaseName: "database"}
		b := NewTableBuilder(db, o)
		b.Column([]TableColumn{
			{
				ColName: "a",
				ColType: "int",
				Default: "1",
			},
		})

		if err := b.Create(); err != nil {
			t.Fatal(err)
		}
	})
}

func TestTableNotNullDefaultCreate(t *testing.T) {
	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		mock.ExpectExec(
			`CREATE TABLE "database"."schema"."table" \(a int NOT NULL DEFAULT NULL\);`,
		).WillReturnResult(sqlmock.NewResult(1, 1))

		o := MaterializeObject{Name: "table", SchemaName: "schema", DatabaseName: "database"}
		b := NewTableBuilder(db, o)
		b.Column([]TableColumn{
			{
				ColName: "a",
				ColType: "int",
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
