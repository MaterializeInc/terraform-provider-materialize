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

var inSourceMySQLTable = map[string]interface{}{
	"name":          "source",
	"schema_name":   "schema",
	"database_name": "database",
	"cluster_name":  "cluster",
	"mysql_connection": []interface{}{
		map[string]interface{}{
			"name": "mysql_connection",
		},
	},
	"ignore_columns": []interface{}{"column1", "column2"},
	"text_columns":   []interface{}{"column3", "column4"},
	"table": []interface{}{
		map[string]interface{}{"name": "name1", "alias": "alias"},
		map[string]interface{}{"name": "name2"},
	},
}

func TestResourceSourceMySQLCreate(t *testing.T) {
	r := require.New(t)
	d := schema.TestResourceDataRaw(t, SourceMySQL().Schema, inSourceMySQLTable)
	r.NotNil(d)

	testhelpers.WithMockProviderMeta(t, func(db *utils.ProviderMeta, mock sqlmock.Sqlmock) {
		// Create
		mock.ExpectExec(
			`CREATE SOURCE "database"."schema"."source" IN CLUSTER "cluster" FROM MYSQL CONNECTION "materialize"."public"."mysql_connection" \(IGNORE COLUMNS \(column1, column2\), TEXT COLUMNS \(column3, column4\)\) FOR TABLES \(name2 AS name2, name1 AS alias\);`,
		).WillReturnResult(sqlmock.NewResult(1, 1))

		// Query Id
		ip := `WHERE mz_databases.name = 'database' AND mz_schemas.name = 'schema' AND mz_sources.name = 'source'`
		testhelpers.MockSourceScan(mock, ip)

		// Query Params
		pp := `WHERE mz_sources.id = 'u1'`
		testhelpers.MockSourceScan(mock, pp)

		// Query Subsources (added this mock expectation)
		ps := `WHERE mz_object_dependencies.object_id = 'u1' AND mz_objects.type = 'source'`
		testhelpers.MockSubsourceScan(mock, ps)

		if err := sourceMySQLCreate(context.TODO(), d, db); err != nil {
			t.Fatal(err)
		}

		r.NoError(mock.ExpectationsWereMet())
	})
}

func TestResourceSourceMySQLUpdate(t *testing.T) {
	r := require.New(t)
	d := schema.TestResourceDataRaw(t, SourceMySQL().Schema, inSourceMySQLTable)

	d.SetId("u1")
	d.Set("name", "old_source")
	r.NotNil(d)

	testhelpers.WithMockProviderMeta(t, func(db *utils.ProviderMeta, mock sqlmock.Sqlmock) {
		mock.ExpectExec(`ALTER SOURCE "database"."schema"."" RENAME TO "source"`).WillReturnResult(sqlmock.NewResult(1, 1))
		// mock.ExpectExec(`ALTER SOURCE "database"."schema"."source" ADD SUBSOURCE "name1" AS "alias", "name2"`).WillReturnResult(sqlmock.NewResult(1, 1))

		// Query Params
		pp := `WHERE mz_sources.id = 'u1'`
		testhelpers.MockSourceScan(mock, pp)

		// Query Subsources
		ps := `WHERE mz_object_dependencies.object_id = 'u1' AND mz_objects.type = 'source'`
		testhelpers.MockSubsourceScan(mock, ps)

		if err := sourceMySQLUpdate(context.TODO(), d, db); err != nil {
			t.Fatal(err)
		}
	})
}
