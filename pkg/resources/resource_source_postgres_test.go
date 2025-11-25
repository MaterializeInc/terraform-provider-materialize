package resources

import (
	"context"
	"reflect"
	"testing"

	"github.com/MaterializeInc/terraform-provider-materialize/pkg/testhelpers"
	"github.com/MaterializeInc/terraform-provider-materialize/pkg/utils"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/require"
)

var inSourcePostgresTable = map[string]interface{}{
	"name":          "source",
	"schema_name":   "schema",
	"database_name": "database",
	"cluster_name":  "cluster",
	"postgres_connection": []interface{}{
		map[string]interface{}{
			"name": "pg_connection",
		},
	},
	"publication":  "mz_source",
	"text_columns": []interface{}{"table.unsupported_type_1"},
	"table": []interface{}{
		map[string]interface{}{"upstream_name": "name1", "name": "local_name"},
	},
}

func TestResourceSourcePostgresCreateTable(t *testing.T) {
	r := require.New(t)
	d := schema.TestResourceDataRaw(t, SourcePostgres().Schema, inSourcePostgresTable)
	r.NotNil(d)

	testhelpers.WithMockProviderMeta(t, func(db *utils.ProviderMeta, mock sqlmock.Sqlmock) {
		// Create
		mock.ExpectExec(
			`CREATE SOURCE "database"."schema"."source"
			IN CLUSTER "cluster"
			FROM POSTGRES CONNECTION "materialize"."public"."pg_connection"
			\(PUBLICATION 'mz_source',
			TEXT COLUMNS \(table.unsupported_type_1\)\)
			FOR TABLES \("schema"."name1" AS "database"."schema"."local_name"\)`,
		).WillReturnResult(sqlmock.NewResult(1, 1))

		// Query Id
		ip := `WHERE mz_databases.name = 'database' AND mz_schemas.name = 'schema' AND mz_sources.name = 'source'`
		testhelpers.MockSourceScan(mock, ip)

		// Query Params
		pp := `WHERE mz_sources.id = 'u1'`
		testhelpers.MockSourceScan(mock, pp)

		// Query Tables
		pt := `WHERE mz_object_dependencies.referenced_object_id = 'u1' AND mz_sources.type = 'subsource'`
		testhelpers.MockPosgresSubsourceScan(mock, pt)

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
	"postgres_connection": []interface{}{
		map[string]interface{}{
			"name": "pg_connection",
		},
	},
	"publication":  "mz_source",
	"text_columns": []interface{}{"table.unsupported_type_1"},
}

func TestResourceSourcePostgresUpdate(t *testing.T) {
	r := require.New(t)
	d := schema.TestResourceDataRaw(t, SourcePostgres().Schema, inSourcePostgresTable)

	d.SetId("u1")
	d.Set("name", "old_source")
	r.NotNil(d)

	testhelpers.WithMockProviderMeta(t, func(db *utils.ProviderMeta, mock sqlmock.Sqlmock) {
		mock.ExpectExec(`ALTER SOURCE "database"."schema"."" RENAME TO "source"`).WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectExec(`ALTER SOURCE "database"."schema"."old_source" ADD SUBSOURCE "schema"."name1" AS "database"."schema"."local_name"`).WillReturnResult(sqlmock.NewResult(1, 1))

		// Query Params
		pp := `WHERE mz_sources.id = 'u1'`
		testhelpers.MockSourceScan(mock, pp)

		// Query Tables
		pt := `WHERE mz_object_dependencies.referenced_object_id = 'u1' AND mz_sources.type = 'subsource'`
		testhelpers.MockPosgresSubsourceScan(mock, pt)

		// Query Subsources
		ps := `WHERE filter_id = 'u1' AND type = 'source'`
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

var inSourcePostgresWithExcludeColumns = map[string]interface{}{
	"name":          "source",
	"schema_name":   "schema",
	"database_name": "database",
	"cluster_name":  "cluster",
	"postgres_connection": []interface{}{
		map[string]interface{}{
			"name": "pg_connection",
		},
	},
	"publication":     "mz_source",
	"exclude_columns": []interface{}{"public.users.image_data", "public.posts.binary_data"},
	"table": []interface{}{
		map[string]interface{}{"upstream_name": "users", "upstream_schema_name": "public", "name": "users"},
		map[string]interface{}{"upstream_name": "posts", "upstream_schema_name": "public", "name": "posts"},
	},
}

func TestResourceSourcePostgresCreateWithExcludeColumns(t *testing.T) {
	r := require.New(t)
	d := schema.TestResourceDataRaw(t, SourcePostgres().Schema, inSourcePostgresWithExcludeColumns)
	r.NotNil(d)

	testhelpers.WithMockProviderMeta(t, func(db *utils.ProviderMeta, mock sqlmock.Sqlmock) {
		// Create
		mock.ExpectExec(
			`CREATE SOURCE "database"."schema"."source" IN CLUSTER "cluster" FROM POSTGRES CONNECTION "materialize"."public"."pg_connection" \(PUBLICATION 'mz_source', EXCLUDE COLUMNS \(public.users.image_data, public.posts.binary_data\)\) FOR TABLES \("public"."users" AS "database"."schema"."users", "public"."posts" AS "database"."schema"."posts"\)`,
		).WillReturnResult(sqlmock.NewResult(1, 1))

		// Query Id
		ip := `WHERE mz_databases.name = 'database' AND mz_schemas.name = 'schema' AND mz_sources.name = 'source'`
		testhelpers.MockSourceScan(mock, ip)

		// Query Params
		pp := `WHERE mz_sources.id = 'u1'`
		testhelpers.MockSourceScan(mock, pp)

		// Query Tables
		pt := `WHERE mz_object_dependencies.referenced_object_id = 'u1' AND mz_sources.type = 'subsource'`
		testhelpers.MockPosgresSubsourceScan(mock, pt)

		if err := sourcePostgresCreate(context.TODO(), d, db); err != nil {
			t.Fatal(err)
		}
	})
}

var inSourcePostgresWithTextAndExcludeColumns = map[string]interface{}{
	"name":          "source",
	"schema_name":   "schema",
	"database_name": "database",
	"cluster_name":  "cluster",
	"postgres_connection": []interface{}{
		map[string]interface{}{
			"name": "pg_connection",
		},
	},
	"publication":     "mz_source",
	"text_columns":    []interface{}{"public.users.description", "public.posts.content"},
	"exclude_columns": []interface{}{"public.users.image_data", "public.posts.binary_data"},
	"table": []interface{}{
		map[string]interface{}{"upstream_name": "users", "upstream_schema_name": "public", "name": "users"},
		map[string]interface{}{"upstream_name": "posts", "upstream_schema_name": "public", "name": "posts"},
	},
}

func TestResourceSourcePostgresCreateWithTextAndExcludeColumns(t *testing.T) {
	r := require.New(t)
	d := schema.TestResourceDataRaw(t, SourcePostgres().Schema, inSourcePostgresWithTextAndExcludeColumns)
	r.NotNil(d)

	testhelpers.WithMockProviderMeta(t, func(db *utils.ProviderMeta, mock sqlmock.Sqlmock) {
		// Create
		mock.ExpectExec(
			`CREATE SOURCE "database"."schema"."source" IN CLUSTER "cluster" FROM POSTGRES CONNECTION "materialize"."public"."pg_connection" \(PUBLICATION 'mz_source', TEXT COLUMNS \(public.users.description, public.posts.content\), EXCLUDE COLUMNS \(public.users.image_data, public.posts.binary_data\)\) FOR TABLES \("public"."users" AS "database"."schema"."users", "public"."posts" AS "database"."schema"."posts"\)`,
		).WillReturnResult(sqlmock.NewResult(1, 1))

		// Query Id
		ip := `WHERE mz_databases.name = 'database' AND mz_schemas.name = 'schema' AND mz_sources.name = 'source'`
		testhelpers.MockSourceScan(mock, ip)

		// Query Params
		pp := `WHERE mz_sources.id = 'u1'`
		testhelpers.MockSourceScan(mock, pp)

		// Query Tables
		pt := `WHERE mz_object_dependencies.referenced_object_id = 'u1' AND mz_sources.type = 'subsource'`
		testhelpers.MockPosgresSubsourceScan(mock, pt)

		if err := sourcePostgresCreate(context.TODO(), d, db); err != nil {
			t.Fatal(err)
		}
	})
}
