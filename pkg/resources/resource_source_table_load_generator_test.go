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

var inSourceTableLoadGen = map[string]interface{}{
	"name":          "table",
	"schema_name":   "schema",
	"database_name": "database",
	"source": []interface{}{
		map[string]interface{}{
			"name":          "loadgen",
			"schema_name":   "public",
			"database_name": "materialize",
		},
	},
	"upstream_name":        "upstream_table",
	"upstream_schema_name": "upstream_schema",
	"text_columns":         []interface{}{"column1", "column2"},
	"ignore_columns":       []interface{}{"column3", "column4"},
}

func TestResourceSourceTableLoadGenCreate(t *testing.T) {
	r := require.New(t)
	d := schema.TestResourceDataRaw(t, SourceTableLoadGen().Schema, inSourceTableLoadGen)
	r.NotNil(d)

	testhelpers.WithMockProviderMeta(t, func(db *utils.ProviderMeta, mock sqlmock.Sqlmock) {
		// Create
		mock.ExpectExec(
			`CREATE TABLE "database"."schema"."table"
            FROM SOURCE "materialize"."public"."loadgen"
            \(REFERENCE "upstream_schema"."upstream_table"\);`,
		).WillReturnResult(sqlmock.NewResult(1, 1))

		// Query Id
		ip := `WHERE mz_databases.name = 'database' AND mz_schemas.name = 'schema' AND mz_tables.name = 'table'`
		testhelpers.MockSourceTableScan(mock, ip)

		// Query Params
		pp := `WHERE mz_tables.id = 'u1'`
		testhelpers.MockSourceTableScan(mock, pp)

		if err := sourceTableLoadGenCreate(context.TODO(), d, db); err != nil {
			t.Fatal(err)
		}
	})
}

func TestResourceSourceTableLoadGenRead(t *testing.T) {
	r := require.New(t)
	d := schema.TestResourceDataRaw(t, SourceTableLoadGen().Schema, inSourceTableLoadGen)
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

func TestResourceSourceTableLoadGenUpdate(t *testing.T) {
	r := require.New(t)
	d := schema.TestResourceDataRaw(t, SourceTableLoadGen().Schema, inSourceTableLoadGen)
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

func TestResourceSourceTableLoadGenDelete(t *testing.T) {
	r := require.New(t)
	d := schema.TestResourceDataRaw(t, SourceTableLoadGen().Schema, inSourceTableLoadGen)
	d.SetId("u1")
	r.NotNil(d)

	testhelpers.WithMockProviderMeta(t, func(db *utils.ProviderMeta, mock sqlmock.Sqlmock) {
		mock.ExpectExec(`DROP TABLE "database"."schema"."table"`).WillReturnResult(sqlmock.NewResult(1, 1))

		if err := sourceTableDelete(context.TODO(), d, db); err != nil {
			t.Fatal(err)
		}
	})
}
