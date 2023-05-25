package materialize

import (
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/MaterializeInc/terraform-provider-materialize/pkg/testhelpers"
	"github.com/jmoiron/sqlx"
)

func TestViewCreateQuery(t *testing.T) {
	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		mock.ExpectExec(
			`CREATE VIEW "database"."schema"."view" AS SELECT 1 FROM t1;`,
		).WillReturnResult(sqlmock.NewResult(1, 1))

		b := NewViewBuilder(db, "view", "schema", "database")
		b.SelectStmt("SELECT 1 FROM t1")

		b.Create()
	})
}

func TestViewRenameQuery(t *testing.T) {
	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		mock.ExpectExec(
			`ALTER VIEW "database"."schema"."view" RENAME TO "database"."schema"."new_view";`,
		).WillReturnResult(sqlmock.NewResult(1, 1))

		b := NewViewBuilder(db, "view", "schema", "database")

		b.Rename("new_view")
	})
}

func TestViewDropQuery(t *testing.T) {
	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		mock.ExpectExec(
			`DROP VIEW "database"."schema"."view";`,
		).WillReturnResult(sqlmock.NewResult(1, 1))

		b := NewViewBuilder(db, "view", "schema", "database")

		b.Drop()
	})
}
