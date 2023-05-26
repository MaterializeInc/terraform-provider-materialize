package materialize

import (
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/MaterializeInc/terraform-provider-materialize/pkg/testhelpers"
	"github.com/jmoiron/sqlx"
)

func TestSchemaCreate(t *testing.T) {
	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		mock.ExpectExec(
			`CREATE SCHEMA "database"."schema";`,
		).WillReturnResult(sqlmock.NewResult(1, 1))

		if err := NewSchemaBuilder(db, "schema", "database").Create(); err != nil {
			t.Fatal(err)
		}
	})
}

func TestSchemaDrop(t *testing.T) {
	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		mock.ExpectExec(
			`DROP SCHEMA "database"."schema";`,
		).WillReturnResult(sqlmock.NewResult(1, 1))

		if err := NewSchemaBuilder(db, "schema", "database").Drop(); err != nil {
			t.Fatal(err)
		}
	})
}
