package materialize

import (
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/MaterializeInc/terraform-provider-materialize/pkg/testhelpers"
	"github.com/jmoiron/sqlx"
)

func TestRolePrivilegeGrant(t *testing.T) {
	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		mock.ExpectExec(`GRANT dev_role TO user;`).WillReturnResult(sqlmock.NewResult(1, 1))

		b := NewRolePrivilegeBuilder(db, "dev_role", "user")
		if err := b.Grant(); err != nil {
			t.Fatal(err)
		}
	})
}

func TestRolePrivilegeRevoke(t *testing.T) {
	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		mock.ExpectExec(`REVOKE dev_role FROM user;`).WillReturnResult(sqlmock.NewResult(1, 1))

		b := NewRolePrivilegeBuilder(db, "dev_role", "user")
		if err := b.Revoke(); err != nil {
			t.Fatal(err)
		}
	})
}
