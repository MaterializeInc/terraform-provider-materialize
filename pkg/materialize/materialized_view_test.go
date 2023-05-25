package materialize

import (
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/MaterializeInc/terraform-provider-materialize/pkg/testhelpers"
	"github.com/jmoiron/sqlx"
)

func TestMaterializedViewCreate(t *testing.T) {
	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		mock.ExpectExec(
			`CREATE MATERIALIZED VIEW "database"."schema"."materialized_view" AS SELECT 1 FROM t1;`,
		).WillReturnResult(sqlmock.NewResult(1, 1))

		b := NewMaterializedViewBuilder(db, "materialized_view", "schema", "database")
		b.SelectStmt("SELECT 1 FROM t1")

		b.Create()
	})
}
