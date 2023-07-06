package materialize

import (
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/MaterializeInc/terraform-provider-materialize/pkg/testhelpers"
	"github.com/jmoiron/sqlx"
)

func TestSystemPrivilegeGrant(t *testing.T) {
	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		mock.ExpectExec(`GRANT CREATEDB ON SYSTEM TO joe;`).WillReturnResult(sqlmock.NewResult(1, 1))

		b := NewSystemPrivilegeBuilder(db, "joe", "CREATEDB")
		if err := b.Grant(); err != nil {
			t.Fatal(err)
		}
	})
}

func TestSystemPrivilegeRevoke(t *testing.T) {
	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		mock.ExpectExec(`REVOKE CREATEDB ON SYSTEM FROM joe;`).WillReturnResult(sqlmock.NewResult(1, 1))

		b := NewSystemPrivilegeBuilder(db, "joe", "CREATEDB")
		if err := b.Revoke(); err != nil {
			t.Fatal(err)
		}
	})
}

func TestSystemPrivilegeId(t *testing.T) {
	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		// Query Id
		ip := `WHERE mz_roles.name = 'joe'`
		testhelpers.MockRoleScan(mock, ip)

		i, err := SystemPrivilegeId(db, "joe", "CREATECLUSTER")
		if err != nil {
			t.Fatal(err)
		}

		if i != "GRANT SYSTEM|u1|CREATECLUSTER" {
			t.Fatalf("unexpected id %s", i)
		}
	})
}

func TestScanSystemPrivileges(t *testing.T) {
	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		testhelpers.MockSystemPrivilege(mock)

		p, err := ScanSystemPrivileges(db)
		if err != nil {
			t.Fatal(err)
		}

		if p[0].Privileges.String != "s1=RBN/s1" {
			t.Fatalf("unexpected privileges")
		}
	})
}
