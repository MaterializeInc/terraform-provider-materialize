package resources

import (
	"context"
	"reflect"
	"testing"

	"github.com/MaterializeInc/terraform-provider-materialize/pkg/testhelpers"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/require"
)

var inSourcePostgresTable = map[string]interface{}{
	"name":          "source",
	"schema_name":   "schema",
	"database_name": "database",
	"cluster_name":  "cluster",
	"size":          "small",
	"postgres_connection": []interface{}{
		map[string]interface{}{
			"name": "pg_connection",
		},
	},
	"publication":  "mz_source",
	"text_columns": []interface{}{"table.unsupported_type_1"},
	"table": []interface{}{
		map[string]interface{}{"name": "name1", "alias": "alias"},
		map[string]interface{}{"name": "name2"},
	},
	"schema": []interface{}{"schema1", "schema2"},
}

func TestResourceSourcePostgresCreateTable(t *testing.T) {
	r := require.New(t)
	d := schema.TestResourceDataRaw(t, SourcePostgres().Schema, inSourcePostgresTable)
	r.NotNil(d)

	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		// Create
		mock.ExpectExec(
			`CREATE SOURCE "database"."schema"."source" IN CLUSTER "cluster" FROM POSTGRES CONNECTION "materialize"."public"."pg_connection" \(PUBLICATION 'mz_source', TEXT COLUMNS \(table.unsupported_type_1\)\) FOR TABLES \(name1 AS alias, name2 AS name2\) WITH \(SIZE = 'small'\);`,
		).WillReturnResult(sqlmock.NewResult(1, 1))

		// Query Id
		ip := `WHERE mz_databases.name = 'database' AND mz_schemas.name = 'schema' AND mz_sources.name = 'source'`
		testhelpers.MockSourceScan(mock, ip)

		// Query Params
		pp := `WHERE mz_sources.id = 'u1'`
		testhelpers.MockSourceScan(mock, pp)

		// Query Subsources
		ps := `WHERE mz_object_dependencies.object_id = 'u1' AND mz_objects.type = 'source'`
		testhelpers.MockSubsourceScan(mock, ps)

		if err := sourcePostgresCreate(context.TODO(), d, db); err != nil {
			t.Fatal(err)
		}
	})
}

var inSourcePostgresSchema = map[string]interface{}{
	"name":          "source",
	"schema_name":   "schema",
	"database_name": "database",
	"cluster_name":  "cluster",
	"size":          "small",
	"postgres_connection": []interface{}{
		map[string]interface{}{
			"name": "pg_connection",
		},
	},
	"publication":  "mz_source",
	"text_columns": []interface{}{"table.unsupported_type_1"},
	"schemas":      []interface{}{"schema1", "schema2"},
}

func TestResourceSourcePostgresCreateSchema(t *testing.T) {
	r := require.New(t)
	d := schema.TestResourceDataRaw(t, SourcePostgres().Schema, inSourcePostgresSchema)
	r.NotNil(d)

	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		// Create
		mock.ExpectExec(
			`CREATE SOURCE "database"."schema"."source" IN CLUSTER "cluster" FROM POSTGRES CONNECTION "materialize"."public"."pg_connection" \(PUBLICATION 'mz_source', TEXT COLUMNS \(table.unsupported_type_1\)\) FOR SCHEMAS \(schema1, schema2\) WITH \(SIZE = 'small'\);`,
		).WillReturnResult(sqlmock.NewResult(1, 1))

		// Query Id
		ip := `WHERE mz_databases.name = 'database' AND mz_schemas.name = 'schema' AND mz_sources.name = 'source'`
		testhelpers.MockSourceScan(mock, ip)

		// Query Params
		pp := `WHERE mz_sources.id = 'u1'`
		testhelpers.MockSourceScan(mock, pp)

		// Query Subsources
		ps := `WHERE mz_object_dependencies.object_id = 'u1' AND mz_objects.type = 'source'`
		testhelpers.MockSubsourceScan(mock, ps)

		if err := sourcePostgresCreate(context.TODO(), d, db); err != nil {
			t.Fatal(err)
		}
	})
}

func TestResourceSourcePostgresUpdate(t *testing.T) {
	r := require.New(t)
	d := schema.TestResourceDataRaw(t, SourcePostgres().Schema, inSourcePostgresTable)

	d.SetId("u1")
	d.Set("name", "old_source")
	d.Set("size", "large")
	r.NotNil(d)

	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		mock.ExpectExec(`ALTER SOURCE "database"."schema"."" RENAME TO "source"`).WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectExec(`ALTER SOURCE "database"."schema"."old_source" SET \(SIZE = 'small'\)`).WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectExec(`ALTER SOURCE "database"."schema"."old_source" ADD SUBSOURCE "name1" AS "alias", "name2"`).WillReturnResult(sqlmock.NewResult(1, 1))

		// Query Params
		pp := `WHERE mz_sources.id = 'u1'`
		testhelpers.MockSourceScan(mock, pp)

		// Query Subsources
		ps := `WHERE mz_object_dependencies.object_id = 'u1' AND mz_objects.type = 'source'`
		testhelpers.MockSubsourceScan(mock, ps)

		if err := sourcePostgresUpdate(context.TODO(), d, db); err != nil {
			t.Fatal(err)
		}
	})
}

func TestDiffTextColumns(t *testing.T) {
	arr1 := []interface{}{"t1.column_1", "t2.column_2"}
	arr2 := []interface{}{"t1.column_1", "t3.column_2"}
	o := diffTextColumns(arr1, arr2)
	e := []string{"t2.column_2"}

	if !reflect.DeepEqual(o, e) {
		t.Fatalf("Expected diffTextColumns to be equal")
	}
}
