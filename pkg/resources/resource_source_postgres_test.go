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

var inSourcePostgres = map[string]interface{}{
	"name":                "source",
	"schema_name":         "schema",
	"database_name":       "database",
	"cluster_name":        "cluster",
	"size":                "small",
	"postgres_connection": []interface{}{map[string]interface{}{"name": "pg_connection"}},
	"publication":         "mz_source",
	"text_columns":        []interface{}{"table.unsupported_type_1"},
	"table":               []interface{}{map[string]interface{}{"name": "name", "alias": "alias"}},
}

func TestResourceSourcePostgresCreate(t *testing.T) {
	r := require.New(t)
	d := schema.TestResourceDataRaw(t, SourcePostgres().Schema, inSourcePostgres)
	r.NotNil(d)

	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		// Create
		mock.ExpectExec(
			`CREATE SOURCE "database"."schema"."source" IN CLUSTER "cluster" FROM POSTGRES CONNECTION "database"."schema"."pg_connection" \(PUBLICATION 'mz_source'\) FOR TABLES \(name AS alias\) WITH \(SIZE = 'small'\);`,
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
