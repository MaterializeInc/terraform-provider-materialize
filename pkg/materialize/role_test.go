package materialize

import (
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/MaterializeInc/terraform-provider-materialize/pkg/testhelpers"
	"github.com/jmoiron/sqlx"
)

func TestRoleCreateQuery(t *testing.T) {
	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		mock.ExpectExec(
			`CREATE ROLE "role" INHERIT CREATEROLE CREATEDB CREATECLUSTER;`,
		).WillReturnResult(sqlmock.NewResult(1, 1))
		b := NewRoleBuilder(db, "role")
		b.Inherit()
		b.CreateRole()
		b.CreateDb()
		b.CreateCluster()

		b.Create()
	})
}
