package materialize

import (
	"database/sql"
	"reflect"
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/MaterializeInc/terraform-provider-materialize/pkg/testhelpers"
	"github.com/jmoiron/sqlx"
)

func TestParseRolePrivileges(t *testing.T) {
	input := []RolePrivilegeParams{
		{
			RoleId:  sql.NullString{String: "u1", Valid: true},
			Member:  sql.NullString{String: "u2", Valid: true},
			Grantor: sql.NullString{String: "s1", Valid: true},
		},
		{
			RoleId:  sql.NullString{String: "u1", Valid: true},
			Member:  sql.NullString{String: "u3", Valid: true},
			Grantor: sql.NullString{String: "s1", Valid: true},
		},
		{
			RoleId:  sql.NullString{String: "u2", Valid: true},
			Member:  sql.NullString{String: "u5", Valid: true},
			Grantor: sql.NullString{String: "s1", Valid: true},
		},
	}

	output, err := ParseRolePrivileges(input)
	if err != nil {
		t.Fatal(err)
	}

	expected := map[string][]string{"u1": {"u2", "u3"}, "u2": {"u5"}}

	if !reflect.DeepEqual(output, expected) {
		t.Fatal("ouptut does not equal expected")
	}
}

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
