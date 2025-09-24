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
			`CREATE ROLE "role" INHERIT;`,
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
			`CREATE ROLE "role" INHERIT SUPERUSER;`,
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
			`CREATE ROLE "role" INHERIT NOSUPERUSER;`,
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
