package materialize

import (
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/MaterializeInc/terraform-provider-materialize/pkg/testhelpers"
	"github.com/jmoiron/sqlx"
)

var sourceMySQL = MaterializeObject{Name: "source", SchemaName: "schema", DatabaseName: "database"}
var tableInputMySQL = []TableStruct{
	{UpstreamName: "table_1"},
	{UpstreamName: "table_2", Name: "table_alias"},
}

func TestSourceMySQLAllTablesCreate(t *testing.T) {
	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		mock.ExpectExec(
			`CREATE SOURCE "database"."schema"."source" FROM MYSQL CONNECTION "database"."schema"."mysql_connection" FOR ALL TABLES;`,
		).WillReturnResult(sqlmock.NewResult(1, 1))

		b := NewSourceMySQLBuilder(db, sourceMySQL)
		b.MySQLConnection(IdentifierSchemaStruct{Name: "mysql_connection", SchemaName: "schema", DatabaseName: "database"})
		b.AllTables()

		if err := b.Create(); err != nil {
			t.Fatal(err)
		}
	})
}

func TestSourceMySQLSpecificTablesCreate(t *testing.T) {
	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		mock.ExpectExec(
			`CREATE SOURCE "database"."schema"."source" FROM MYSQL CONNECTION "database"."schema"."mysql_connection" FOR TABLES \("schema1"."table_1" AS "database"."schema"."s1_table_1", "schema2"."table_2" AS "database"."schema"."table_alias"\) EXPOSE PROGRESS AS "database"."schema"."progress";`,
		).WillReturnResult(sqlmock.NewResult(1, 1))

		b := NewSourceMySQLBuilder(db, sourceMySQL)
		b.MySQLConnection(IdentifierSchemaStruct{Name: "mysql_connection", SchemaName: "schema", DatabaseName: "database"})
		b.Tables([]TableStruct{
			{
				UpstreamName:       "table_1",
				UpstreamSchemaName: "schema1",
				Name:               "s1_table_1",
			},
			{
				UpstreamName:       "table_2",
				UpstreamSchemaName: "schema2",
				Name:               "table_alias",
			},
		})
		b.ExposeProgress(IdentifierSchemaStruct{Name: "progress", DatabaseName: "database", SchemaName: "schema"})

		if err := b.Create(); err != nil {
			t.Fatal(err)
		}
	})
}

func TestSourceMySQLAddSubsource(t *testing.T) {
	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		mock.ExpectExec(
			`ALTER SOURCE "database"."schema"."source" ADD SUBSOURCE "schema"."table_1", "schema"."table_2" AS "database"."schema"."table_alias";`,
		).WillReturnResult(sqlmock.NewResult(1, 1))

		b := NewSource(db, sourceMySQL)
		if err := b.AddSubsource(tableInputMySQL, []string{}); err != nil {
			t.Fatal(err)
		}
	})
}

func TestSourceMySQLDropSubsource(t *testing.T) {
	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		mock.ExpectExec(
			`DROP SOURCE "database"."schema"."table_1", "database"."schema"."table_alias"`,
		).WillReturnResult(sqlmock.NewResult(1, 1))

		b := NewSourceMySQLBuilder(db, sourceMySQL)
		if err := b.DropSubsource(tableInputMySQL); err != nil {
			t.Fatal(err)
		}
	})
}
