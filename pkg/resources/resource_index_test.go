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

func TestResourceIndexCreate(t *testing.T) {
	r := require.New(t)

	in := map[string]interface{}{
		"name":         "index",
		"default":      false,
		"obj_name":     []interface{}{map[string]interface{}{"name": "source", "schema_name": "schema", "database_name": "database"}},
		"cluster_name": "cluster",
	}
	d := schema.TestResourceDataRaw(t, Index().Schema, in)
	r.NotNil(d)

	testhelpers.WithMockProviderMeta(t, func(db *utils.ProviderMeta, mock sqlmock.Sqlmock) {
		// Create
		mock.ExpectExec(
			`CREATE INDEX index IN CLUSTER cluster ON "database"."schema"."source" USING ARRANGEMENT \(\);`,
		).WillReturnResult(sqlmock.NewResult(1, 1))

		// Query Id
		ip := `WHERE mz_indexes.name = 'index' AND mz_objects.type IN \('source', 'view', 'materialized-view'\)`
		testhelpers.MockIndexScan(mock, ip)

		// Query Params
		pp := `WHERE mz_indexes.id = 'u1' AND mz_objects.type IN \('source', 'view', 'materialized-view'\)`
		testhelpers.MockIndexScan(mock, pp)

		// Query Columns
		cp := `WHERE mz_indexes.id = 'u1'`
		testhelpers.MockIndexColumnScan(mock, cp)

		if err := indexCreate(context.TODO(), d, db); err != nil {
			t.Fatal(err)
		}
	})
}

// Confirm id is updated with region for 0.4.0
func TestResourceIndexReadIdMigration(t *testing.T) {
	r := require.New(t)

	in := map[string]interface{}{
		"name": "index",
	}
	d := schema.TestResourceDataRaw(t, Index().Schema, in)
	r.NotNil(d)

	// Set id before migration
	d.SetId("u1")

	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		// Query Params
		pp := `WHERE mz_indexes.id = 'u1' AND mz_objects.type IN \('source', 'view', 'materialized-view'\)`
		testhelpers.MockIndexScan(mock, pp)

		// Query Columns
		cp := `WHERE mz_indexes.id = 'u1'`
		testhelpers.MockIndexColumnScan(mock, cp)

		if err := indexRead(context.TODO(), d, db); err != nil {
			t.Fatal(err)
		}

		if d.Id() != "aws/us-east-1:u1" {
			t.Fatalf("unexpected id of %s", d.Id())
		}
	})
}

func TestResourceIndexDelete(t *testing.T) {
	r := require.New(t)

	in := map[string]interface{}{
		"name":     "index",
		"default":  false,
		"obj_name": []interface{}{map[string]interface{}{"name": "source", "schema_name": "schema", "database_name": "database"}},
	}
	d := schema.TestResourceDataRaw(t, Index().Schema, in)
	r.NotNil(d)

	testhelpers.WithMockProviderMeta(t, func(db *utils.ProviderMeta, mock sqlmock.Sqlmock) {
		mock.ExpectExec(`DROP INDEX "database"."schema"."index" RESTRICT;`).WillReturnResult(sqlmock.NewResult(1, 1))

		if err := indexDelete(context.TODO(), d, db); err != nil {
			t.Fatal(err)
		}
	})
}
