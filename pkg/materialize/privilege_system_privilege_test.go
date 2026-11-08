package materialize

import (
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
)

func TestSystemPrivilegeGrant(t *testing.T) {
	withMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		mock.ExpectExec(`GRANT CREATEDB ON SYSTEM TO "joe";`).WillReturnResult(sqlmock.NewResult(1, 1))

		b := NewSystemPrivilegeBuilder(db, "joe", "CREATEDB")
		if err := b.Grant(); err != nil {
			t.Fatal(err)
		}
	})
}

func TestSystemPrivilegeRevoke(t *testing.T) {
	withMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		mock.ExpectExec(`REVOKE CREATEDB ON SYSTEM FROM "joe";`).WillReturnResult(sqlmock.NewResult(1, 1))

		b := NewSystemPrivilegeBuilder(db, "joe", "CREATEDB")
		if err := b.Revoke(); err != nil {
			t.Fatal(err)
		}
	})
}
