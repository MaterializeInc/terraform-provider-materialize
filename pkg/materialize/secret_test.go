package materialize

import (
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/MaterializeInc/terraform-provider-materialize/pkg/testhelpers"
	"github.com/jmoiron/sqlx"
)

func TestSecretCreate(t *testing.T) {
	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		mock.ExpectExec(
			`CREATE SECRET "database"."schema"."secret" AS 'c2VjcmV0Cg';`,
		).WillReturnResult(sqlmock.NewResult(1, 1))
		b := NewSecretBuilder(db, "secret", "schema", "database")
		b.Value(`c2VjcmV0Cg`)

		b.Create()
	})
}

func TestSecretCreateEscapedValue(t *testing.T) {
	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		mock.ExpectExec(
			`CREATE SECRET "database"."schema"."secret" AS 'c2Vjcm''V0Cg';`,
		).WillReturnResult(sqlmock.NewResult(1, 1))

		b := NewSecretBuilder(db, "secret", "schema", "database")
		b.Value(`c2Vjcm'V0Cg`)

		b.Create()
	})
}

func TestSecretRename(t *testing.T) {
	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		mock.ExpectExec(
			`ALTER SECRET "database"."schema"."secret" RENAME TO "database"."schema"."new_secret";`,
		).WillReturnResult(sqlmock.NewResult(1, 1))

		b := NewSecretBuilder(db, "secret", "schema", "database")

		b.Rename("new_secret")
	})
}

func TestSecretUpdateValue(t *testing.T) {
	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		mock.ExpectExec(
			`ALTER SECRET "database"."schema"."secret" AS 'c2VjcmV0Cgdd';`,
		).WillReturnResult(sqlmock.NewResult(1, 1))

		b := NewSecretBuilder(db, "secret", "schema", "database")

		b.UpdateValue(`c2VjcmV0Cgdd`)
	})
}

func TestSecretUpdateEscapedValue(t *testing.T) {
	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		mock.ExpectExec(
			`ALTER SECRET "database"."schema"."secret" AS 'c2Vjcm''V0Cgdd';`,
		).WillReturnResult(sqlmock.NewResult(1, 1))

		b := NewSecretBuilder(db, "secret", "schema", "database")

		b.UpdateValue(`c2Vjcm'V0Cgdd`)
	})
}

func TestSecretDrop(t *testing.T) {
	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		mock.ExpectExec(
			`DROP SECRET "database"."schema"."secret";`,
		).WillReturnResult(sqlmock.NewResult(1, 1))

		b := NewSecretBuilder(db, "secret", "schema", "database")

		b.Drop()
	})
}
