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

var inSourceLoadgen = map[string]interface{}{
	"name":                "source",
	"schema_name":         "schema",
	"database_name":       "database",
	"cluster_name":        "cluster",
	"size":                "small",
	"load_generator_type": "TPCH",
	"tpch_options": []interface{}{map[string]interface{}{
		"tick_interval": "1s",
		"scale_factor":  0.5,
		"table":         []interface{}{map[string]interface{}{"name": "name", "alias": "alias"}}}},
}

func TestResourceSourceLoadgenCreate(t *testing.T) {
	r := require.New(t)
	d := schema.TestResourceDataRaw(t, SourceLoadgen().Schema, inSourceLoadgen)
	r.NotNil(d)

	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		// Create
		mock.ExpectExec(
			`CREATE SOURCE "database"."schema"."source" IN CLUSTER "cluster" FROM LOAD GENERATOR TPCH \(TICK INTERVAL '1s', SCALE FACTOR 0.50\) FOR TABLES \(name AS alias\) WITH \(SIZE = 'small'\);`,
		).WillReturnResult(sqlmock.NewResult(1, 1))

		// Query Id
		ir := mock.NewRows([]string{"id"}).AddRow("u1")
		mock.ExpectQuery(`
			SELECT mz_sources.id
			FROM mz_sources
			JOIN mz_schemas
				ON mz_sources.schema_id = mz_schemas.id
			JOIN mz_databases
				ON mz_schemas.database_id = mz_databases.id
			LEFT JOIN mz_connections
				ON mz_sources.connection_id = mz_connections.id
			JOIN mz_clusters
				ON mz_sources.cluster_id = mz_clusters.id
			WHERE mz_sources.name = 'source'
			AND mz_schemas.name = 'schema'
			AND mz_databases.name = 'database';
		`).WillReturnRows(ir)

		// Query Params
		mockSourceParams(mock)

		if err := sourceLoadgenCreate(context.TODO(), d, db); err != nil {
			t.Fatal(err)
		}
	})

}

func TestResourceSourceLoadgenUpdate(t *testing.T) {
	r := require.New(t)
	d := schema.TestResourceDataRaw(t, SourceLoadgen().Schema, inSourceLoadgen)

	// Set current state
	d.SetId("u1")
	d.Set("name", "old_source")
	d.Set("size", "medium")
	r.NotNil(d)

	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		mock.ExpectExec(`ALTER SOURCE "database"."schema"."old_source" SET \(SIZE = 'small'\);`).WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectExec(`ALTER SOURCE "database"."schema"."old_source" RENAME TO "database"."schema"."source";`).WillReturnResult(sqlmock.NewResult(1, 1))

		// Query Params
		mockSourceParams(mock)

		if err := sourceUpdate(context.TODO(), d, db); err != nil {
			t.Fatal(err)
		}
	})

}

func TestResourceSourceLoadgenDelete(t *testing.T) {
	r := require.New(t)

	in := map[string]interface{}{
		"name":          "source",
		"schema_name":   "schema",
		"database_name": "database",
	}
	d := schema.TestResourceDataRaw(t, SourceLoadgen().Schema, in)
	r.NotNil(d)

	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		mock.ExpectExec(`DROP SOURCE "database"."schema"."source";`).WillReturnResult(sqlmock.NewResult(1, 1))

		if err := sourceDelete(context.TODO(), d, db); err != nil {
			t.Fatal(err)
		}
	})

}
