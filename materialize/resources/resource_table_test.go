package resources

import (
	"context"
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/require"
)

func TestResourceTableCreate(t *testing.T) {
	r := require.New(t)

	in := map[string]interface{}{
		"name":          "table",
		"schema_name":   "schema",
		"database_name": "database",
		"temporary":     true,
		"columns":       []interface{}{map[string]interface{}{"col_name": "column", "col_type": "text", "not_null": true}},
	}
	d := schema.TestResourceDataRaw(t, Table().Schema, in)
	r.NotNil(d)

	WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		// Create
		mock.ExpectExec(`CREATE TEMPORARY TABLE "database"."schema"."table" \(column text NOT NULL\);`).WillReturnResult(sqlmock.NewResult(1, 1))

		// Query Id
		ir := mock.NewRows([]string{"id"}).AddRow("u1")
		mock.ExpectQuery(`
			SELECT mz_tables.id
			FROM mz_tables
			JOIN mz_schemas
				ON mz_tables.schema_id = mz_schemas.id
			JOIN mz_databases
				ON mz_schemas.database_id = mz_databases.id
			WHERE mz_tables.name = 'table'
			AND mz_schemas.name = 'schema'
			AND mz_databases.name = 'database';
		`).WillReturnRows(ir)

		// Query Params
		ip := sqlmock.NewRows([]string{"name", "schema", "database"}).AddRow("table", "schema", "database")
		mock.ExpectQuery(`
			SELECT
				mz_tables.name,
				mz_schemas.name,
				mz_databases.name
			FROM mz_tables
			JOIN mz_schemas
				ON mz_tables.schema_id = mz_schemas.id
			JOIN mz_databases
				ON mz_schemas.database_id = mz_databases.id
			WHERE mz_tables.id = 'u1';
		`).WillReturnRows(ip)

		if err := tableCreate(context.TODO(), d, db); err != nil {
			t.Fatal(err)
		}
	})

}

func TestResourceTableDelete(t *testing.T) {
	r := require.New(t)

	in := map[string]interface{}{
		"name":          "table",
		"schema_name":   "schema",
		"database_name": "database",
	}
	d := schema.TestResourceDataRaw(t, Table().Schema, in)
	r.NotNil(d)

	WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		mock.ExpectExec(`DROP TABLE "database"."schema"."table";`).WillReturnResult(sqlmock.NewResult(1, 1))

		if err := tableDelete(context.TODO(), d, db); err != nil {
			t.Fatal(err)
		}
	})

}

func TestTableCreateQuery(t *testing.T) {
	r := require.New(t)
	b := newTableBuilder("table", "schema", "database")
	b.Temporary()
	b.Columns([]TableColumn{
		{
			colName: "column_1",
			colType: "int",
		},
		{
			colName: "column_2",
			colType: "text",
			notNull: true,
		},
	})
	r.Equal(`CREATE TEMPORARY TABLE "database"."schema"."table" (column_1 int, column_2 text NOT NULL);`, b.Create())
}

func TestTableRenameQuery(t *testing.T) {
	r := require.New(t)
	b := newSourceKafkaBuilder("table", "schema", "database")
	r.Equal(`ALTER SOURCE "database"."schema"."table" RENAME TO "database"."schema"."new_table";`, b.Rename("new_table"))
}

func TestTableDropQuery(t *testing.T) {
	r := require.New(t)
	b := newTableBuilder("table", "schema", "database")
	r.Equal(`DROP TABLE "database"."schema"."table";`, b.Drop())
}
