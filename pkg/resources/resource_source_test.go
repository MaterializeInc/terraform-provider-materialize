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

// Confirm id is updated with region for 0.4.0
func TestResourceSourceReadIdMigration(t *testing.T) {
	r := require.New(t)
	d := schema.TestResourceDataRaw(t, SourcePostgres().Schema, inSourcePostgresTable)
	r.NotNil(d)

	// Set current state
	d.SetId("u1")

	testhelpers.WithMockProviderMeta(t, func(db *utils.ProviderMeta, mock sqlmock.Sqlmock) {
		// Query Params
		pp := `WHERE mz_sources.id = 'u1'`
		testhelpers.MockSourceScan(mock, pp)

		// Query Subsources
		ps := `WHERE mz_object_dependencies.object_id = 'u1' AND mz_objects.type = 'source'`
		testhelpers.MockSubsourceScan(mock, ps)

		if err := sourceRead(context.TODO(), d, db); err != nil {
			t.Fatal(err)
		}

		if d.Id() != "aws/us-east-1:u1" {
			t.Fatalf("unexpected id of %s", d.Id())
		}
	})
}

func TestResourceSourceUpdate(t *testing.T) {
	r := require.New(t)
	d := schema.TestResourceDataRaw(t, SourcePostgres().Schema, inSourcePostgresTable)

	// Set current state
	d.SetId("u1")
	d.Set("name", "old_source")
	d.Set("size", "medium")
	r.NotNil(d)

	testhelpers.WithMockProviderMeta(t, func(db *utils.ProviderMeta, mock sqlmock.Sqlmock) {
		mock.ExpectExec(`ALTER SOURCE "database"."schema"."" RENAME TO "source";`).WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectExec(`ALTER SOURCE "database"."schema"."old_source" SET \(SIZE = 'small'\);`).WillReturnResult(sqlmock.NewResult(1, 1))
		// Query Params
		pp := `WHERE mz_sources.id = 'u1'`
		testhelpers.MockSourceScan(mock, pp)

		// Query Subsources
		ps := `WHERE mz_object_dependencies.object_id = 'u1' AND mz_objects.type = 'source'`
		testhelpers.MockSubsourceScan(mock, ps)

		if err := sourceUpdate(context.TODO(), d, db); err != nil {
			t.Fatal(err)
		}
	})
}

func TestResourceSourceDelete(t *testing.T) {
	r := require.New(t)

	in := map[string]interface{}{
		"name":          "source",
		"schema_name":   "schema",
		"database_name": "database",
	}
	d := schema.TestResourceDataRaw(t, SourcePostgres().Schema, in)
	r.NotNil(d)

	testhelpers.WithMockProviderMeta(t, func(db *utils.ProviderMeta, mock sqlmock.Sqlmock) {
		mock.ExpectExec(`DROP SOURCE "database"."schema"."source"`).WillReturnResult(sqlmock.NewResult(1, 1))

		if err := sourceDelete(context.TODO(), d, db); err != nil {
			t.Fatal(err)
		}
	})
}
