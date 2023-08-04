package materialize

import (
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/MaterializeInc/terraform-provider-materialize/pkg/testhelpers"
	"github.com/jmoiron/sqlx"
)

var source = ObjectSchemaStruct{
	Name:         "source",
	SchemaName:   "schema",
	DatabaseName: "database",
}
var tableInput = []TableStruct{
	{Name: "table_1"},
	{Name: "table_2", Alias: "table_alias"},
}

func TestSourceAddSubsource(t *testing.T) {
	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		mock.ExpectExec(
			`ALTER SOURCE "database"."schema"."source" ADD SUBSOURCE table_1, table_2 AS table_alias;`,
		).WillReturnResult(sqlmock.NewResult(1, 1))

		b := NewSource(db, source)
		if err := b.AddSubsource(tableInput); err != nil {
			t.Fatal(err)
		}
	})
}

func TestSourceDropSubsource(t *testing.T) {
	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		mock.ExpectExec(
			`ALTER SOURCE "database"."schema"."source" DROP SUBSOURCE table_1, table_alias;`,
		).WillReturnResult(sqlmock.NewResult(1, 1))

		b := NewSource(db, source)
		if err := b.DropSubsource(tableInput); err != nil {
			t.Fatal(err)
		}
	})
}
