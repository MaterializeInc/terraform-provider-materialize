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

var inConnection = map[string]interface{}{
	"name":          "conn",
	"schema_name":   "schema",
	"database_name": "database",
}

// All connections (other than AWS Privatelink and SSH Tunnel)
// share the same update function
func TestResourceConnectionUpdate(t *testing.T) {
	r := require.New(t)
	d := schema.TestResourceDataRaw(t, ConnectionKafka().Schema, inConnection)

	// Set current state
	d.SetId("u1")
	d.Set("name", "old_conn")
	r.NotNil(d)

	testhelpers.WithMockProviderMeta(t, func(db *utils.ProviderMeta, mock sqlmock.Sqlmock) {
		mock.ExpectExec(`ALTER CONNECTION "database"."schema"."" RENAME TO "conn";`).WillReturnResult(sqlmock.NewResult(1, 1))

		// Query Params
		p := `WHERE mz_connections.id = 'u1'`
		testhelpers.MockConnectionScan(mock, p)

		if err := connectionUpdate(context.TODO(), d, db); err != nil {
			t.Fatal(err)
		}
	})
}

// Confirm id is updated with region for 0.4.0
// All connections (other than AWS Privatelink and SSH Tunnel)
// share the same read function
func TestResourceConnectionReadIdMigration(t *testing.T) {
	utils.SetRegionFromHostname("localhost")
	r := require.New(t)
	d := schema.TestResourceDataRaw(t, ConnectionKafka().Schema, inConnection)
	r.NotNil(d)

	// Set id before migration
	d.SetId("u1")

	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		// Query Params
		p := `WHERE mz_connections.id = 'u1'`
		testhelpers.MockConnectionScan(mock, p)

		if err := connectionRead(context.TODO(), d, db); err != nil {
			t.Fatal(err)
		}

		if d.Id() != "aws/us-east-1:u1" {
			t.Fatalf("unexpected id of %s", d.Id())
		}
	})
}

// All connections share the same delete function
func TestResourceConnectionDelete(t *testing.T) {
	r := require.New(t)

	d := schema.TestResourceDataRaw(t, ConnectionKafka().Schema, inConnection)
	r.NotNil(d)

	testhelpers.WithMockProviderMeta(t, func(db *utils.ProviderMeta, mock sqlmock.Sqlmock) {
		mock.ExpectExec(`DROP CONNECTION "database"."schema"."conn";`).WillReturnResult(sqlmock.NewResult(1, 1))

		if err := connectionDelete(context.TODO(), d, db); err != nil {
			t.Fatal(err)
		}
	})
}
