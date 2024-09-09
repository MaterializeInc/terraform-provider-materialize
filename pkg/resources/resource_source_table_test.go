package resources

import (
	"context"
	"testing"

	"github.com/MaterializeInc/terraform-provider-materialize/pkg/testhelpers"
	"github.com/MaterializeInc/terraform-provider-materialize/pkg/utils"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/require"
)

var inSourceTable = map[string]interface{}{
	"name":          "table",
	"schema_name":   "schema",
	"database_name": "database",
	"source": []interface{}{
		map[string]interface{}{
			"name":          "source",
			"schema_name":   "public",
			"database_name": "materialize",
		},
	},
	"upstream_name":        "upstream_table",
	"upstream_schema_name": "upstream_schema",
	"text_columns":         []interface{}{"column1", "column2"},
	"ignore_columns":       []interface{}{"column3", "column4"},
}

func TestResourceSourceTableCreate(t *testing.T) {
	r := require.New(t)
	d := schema.TestResourceDataRaw(t, SourceTable().Schema, inSourceTable)
	r.NotNil(d)

	testhelpers.WithMockProviderMeta(t, func(db *utils.ProviderMeta, mock sqlmock.Sqlmock) {
		// Expect source type query
		sourceTypeQuery := `WHERE mz_databases.name = 'materialize' AND mz_schemas.name = 'public' AND mz_sources.name = 'source'`
		testhelpers.MockSourceScanWithType(mock, sourceTypeQuery, "mysql")

		// Create
		mock.ExpectExec(
			`CREATE TABLE "database"."schema"."table"
            FROM SOURCE "materialize"."public"."source"
            \(REFERENCE "upstream_schema"."upstream_table"\)
            WITH \(TEXT COLUMNS \(column1, column2\), IGNORE COLUMNS \(column3, column4\)\);`,
		).WillReturnResult(sqlmock.NewResult(1, 1))

		// Query Id
		ip := `WHERE mz_databases.name = 'database' AND mz_schemas.name = 'schema' AND mz_tables.name = 'table'`
		testhelpers.MockSourceTableScan(mock, ip)

		// Query Params
		pp := `WHERE mz_tables.id = 'u1'`
		testhelpers.MockSourceTableScan(mock, pp)

		if err := sourceTableCreate(context.TODO(), d, db); err != nil {
			t.Fatal(err)
		}
	})
}

func TestResourceSourceTableCreateNonMySQL(t *testing.T) {
	r := require.New(t)
	d := schema.TestResourceDataRaw(t, SourceTable().Schema, inSourceTable)
	r.NotNil(d)

	testhelpers.WithMockProviderMeta(t, func(db *utils.ProviderMeta, mock sqlmock.Sqlmock) {
		// Expect source type query
		sourceTypeQuery := `WHERE mz_databases.name = 'materialize' AND mz_schemas.name = 'public' AND mz_sources.name = 'source'`
		testhelpers.MockSourceScan(mock, sourceTypeQuery)

		// Create (without IGNORE COLUMNS)
		mock.ExpectExec(`CREATE TABLE "database"."schema"."table"
            FROM SOURCE "materialize"."public"."source"
            \(REFERENCE "upstream_schema"."upstream_table"\)
            WITH \(TEXT COLUMNS \(column1, column2\)\);`).
			WillReturnResult(sqlmock.NewResult(1, 1))

		// Query Id
		ip := `WHERE mz_databases.name = 'database' AND mz_schemas.name = 'schema' AND mz_tables.name = 'table'`
		testhelpers.MockSourceTableScan(mock, ip)

		// Query Params
		pp := `WHERE mz_tables.id = 'u1'`
		testhelpers.MockSourceTableScan(mock, pp)

		if err := sourceTableCreate(context.TODO(), d, db); err != nil {
			t.Fatal(err)
		}
	})
}

func TestResourceSourceTableRead(t *testing.T) {
	r := require.New(t)
	d := schema.TestResourceDataRaw(t, SourceTable().Schema, inSourceTable)
	d.SetId("u1")
	r.NotNil(d)

	testhelpers.WithMockProviderMeta(t, func(db *utils.ProviderMeta, mock sqlmock.Sqlmock) {
		// Query Params
		pp := `WHERE mz_tables.id = 'u1'`
		testhelpers.MockSourceTableScan(mock, pp)

		if err := sourceTableRead(context.TODO(), d, db); err != nil {
			t.Fatal(err)
		}

		r.Equal("table", d.Get("name").(string))
		r.Equal("schema", d.Get("schema_name").(string))
		r.Equal("database", d.Get("database_name").(string))
	})
}

func TestResourceSourceTableUpdate(t *testing.T) {
	r := require.New(t)
	d := schema.TestResourceDataRaw(t, SourceTable().Schema, inSourceTable)
	d.SetId("u1")
	d.Set("name", "old_table")
	r.NotNil(d)

	testhelpers.WithMockProviderMeta(t, func(db *utils.ProviderMeta, mock sqlmock.Sqlmock) {
		mock.ExpectExec(`ALTER TABLE "database"."schema"."" RENAME TO "database"."schema"."table"`).WillReturnResult(sqlmock.NewResult(1, 1))

		// Query Params
		pp := `WHERE mz_tables.id = 'u1'`
		testhelpers.MockSourceTableScan(mock, pp)

		if err := sourceTableUpdate(context.TODO(), d, db); err != nil {
			t.Fatal(err)
		}
	})
}

func TestResourceSourceTableDelete(t *testing.T) {
	r := require.New(t)
	d := schema.TestResourceDataRaw(t, SourceTable().Schema, inSourceTable)
	d.SetId("u1")
	r.NotNil(d)

	testhelpers.WithMockProviderMeta(t, func(db *utils.ProviderMeta, mock sqlmock.Sqlmock) {
		mock.ExpectExec(`DROP TABLE "database"."schema"."table"`).WillReturnResult(sqlmock.NewResult(1, 1))

		if err := sourceTableDelete(context.TODO(), d, db); err != nil {
			t.Fatal(err)
		}
	})
}
