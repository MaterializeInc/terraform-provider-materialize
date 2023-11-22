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
	"name":           "table",
	"schema_name":    "schema",
	"database_name":  "database",
	"ownership_role": "joe",
	"comment":        "object comment",
	"column": []interface{}{map[string]interface{}{
		"name":     "column",
		"type":     "text",
		"nullable": true,
		"comment":  "column comment",
	},
	},
}

func TestResourceTableCreate(t *testing.T) {
	r := require.New(t)
	d := schema.TestResourceDataRaw(t, Table().Schema, inTable)
	r.NotNil(d)

	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		// Create
		mock.ExpectExec(`
			CREATE TABLE "database"."schema"."table" \(column text NOT NULL\);
		`).WillReturnResult(sqlmock.NewResult(1, 1))

		// Ownership
		mock.ExpectExec(`ALTER TABLE "database"."schema"."table" OWNER TO "joe";`).WillReturnResult(sqlmock.NewResult(1, 1))

		// Comment
		mock.ExpectExec(`COMMENT ON TABLE "database"."schema"."table" IS 'object comment';`).WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectExec(`COMMENT ON COLUMN "database"."schema"."table"."column" IS 'column comment';`).WillReturnResult(sqlmock.NewResult(1, 1))

		// Query Id
		ip := `WHERE mz_databases.name = 'database' AND mz_schemas.name = 'schema' AND mz_tables.name = 'table'`
		testhelpers.MockTableScan(mock, ip)

		// Query Params
		pp := `WHERE mz_tables.id = 'u1'`
		testhelpers.MockTableScan(mock, pp)

		// Query Columns
		cp := `WHERE mz_columns.id = 'u1'`
		testhelpers.MockTableColumnScan(mock, cp)

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

	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		mock.ExpectExec(`DROP TABLE "database"."schema"."table";`).WillReturnResult(sqlmock.NewResult(1, 1))

		if err := tableDelete(context.TODO(), d, db); err != nil {
			t.Fatal(err)
		}
	})
}
