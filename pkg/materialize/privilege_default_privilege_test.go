package materialize

import (
	"database/sql"
	"reflect"
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/MaterializeInc/terraform-provider-materialize/pkg/testhelpers"
	"github.com/jmoiron/sqlx"
)

func TestParseDefaultPrivileges(t *testing.T) {
	input := []DefaultPrivilegeParams{
		{
			ObjectType: sql.NullString{String: "TYPE", Valid: true},
			GranteeId:  sql.NullString{String: "p", Valid: true},
			Privileges: sql.NullString{String: "U", Valid: true},
		},
		{
			ObjectType: sql.NullString{String: "CLUSTER", Valid: true},
			GranteeId:  sql.NullString{String: "s2", Valid: true},
			Privileges: sql.NullString{String: "U", Valid: true},
		},
		{
			ObjectType: sql.NullString{String: "TABLE", Valid: true},
			GranteeId:  sql.NullString{String: "u9", Valid: true},
			Privileges: sql.NullString{String: "ar", Valid: true},
		},
		{
			ObjectType: sql.NullString{String: "TABLE", Valid: true},
			GranteeId:  sql.NullString{String: "u9", Valid: true},
			Privileges: sql.NullString{String: "w", Valid: true},
			DatabaseId: sql.NullString{String: "u3", Valid: true},
			SchemaId:   sql.NullString{String: "u9", Valid: true},
		},
	}

	output, err := ParseDefaultPrivileges(input)
	if err != nil {
		t.Fatal(err)
	}

	expected := map[DefaultPrivilegeMapKey][]string{
		{GranteeId: "p", ObjectType: "TYPE"}:                                     {"USAGE"},
		{GranteeId: "s2", ObjectType: "CLUSTER"}:                                 {"USAGE"},
		{GranteeId: "u9", ObjectType: "TABLE"}:                                   {"INSERT", "SELECT"},
		{GranteeId: "u9", ObjectType: "TABLE", DatabaseId: "u3", SchemaId: "u9"}: {"UPDATE"},
	}

	if !reflect.DeepEqual(output, expected) {
		t.Fatal("ouptut does not equal expected")
	}
}

func TestDefaultPrivilegeGrantSimple(t *testing.T) {
	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		mock.ExpectExec(`ALTER DEFAULT PRIVILEGES GRANT SELECT ON TABLES TO joe;`).WillReturnResult(sqlmock.NewResult(1, 1))

		b := NewDefaultPrivilegeBuilder(db, "TABLE", "joe", "SELECT")
		if err := b.Grant(); err != nil {
			t.Fatal(err)
		}
	})
}

func TestDefaultPrivilegeGrantComplex(t *testing.T) {
	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		mock.ExpectExec(`ALTER DEFAULT PRIVILEGES FOR ROLE interns IN DATABASE "dev" GRANT ALL PRIVILEGES ON TABLES TO intern_managers;`).WillReturnResult(sqlmock.NewResult(1, 1))

		b := NewDefaultPrivilegeBuilder(db, "TABLE", "intern_managers", "ALL PRIVILEGES")
		b.TargetRole("interns")
		b.DatabaseName("dev")
		if err := b.Grant(); err != nil {
			t.Fatal(err)
		}
	})
}

func TestDefaultPrivilegeRevokeSimple(t *testing.T) {
	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		mock.ExpectExec(`ALTER DEFAULT PRIVILEGES FOR ROLE developers REVOKE USAGE ON SECRETS FROM project_managers;`).WillReturnResult(sqlmock.NewResult(1, 1))

		b := NewDefaultPrivilegeBuilder(db, "SECRET", "project_managers", "USAGE")
		b.TargetRole("developers")
		if err := b.Revoke(); err != nil {
			t.Fatal(err)
		}
	})
}

func TestDefaultPrivilegeGrantAllRoles(t *testing.T) {
	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		mock.ExpectExec(`ALTER DEFAULT PRIVILEGES FOR ALL ROLES GRANT SELECT ON TABLES TO managers;`).WillReturnResult(sqlmock.NewResult(1, 1))

		b := NewDefaultPrivilegeBuilder(db, "TABLE", "managers", "SELECT")
		b.TargetRole("ALL")
		if err := b.Grant(); err != nil {
			t.Fatal(err)
		}
	})
}
