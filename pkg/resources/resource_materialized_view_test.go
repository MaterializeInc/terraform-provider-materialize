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

var inMaterializedView = map[string]interface{}{
	"name":               "materialized_view",
	"schema_name":        "schema",
	"database_name":      "database",
	"cluster_name":       "cluster",
	"not_null_assertion": []interface{}{"column_1", "column_2"},
	"statement":          "SELECT 1 FROM 1",
}

func TestResourceMaterializedViewCreate(t *testing.T) {
	r := require.New(t)
	d := schema.TestResourceDataRaw(t, MaterializedView().Schema, inMaterializedView)
	r.NotNil(d)

	testhelpers.WithMockProviderMeta(t, func(db *utils.ProviderMeta, mock sqlmock.Sqlmock) {
		// Create
		mock.ExpectExec(
			`CREATE MATERIALIZED VIEW "database"."schema"."materialized_view" IN CLUSTER "cluster" WITH \(ASSERT NOT NULL "column_1", ASSERT NOT NULL "column_2"\) AS SELECT 1 FROM 1;`,
		).WillReturnResult(sqlmock.NewResult(1, 1))

		// Query Id
		ip := `WHERE mz_databases.name = 'database' AND mz_materialized_views.name = 'materialized_view' AND mz_schemas.name = 'schema'`
		testhelpers.MockMaterializeViewScan(mock, ip)

		// Query Params
		pp := `WHERE mz_materialized_views.id = 'u1'`
		testhelpers.MockMaterializeViewScan(mock, pp)

		if err := materializedViewCreate(context.TODO(), d, db); err != nil {
			t.Fatal(err)
		}
	})
}

// Confirm id is updated with region for 0.4.0
func TestResourceMaterializedViewReadIdMigration(t *testing.T) {
	r := require.New(t)
	d := schema.TestResourceDataRaw(t, MaterializedView().Schema, inMaterializedView)
	r.NotNil(d)

	// Set id before migration
	d.SetId("u1")

	testhelpers.WithMockProviderMeta(t, func(db *utils.ProviderMeta, mock sqlmock.Sqlmock) {
		// Query Params
		pp := `WHERE mz_materialized_views.id = 'u1'`
		testhelpers.MockMaterializeViewScan(mock, pp)

		if err := materializedViewRead(context.TODO(), d, db); err != nil {
			t.Fatal(err)
		}

		if d.Id() != "aws/us-east-1:u1" {
			t.Fatalf("unexpected id of %s", d.Id())
		}
	})
}

func TestResourceMaterializedViewUpdate(t *testing.T) {
	r := require.New(t)
	d := schema.TestResourceDataRaw(t, MaterializedView().Schema, inMaterializedView)

	// Set current state
	d.SetId("u1")
	d.Set("name", "old_materialized_view")
	r.NotNil(d)

	testhelpers.WithMockProviderMeta(t, func(db *utils.ProviderMeta, mock sqlmock.Sqlmock) {
		mock.ExpectExec(`ALTER MATERIALIZED VIEW "database"."schema"."" RENAME TO "materialized_view";`).WillReturnResult(sqlmock.NewResult(1, 1))

		// Query Params
		pp := `WHERE mz_materialized_views.id = 'u1'`
		testhelpers.MockMaterializeViewScan(mock, pp)

		if err := materializedViewUpdate(context.TODO(), d, db); err != nil {
			t.Fatal(err)
		}
	})
}

func TestResourceMaterializedViewDelete(t *testing.T) {
	r := require.New(t)

	in := map[string]interface{}{
		"name":          "materialized_view",
		"schema_name":   "schema",
		"database_name": "database",
	}
	d := schema.TestResourceDataRaw(t, MaterializedView().Schema, in)
	r.NotNil(d)

	testhelpers.WithMockProviderMeta(t, func(db *utils.ProviderMeta, mock sqlmock.Sqlmock) {
		mock.ExpectExec(`DROP MATERIALIZED VIEW "database"."schema"."materialized_view";`).WillReturnResult(sqlmock.NewResult(1, 1))

		if err := materializedViewDelete(context.TODO(), d, db); err != nil {
			t.Fatal(err)
		}
	})
}
