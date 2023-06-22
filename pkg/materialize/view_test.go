package materialize

import (
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/MaterializeInc/terraform-provider-materialize/pkg/testhelpers"
	"github.com/jmoiron/sqlx"
)

func TestViewCreate(t *testing.T) {
	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		mock.ExpectExec(
			`CREATE VIEW "database"."schema"."view" AS SELECT 1 FROM t1;`,
		).WillReturnResult(sqlmock.NewResult(1, 1))

		o := ObjectSchemaStruct{Name: "view", SchemaName: "schema", DatabaseName: "database"}
		b := NewViewBuilder(db, o)
		b.SelectStmt("SELECT 1 FROM t1")

		if err := b.Create(); err != nil {
			t.Fatal(err)
		}
	})
}

func TestViewRename(t *testing.T) {
	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		mock.ExpectExec(
			`ALTER VIEW "database"."schema"."view" RENAME TO "new_view";`,
		).WillReturnResult(sqlmock.NewResult(1, 1))

		o := ObjectSchemaStruct{Name: "view", SchemaName: "schema", DatabaseName: "database"}
		if err := NewViewBuilder(db, o).Rename("new_view"); err != nil {
			t.Fatal(err)
		}
	})
}

func TestViewDrop(t *testing.T) {
	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		mock.ExpectExec(
			`DROP VIEW "database"."schema"."view";`,
		).WillReturnResult(sqlmock.NewResult(1, 1))

		o := ObjectSchemaStruct{Name: "view", SchemaName: "schema", DatabaseName: "database"}
		if err := NewViewBuilder(db, o).Drop(); err != nil {
			t.Fatal(err)
		}
	})
}
