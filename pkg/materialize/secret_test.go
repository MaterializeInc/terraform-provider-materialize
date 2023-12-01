package materialize

import (
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/MaterializeInc/terraform-provider-materialize/pkg/testhelpers"
	"github.com/jmoiron/sqlx"
)

// https://materialize.com/docs/sql/create-secret/

var secret = MaterializeObject{Name: "secret", SchemaName: "schema", DatabaseName: "database"}

func TestSecretCreate(t *testing.T) {
	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		mock.ExpectExec(
			`CREATE SECRET "database"."schema"."secret" AS 'c2VjcmV0Cg';`,
		).WillReturnResult(sqlmock.NewResult(1, 1))

		b := NewSecretBuilder(db, secret)
		b.Value(`c2VjcmV0Cg`)

		if err := b.Create(); err != nil {
			t.Fatal(err)
		}
	})
}

func TestSecretCreateEscapedValue(t *testing.T) {
	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		mock.ExpectExec(
			`CREATE SECRET "database"."schema"."secret" AS 'c2Vjcm''V0Cg';`,
		).WillReturnResult(sqlmock.NewResult(1, 1))

		b := NewSecretBuilder(db, secret)
		b.Value(`c2Vjcm'V0Cg`)

		if err := b.Create(); err != nil {
			t.Fatal(err)
		}
	})
}

func TestSecretRename(t *testing.T) {
	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		mock.ExpectExec(
			`ALTER SECRET "database"."schema"."secret" RENAME TO "new_secret";`,
		).WillReturnResult(sqlmock.NewResult(1, 1))

		b := NewSecretBuilder(db, secret)

		if err := b.Rename("new_secret"); err != nil {
			t.Fatal(err)
		}
	})
}

func TestSecretUpdateValue(t *testing.T) {
	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		mock.ExpectExec(
			`ALTER SECRET "database"."schema"."secret" AS 'c2VjcmV0Cgdd';`,
		).WillReturnResult(sqlmock.NewResult(1, 1))

		b := NewSecretBuilder(db, secret)

		if err := b.UpdateValue(`c2VjcmV0Cgdd`); err != nil {
			t.Fatal(err)
		}
	})
}

func TestSecretUpdateEscapedValue(t *testing.T) {
	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		mock.ExpectExec(
			`ALTER SECRET "database"."schema"."secret" AS 'c2Vjcm''V0Cgdd';`,
		).WillReturnResult(sqlmock.NewResult(1, 1))

		b := NewSecretBuilder(db, secret)

		if err := b.UpdateValue(`c2Vjcm'V0Cgdd`); err != nil {
			t.Fatal(err)
		}
	})
}

func TestSecretDrop(t *testing.T) {
	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		mock.ExpectExec(
			`DROP SECRET "database"."schema"."secret";`,
		).WillReturnResult(sqlmock.NewResult(1, 1))

		if err := NewSecretBuilder(db, secret).Drop(); err != nil {
			t.Fatal(err)
		}
	})
}
