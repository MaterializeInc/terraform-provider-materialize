package materialize

import (
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/MaterializeInc/terraform-provider-materialize/pkg/testhelpers"
	"github.com/jmoiron/sqlx"
)

// https://materialize.com/docs/sql/create-role/

func TestRoleCreate(t *testing.T) {
	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		mock.ExpectExec(
			`CREATE ROLE "role" WITH INHERIT;`,
		).WillReturnResult(sqlmock.NewResult(1, 1))

		o := MaterializeObject{Name: "role"}
		b := NewRoleBuilder(db, o)
		b.Inherit()

		if err := b.Create(); err != nil {
			t.Fatal(err)
		}
	})
}

func TestRoleAlter(t *testing.T) {
	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		mock.ExpectExec(
			`ALTER ROLE "role" INHERIT;`,
		).WillReturnResult(sqlmock.NewResult(1, 1))

		o := MaterializeObject{Name: "role"}
		if err := NewRoleBuilder(db, o).Alter("INHERIT"); err != nil {
			t.Fatal(err)
		}
	})
}

func TestRoleCreateWithSuperuser(t *testing.T) {
	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		mock.ExpectExec(
			`CREATE ROLE "role" WITH INHERIT SUPERUSER;`,
		).WillReturnResult(sqlmock.NewResult(1, 1))

		o := MaterializeObject{Name: "role"}
		b := NewRoleBuilder(db, o)
		b.Inherit()
		b.Superuser(true)

		if err := b.Create(); err != nil {
			t.Fatal(err)
		}
	})
}

func TestRoleCreateWithNoSuperuser(t *testing.T) {
	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		mock.ExpectExec(
			`CREATE ROLE "role" WITH INHERIT NOSUPERUSER;`,
		).WillReturnResult(sqlmock.NewResult(1, 1))

		o := MaterializeObject{Name: "role"}
		b := NewRoleBuilder(db, o)
		b.Inherit()
		b.Superuser(false)

		if err := b.Create(); err != nil {
			t.Fatal(err)
		}
	})
}

func TestRoleCreateWithLogin(t *testing.T) {
	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		mock.ExpectExec(
			`CREATE ROLE "role" WITH INHERIT LOGIN;`,
		).WillReturnResult(sqlmock.NewResult(1, 1))

		o := MaterializeObject{Name: "role"}
		b := NewRoleBuilder(db, o)
		b.Inherit()
		b.Login(true)

		if err := b.Create(); err != nil {
			t.Fatal(err)
		}
	})
}

func TestRoleCreateWithPasswordAndLogin(t *testing.T) {
	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		mock.ExpectExec(
			`CREATE ROLE "role" WITH INHERIT LOGIN PASSWORD 'password123';`,
		).WillReturnResult(sqlmock.NewResult(1, 1))

		o := MaterializeObject{Name: "role"}
		b := NewRoleBuilder(db, o)
		b.Inherit()
		b.Password("password123")
		b.Login(true)

		if err := b.Create(); err != nil {
			t.Fatal(err)
		}
	})
}

func TestRoleCreateWithPasswordNoLogin(t *testing.T) {
	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		mock.ExpectExec(
			`CREATE ROLE "role" WITH INHERIT PASSWORD 'password123';`,
		).WillReturnResult(sqlmock.NewResult(1, 1))

		o := MaterializeObject{Name: "role"}
		b := NewRoleBuilder(db, o)
		b.Inherit()
		b.Password("password123")
		b.Login(false)

		if err := b.Create(); err != nil {
			t.Fatal(err)
		}
	})
}

func TestRoleCreateWithAllOptions(t *testing.T) {
	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		mock.ExpectExec(
			`CREATE ROLE "role" WITH INHERIT LOGIN PASSWORD 'password123' SUPERUSER;`,
		).WillReturnResult(sqlmock.NewResult(1, 1))

		o := MaterializeObject{Name: "role"}
		b := NewRoleBuilder(db, o)
		b.Inherit()
		b.Password("password123")
		b.Login(true)
		b.Superuser(true)

		if err := b.Create(); err != nil {
			t.Fatal(err)
		}
	})
}

func TestRoleCreateNoOptions(t *testing.T) {
	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		mock.ExpectExec(
			`CREATE ROLE "role";`,
		).WillReturnResult(sqlmock.NewResult(1, 1))

		o := MaterializeObject{Name: "role"}
		b := NewRoleBuilder(db, o)

		if err := b.Create(); err != nil {
			t.Fatal(err)
		}
	})
}

func TestRoleAlterLogin(t *testing.T) {
	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		mock.ExpectExec(
			`ALTER ROLE "role" LOGIN;`,
		).WillReturnResult(sqlmock.NewResult(1, 1))

		o := MaterializeObject{Name: "role"}
		if err := NewRoleBuilder(db, o).AlterLogin(true); err != nil {
			t.Fatal(err)
		}
	})
}

func TestRoleAlterNoLogin(t *testing.T) {
	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		mock.ExpectExec(
			`ALTER ROLE "role" NOLOGIN;`,
		).WillReturnResult(sqlmock.NewResult(1, 1))

		o := MaterializeObject{Name: "role"}
		if err := NewRoleBuilder(db, o).AlterLogin(false); err != nil {
			t.Fatal(err)
		}
	})
}

func TestRoleDrop(t *testing.T) {
	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		mock.ExpectExec(
			`DROP ROLE "role";`,
		).WillReturnResult(sqlmock.NewResult(1, 1))

		o := MaterializeObject{Name: "role"}
		if err := NewRoleBuilder(db, o).Drop(); err != nil {
			t.Fatal(err)
		}
	})
}

func TestListRolesWithoutPattern(t *testing.T) {
	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		testhelpers.MockRoleScan(mock, "")

		if _, err := ListRoles(db, ""); err != nil {
			t.Fatal(err)
		}
	})
}

func TestListRolesWithPattern(t *testing.T) {
	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		testhelpers.MockRoleScan(mock, "WHERE mz_roles.name LIKE 'prod_%'")

		if _, err := ListRoles(db, "prod_%"); err != nil {
			t.Fatal(err)
		}
	})
}
