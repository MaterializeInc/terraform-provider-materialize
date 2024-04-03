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

var inAwsPrivatelink = map[string]interface{}{
	"name":               "conn",
	"schema_name":        "schema",
	"database_name":      "database",
	"service_name":       "service",
	"availability_zones": []interface{}{"use1-az1", "use1-az2"},
}

func TestResourceConnectionAwsPrivatelinkCreate(t *testing.T) {
	r := require.New(t)
	d := schema.TestResourceDataRaw(t, ConnectionAwsPrivatelink().Schema, inAwsPrivatelink)
	r.NotNil(d)

	testhelpers.WithMockProviderMeta(t, func(db *utils.ProviderMeta, mock sqlmock.Sqlmock) {
		// Create
		mock.ExpectExec(
			`CREATE CONNECTION "database"."schema"."conn"
			TO AWS PRIVATELINK \(SERVICE NAME 'service',AVAILABILITY ZONES \('use1-az1', 'use1-az2'\)\)`,
		).WillReturnResult(sqlmock.NewResult(1, 1))

		// Query Id
		ip := `WHERE mz_connections.name = 'conn' AND mz_databases.name = 'database' AND mz_schemas.name = 'schema'`
		testhelpers.MockConnectionScan(mock, ip)

		// Query Params
		pp := `WHERE mz_connections.id = 'u1'`
		testhelpers.MockConnectionAwsPrivatelinkScan(mock, pp)

		if err := connectionAwsPrivatelinkCreate(context.TODO(), d, db); err != nil {
			t.Fatal(err)
		}
	})

}

// Confirm id is updated with region for 0.4.0
func TestResourceConnectionAwsPrivatelinkReadIdMigration(t *testing.T) {
	r := require.New(t)
	d := schema.TestResourceDataRaw(t, ConnectionAwsPrivatelink().Schema, inAwsPrivatelink)
	r.NotNil(d)

	// Set id before migration
	d.SetId("u1")

	testhelpers.WithMockProviderMeta(t, func(db *utils.ProviderMeta, mock sqlmock.Sqlmock) {
		// Query Params
		pp := `WHERE mz_connections.id = 'u1'`
		testhelpers.MockConnectionAwsPrivatelinkScan(mock, pp)

		if err := connectionAwsPrivatelinkRead(context.TODO(), d, db); err != nil {
			t.Fatal(err)
		}

		if d.Id() != "aws/us-east-1:u1" {
			t.Fatalf("unexpected id of %s", d.Id())
		}
	})

}

func TestResourceConnectionAwsPrivatelinkUpdate(t *testing.T) {
	r := require.New(t)
	d := schema.TestResourceDataRaw(t, ConnectionAwsPrivatelink().Schema, inAwsPrivatelink)

	// Set current state
	d.SetId("u1")
	d.Set("name", "old_conn")
	r.NotNil(d)

	testhelpers.WithMockProviderMeta(t, func(db *utils.ProviderMeta, mock sqlmock.Sqlmock) {
		mock.ExpectExec(`ALTER CONNECTION "database"."schema"."" RENAME TO "conn";`).WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectExec(`ALTER CONNECTION "database"."schema"."old_conn" SET \(SERVICE NAME = 'service'\) WITH \(validate false\);`).WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectExec(`ALTER CONNECTION "database"."schema"."old_conn" SET \(AVAILABILITY ZONES = \['use1-az1', 'use1-az2'\]\) WITH \(validate false\);`).WillReturnResult(sqlmock.NewResult(1, 1))

		// Query Params
		pp := `WHERE mz_connections.id = 'u1'`
		testhelpers.MockConnectionAwsPrivatelinkScan(mock, pp)

		if err := connectionAwsPrivatelinkUpdate(context.TODO(), d, db); err != nil {
			t.Fatal(err)
		}
	})

}
