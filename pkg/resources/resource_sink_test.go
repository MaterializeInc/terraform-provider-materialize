package resources

import (
	"context"
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/MaterializeInc/terraform-provider-materialize/pkg/testhelpers"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/require"
)

// Confirm id is updated with region for 0.4.0
func TestResourceSinkReadIdMigration(t *testing.T) {
	r := require.New(t)
	d := schema.TestResourceDataRaw(t, SinkKafka().Schema, inSinkKafka)
	r.NotNil(d)

	// Set current state
	d.SetId("u1")

	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		// Query Params
		pp := `WHERE mz_sinks.id = 'u1'`
		testhelpers.MockSinkScan(mock, pp)

		if err := sinkRead(context.TODO(), d, db); err != nil {
			t.Fatal(err)
		}

		if d.Id() != "aws/us-east-1:u1" {
			t.Fatalf("unexpected id of %s", d.Id())
		}
	})
}

func TestResourceSinkUpdate(t *testing.T) {
	r := require.New(t)
	d := schema.TestResourceDataRaw(t, SinkKafka().Schema, inSinkKafka)

	// Set current state
	d.SetId("u1")
	d.Set("name", "old_sink")
	d.Set("size", "medium")
	r.NotNil(d)

	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		mock.ExpectExec(`ALTER SINK "database"."schema"."" RENAME TO "sink";`).WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectExec(`ALTER SINK "database"."schema"."old_sink" SET \(SIZE = 'small'\);`).WillReturnResult(sqlmock.NewResult(1, 1))

		// Query Params
		pp := `WHERE mz_sinks.id = 'u1'`
		testhelpers.MockSinkScan(mock, pp)

		if err := sinkUpdate(context.TODO(), d, db); err != nil {
			t.Fatal(err)
		}
	})
}

func TestResourceSinkDelete(t *testing.T) {
	r := require.New(t)

	in := map[string]interface{}{
		"name":          "sink",
		"schema_name":   "schema",
		"database_name": "database",
	}
	d := schema.TestResourceDataRaw(t, SinkKafka().Schema, in)
	r.NotNil(d)

	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		mock.ExpectExec(`DROP SINK "database"."schema"."sink";`).WillReturnResult(sqlmock.NewResult(1, 1))

		if err := sinkDelete(context.TODO(), d, db); err != nil {
			t.Fatal(err)
		}
	})
}
