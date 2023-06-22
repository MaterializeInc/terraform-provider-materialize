package materialize

import (
	"testing"

	"github.com/MaterializeInc/terraform-provider-materialize/pkg/testhelpers"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
)

func TestOwnershipAlter(t *testing.T) {
	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		mock.ExpectExec(`ALTER TABLE "database"."schema"."table" OWNER TO "my_role";`).WillReturnResult(sqlmock.NewResult(1, 1))

		o := ObjectSchemaStruct{
			DatabaseName: "database",
			SchemaName:   "schema",
			Name:         "table",
		}
		b := NewOwnershipBuilder(db, "TABLE", o)

		if err := b.Alter("my_role"); err != nil {
			t.Fatal(err)
		}
	})
}
