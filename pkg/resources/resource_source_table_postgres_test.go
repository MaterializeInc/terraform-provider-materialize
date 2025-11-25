package resources

import (
	"context"
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/MaterializeInc/terraform-provider-materialize/pkg/testhelpers"
	"github.com/MaterializeInc/terraform-provider-materialize/pkg/utils"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/require"
)

var inSourceTablePostgres = map[string]interface{}{
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
	"exclude_columns":       []interface{}{"exclude1", "exclude2"},
}

func TestResourceSourceTablePostgresCreate(t *testing.T) {
	r := require.New(t)
	d := schema.TestResourceDataRaw(t, SourceTablePostgres().Schema, inSourceTablePostgres)
	r.NotNil(d)

	testhelpers.WithMockProviderMeta(t, func(db *utils.ProviderMeta, mock sqlmock.Sqlmock) {
		// Create
		mock.ExpectExec(`CREATE TABLE "database"."schema"."table"
            FROM SOURCE "materialize"."public"."source"
            \(REFERENCE "upstream_schema"."upstream_table"\)
            WITH \(TEXT COLUMNS \("column1", "column2"\)\);`).
			WillReturnResult(sqlmock.NewResult(1, 1))

		// Query Id
		ip := `WHERE mz_databases.name = 'database' AND mz_schemas.name = 'schema' AND mz_tables.name = 'table'`
		testhelpers.MockSourceTablePostgresScan(mock, ip)

		// Query Params
		pp := `WHERE mz_tables.id = 'u1'`
		testhelpers.MockSourceTablePostgresScan(mock, pp)

		if err := sourceTablePostgresCreate(context.TODO(), d, db); err != nil {
			t.Fatal(err)
		}
	})
}

func TestResourceSourceTablePostgresRead(t *testing.T) {
	r := require.New(t)
	d := schema.TestResourceDataRaw(t, SourceTablePostgres().Schema, inSourceTablePostgres)
	d.SetId("u1")
	r.NotNil(d)

	testhelpers.WithMockProviderMeta(t, func(db *utils.ProviderMeta, mock sqlmock.Sqlmock) {
		// Query Params
		pp := `WHERE mz_tables.id = 'u1'`
		testhelpers.MockSourceTablePostgresScan(mock, pp)

		if err := sourceTablePostgresRead(context.TODO(), d, db); err != nil {
			t.Fatal(err)
		}

		r.Equal("table", d.Get("name").(string))
		r.Equal("schema", d.Get("schema_name").(string))
		r.Equal("database", d.Get("database_name").(string))
	})
}

func TestResourceSourceTablePostgresUpdate(t *testing.T) {
	r := require.New(t)
	d := schema.TestResourceDataRaw(t, SourceTablePostgres().Schema, inSourceTablePostgres)
	d.SetId("u1")
	d.Set("name", "old_table")
	r.NotNil(d)

	testhelpers.WithMockProviderMeta(t, func(db *utils.ProviderMeta, mock sqlmock.Sqlmock) {
		mock.ExpectExec(`ALTER TABLE "database"."schema"."" RENAME TO "database"."schema"."table"`).WillReturnResult(sqlmock.NewResult(1, 1))

		// Query Params
		pp := `WHERE mz_tables.id = 'u1'`
		testhelpers.MockSourceTablePostgresScan(mock, pp)

		if err := sourceTablePostgresUpdate(context.TODO(), d, db); err != nil {
			t.Fatal(err)
		}
	})
}

func TestResourceSourceTablePostgresCreateWithExcludeColumns(t *testing.T) {
	r := require.New(t)
	in := map[string]interface{}{
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
		"exclude_columns":       []interface{}{"exclude1", "exclude2"},
	}
	d := schema.TestResourceDataRaw(t, SourceTablePostgres().Schema, in)
	r.NotNil(d)

	testhelpers.WithMockProviderMeta(t, func(db *utils.ProviderMeta, mock sqlmock.Sqlmock) {
		// Create
		mock.ExpectExec(`CREATE TABLE "database"."schema"."table"
            FROM SOURCE "materialize"."public"."source"
            \(REFERENCE "upstream_schema"."upstream_table"\)
            WITH \(EXCLUDE COLUMNS \("exclude1", "exclude2"\)\);`).
			WillReturnResult(sqlmock.NewResult(1, 1))

		// Query Id
		ip := `WHERE mz_databases.name = 'database' AND mz_schemas.name = 'schema' AND mz_tables.name = 'table'`
		testhelpers.MockSourceTablePostgresScan(mock, ip)

		// Query Params
		pp := `WHERE mz_tables.id = 'u1'`
		testhelpers.MockSourceTablePostgresScan(mock, pp)

		if err := sourceTablePostgresCreate(context.TODO(), d, db); err != nil {
			t.Fatal(err)
		}
	})
}

func TestResourceSourceTablePostgresCreateWithTextAndExcludeColumns(t *testing.T) {
	r := require.New(t)
	in := map[string]interface{}{
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
		"exclude_columns":       []interface{}{"exclude1", "exclude2"},
	}
	d := schema.TestResourceDataRaw(t, SourceTablePostgres().Schema, in)
	r.NotNil(d)

	testhelpers.WithMockProviderMeta(t, func(db *utils.ProviderMeta, mock sqlmock.Sqlmock) {
		// Create
		mock.ExpectExec(`CREATE TABLE "database"."schema"."table"
            FROM SOURCE "materialize"."public"."source"
            \(REFERENCE "upstream_schema"."upstream_table"\)
            WITH \(TEXT COLUMNS \("column1", "column2"\), EXCLUDE COLUMNS \("exclude1", "exclude2"\)\);`).
			WillReturnResult(sqlmock.NewResult(1, 1))

		// Query Id
		ip := `WHERE mz_databases.name = 'database' AND mz_schemas.name = 'schema' AND mz_tables.name = 'table'`
		testhelpers.MockSourceTablePostgresScan(mock, ip)

		// Query Params
		pp := `WHERE mz_tables.id = 'u1'`
		testhelpers.MockSourceTablePostgresScan(mock, pp)

		if err := sourceTablePostgresCreate(context.TODO(), d, db); err != nil {
			t.Fatal(err)
		}
	})
}

func TestResourceSourceTablePostgresDelete(t *testing.T) {
	r := require.New(t)
	d := schema.TestResourceDataRaw(t, SourceTablePostgres().Schema, inSourceTablePostgres)
	d.SetId("u1")
	r.NotNil(d)

	testhelpers.WithMockProviderMeta(t, func(db *utils.ProviderMeta, mock sqlmock.Sqlmock) {
		mock.ExpectExec(`DROP TABLE "database"."schema"."table"`).WillReturnResult(sqlmock.NewResult(1, 1))

		if err := sourceTableDelete(context.TODO(), d, db); err != nil {
			t.Fatal(err)
		}
	})
}
