package resources

import (
	"context"
	"testing"

	"github.com/MaterializeInc/terraform-provider-materialize/pkg/testhelpers"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/require"
)

var inTable = map[string]interface{}{
	"name":          "table",
	"schema_name":   "schema",
	"database_name": "database",
	"column":        []interface{}{map[string]interface{}{"name": "column", "type": "text", "nullable": true}},
}

var readTable string = `
SELECT
	mz_tables.name,
	mz_schemas.name,
	mz_databases.name
FROM mz_tables
JOIN mz_schemas
	ON mz_tables.schema_id = mz_schemas.id
JOIN mz_databases
	ON mz_schemas.database_id = mz_databases.id
WHERE mz_tables.id = 'u1';`

func TestResourceTableCreate(t *testing.T) {
	r := require.New(t)
	d := schema.TestResourceDataRaw(t, Table().Schema, inTable)
	r.NotNil(d)

	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		// Create
		mock.ExpectExec(`CREATE TABLE "database"."schema"."table" \(column text NOT NULL\);`).WillReturnResult(sqlmock.NewResult(1, 1))

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
		mock.ExpectQuery(readTable).WillReturnRows(ip)

		if err := tableCreate(context.TODO(), d, db); err != nil {
			t.Fatal(err)
		}
	})

}

func TestResourceTableUpdate(t *testing.T) {
	r := require.New(t)
	d := schema.TestResourceDataRaw(t, Table().Schema, inTable)

	// Set current state
	d.SetId("u1")
	d.Set("name", "old_table")
	r.NotNil(d)

	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		mock.ExpectExec(`ALTER TABLE "database"."schema"."old_table" RENAME TO "database"."schema"."table";`).WillReturnResult(sqlmock.NewResult(1, 1))

		// Query Params
		ip := sqlmock.NewRows([]string{"name", "schema", "database"}).AddRow("table", "schema", "database")
		mock.ExpectQuery(readTable).WillReturnRows(ip)

		if err := tableUpdate(context.TODO(), d, db); err != nil {
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

	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		mock.ExpectExec(`DROP TABLE "database"."schema"."table";`).WillReturnResult(sqlmock.NewResult(1, 1))

		if err := tableDelete(context.TODO(), d, db); err != nil {
			t.Fatal(err)
		}
	})

}
