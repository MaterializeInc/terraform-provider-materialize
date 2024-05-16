package materialize

import (
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/MaterializeInc/terraform-provider-materialize/pkg/testhelpers"
	"github.com/jmoiron/sqlx"
)

var sourcePostgres = MaterializeObject{Name: "source", SchemaName: "schema", DatabaseName: "database"}
var tableInput = []TableStruct{
	{UpstreamName: "table_1"},
	{UpstreamName: "table_2", Name: "table_alias"},
}

func TestSourcePostgresSpecificTablesCreate(t *testing.T) {
	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		mock.ExpectExec(
			`CREATE SOURCE "database"."schema"."source"
			FROM POSTGRES CONNECTION "database"."schema"."pg_connection"
			\(PUBLICATION 'mz_source', TEXT COLUMNS \(table.unsupported_type_1, table.unsupported_type_2\)\)
			FOR TABLES \("schema1"."table_1" AS "database"."schema"."s1_table_1", "schema2"."table_1" AS "database"."schema"."s2_table_1"\)
			EXPOSE PROGRESS AS "database"."schema"."progress";`,
		).WillReturnResult(sqlmock.NewResult(1, 1))

		b := NewSourcePostgresBuilder(db, sourcePostgres)
		b.PostgresConnection(IdentifierSchemaStruct{Name: "pg_connection", SchemaName: "schema", DatabaseName: "database"})
		b.Publication("mz_source")
		b.TextColumns([]string{"table.unsupported_type_1", "table.unsupported_type_2"})
		b.Table([]TableStruct{
			{
				UpstreamName:       "table_1",
				UpstreamSchemaName: "schema1",
				Name:               "s1_table_1",
			},
			{
				UpstreamName:       "table_1",
				UpstreamSchemaName: "schema2",
				Name:               "s2_table_1",
			},
		})
		b.ExposeProgress(IdentifierSchemaStruct{Name: "progress", DatabaseName: "database", SchemaName: "schema"})

		if err := b.Create(); err != nil {
			t.Fatal(err)
		}
	})

}

func TestSourceAddSubsource(t *testing.T) {
	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		mock.ExpectExec(
			`ALTER SOURCE "database"."schema"."source"
			ADD SUBSOURCE "schema"."table_1", "schema"."table_2" AS "database"."schema"."table_alias";`,
		).WillReturnResult(sqlmock.NewResult(1, 1))

		b := NewSource(db, sourcePostgres)
		if err := b.AddSubsource(tableInput, []string{}); err != nil {
			t.Fatal(err)
		}
	})
}

func TestSourceAddSubsourceTextColumns(t *testing.T) {
	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		mock.ExpectExec(
			`ALTER SOURCE "database"."schema"."source"
			ADD SUBSOURCE "schema"."table_1", "schema"."table_2" AS "database"."schema"."table_alias"
			WITH \(TEXT COLUMNS \[table_1.column_1, table_2.column_2\]\);`,
		).WillReturnResult(sqlmock.NewResult(1, 1))

		b := NewSource(db, sourcePostgres)
		if err := b.AddSubsource(tableInput, []string{"table_1.column_1", "table_2.column_2"}); err != nil {
			t.Fatal(err)
		}
	})
}

func TestSourceDropSubsource(t *testing.T) {
	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		mock.ExpectExec(
			`DROP SOURCE "database"."schema"."table_1", "database"."schema"."table_alias";`,
		).WillReturnResult(sqlmock.NewResult(1, 1))

		b := NewSourcePostgresBuilder(db, sourcePostgres)
		if err := b.DropSubsource(tableInput); err != nil {
			t.Fatal(err)
		}
	})
}
