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

var inAws = map[string]interface{}{
	"name":          "conn",
	"schema_name":   "schema",
	"database_name": "database",
	"access_key_id": []interface{}{map[string]interface{}{"text": "foo"}},
}

func TestResourceConnectionAwsCreate(t *testing.T) {
	r := require.New(t)
	d := schema.TestResourceDataRaw(t, ConnectionAws().Schema, inAws)
	r.NotNil(d)

	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		// Create
		mock.ExpectExec(
			`CREATE CONNECTION "database"."schema"."conn"
			TO AWS WITH \( ACCESS KEY ID = 'foo'\);`,
		).WillReturnResult(sqlmock.NewResult(1, 1))

		// Query Id
		ip := `WHERE mz_connections.name = 'conn' AND mz_databases.name = 'database' AND mz_schemas.name = 'schema'`
		testhelpers.MockConnectionScan(mock, ip)

		// Query Params
		pp := `WHERE mz_connections.id = 'u1'`
		testhelpers.MockConnectionAwsScan(mock, pp)

		if err := connectionAwsCreate(context.TODO(), d, db); err != nil {
			t.Fatal(err)
		}
	})

}

// Confirm id is updated with region for 0.4.0
func TestResourceConnectionAwsReadIdMigration(t *testing.T) {
	r := require.New(t)
	d := schema.TestResourceDataRaw(t, ConnectionAws().Schema, inAws)
	r.NotNil(d)

	// Set id before migration
	d.SetId("u1")

	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		// Query Params
		pp := `WHERE mz_connections.id = 'u1'`
		testhelpers.MockConnectionAwsScan(mock, pp)

		if err := connectionAwsRead(context.TODO(), d, db); err != nil {
			t.Fatal(err)
		}

		if d.Id() != "aws/us-east-1:u1" {
			t.Fatalf("unexpected id of %s", d.Id())
		}
	})

}

func TestResourceConnectionAwsUpdate(t *testing.T) {
	r := require.New(t)
	d := schema.TestResourceDataRaw(t, ConnectionAws().Schema, inAws)

	// Set current state
	d.SetId("u1")
	d.Set("name", "old_conn")
	r.NotNil(d)

	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		mock.ExpectExec(`ALTER CONNECTION "database"."schema"."" RENAME TO "conn";`).WillReturnResult(sqlmock.NewResult(1, 1))

		// Query Params
		pp := `WHERE mz_connections.id = 'u1'`
		testhelpers.MockConnectionAwsScan(mock, pp)

		if err := connectionAwsUpdate(context.TODO(), d, db); err != nil {
			t.Fatal(err)
		}
	})

}
