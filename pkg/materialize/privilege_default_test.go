package materialize

import (
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/MaterializeInc/terraform-provider-materialize/pkg/testhelpers"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/require"
)

func TestDefaultPrivilegeId(t *testing.T) {
	r := require.New(t)
	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		gp := `WHERE mz_roles.name = 'grantee'`
		testhelpers.MockRoleScan(mock, gp)

		i, err := DefaultPrivilegeId(db, "TABLE", "grantee", "", "", "", "SELECT")
		if err != nil {
			t.Fatal(err)
		}

		r.Equal(i, `GRANT DEFAULT|TABLE|u1||||SELECT`)
	})
}

func TestDefaultPrivilegeIdRole(t *testing.T) {
	r := require.New(t)
	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		gp := `WHERE mz_roles.name = 'grantee'`
		testhelpers.MockRoleScan(mock, gp)

		tp := `WHERE mz_roles.name = 'target'`
		testhelpers.MockRoleScan(mock, tp)

		i, err := DefaultPrivilegeId(db, "TABLE", "grantee", "target", "", "", "SELECT")
		if err != nil {
			t.Fatal(err)
		}

		r.Equal(i, `GRANT DEFAULT|TABLE|u1|u1|||SELECT`)
	})
}

func TestDefaultPrivilegeIdRoleQualified(t *testing.T) {
	r := require.New(t)
	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		gp := `WHERE mz_roles.name = 'grantee'`
		testhelpers.MockRoleScan(mock, gp)

		tp := `WHERE mz_roles.name = 'target'`
		testhelpers.MockRoleScan(mock, tp)

		dp := `WHERE mz_databases.name = 'database'`
		testhelpers.MockDatabaseScan(mock, dp)

		sp := `WHERE mz_databases.name = 'database' AND mz_schemas.name = 'schema'`
		testhelpers.MockSchemaScan(mock, sp)

		i, err := DefaultPrivilegeId(db, "TABLE", "grantee", "target", "database", "schema", "SELECT")
		if err != nil {
			t.Fatal(err)
		}

		r.Equal(i, `GRANT DEFAULT|TABLE|u1|u1|u1|u1|SELECT`)
	})
}

func TestDefaultPrivilegeGrantSimple(t *testing.T) {
	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		mock.ExpectExec(`ALTER DEFAULT PRIVILEGES GRANT SELECT ON TABLES TO joe;`).WillReturnResult(sqlmock.NewResult(1, 1))

		b := NewDefaultPrivilegeBuilder(db, "TABLE", "SELECT", "joe")
		if err := b.Grant(); err != nil {
			t.Fatal(err)
		}
	})
}

func TestDefaultPrivilegeGrantComplex(t *testing.T) {
	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		mock.ExpectExec(`ALTER DEFAULT PRIVILEGES FOR ROLE interns IN DATABASE "dev" GRANT ALL PRIVILEGES ON TABLES TO intern_managers;`).WillReturnResult(sqlmock.NewResult(1, 1))

		b := NewDefaultPrivilegeBuilder(db, "TABLE", "ALL PRIVILEGES", "intern_managers")
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

		b := NewDefaultPrivilegeBuilder(db, "SECRET", "USAGE", "project_managers")
		b.TargetRole("developers")
		if err := b.Revoke(); err != nil {
			t.Fatal(err)
		}
	})
}

func TestDefaultPrivilegeGrantAllRoles(t *testing.T) {
	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		mock.ExpectExec(`ALTER DEFAULT PRIVILEGES FOR ALL ROLES GRANT SELECT ON TABLES TO managers;`).WillReturnResult(sqlmock.NewResult(1, 1))

		b := NewDefaultPrivilegeBuilder(db, "TABLE", "SELECT", "managers")
		b.TargetRole("ALL")
		if err := b.Grant(); err != nil {
			t.Fatal(err)
		}
	})
}
