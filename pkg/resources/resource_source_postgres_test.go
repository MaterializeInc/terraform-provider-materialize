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
		map[string]interface{}{"name": "name1", "alias": "alias"},
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
			FOR TABLES \(name1 AS alias\)`,
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
		mock.ExpectExec(`ALTER SOURCE "database"."schema"."old_source" ADD SUBSOURCE "name1" AS "alias"`).WillReturnResult(sqlmock.NewResult(1, 1))

		// Query Params
		pp := `WHERE mz_sources.id = 'u1'`
		testhelpers.MockSourceScan(mock, pp)

		// Query Tables
		pt := `WHERE mz_object_dependencies.referenced_object_id = 'u1' AND mz_sources.type = 'subsource'`
		testhelpers.MockPosgresSubsourceScan(mock, pt)

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
