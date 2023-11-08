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

func TestResourceGrantConnectionCreate(t *testing.T) {
	utils.SetRegionFromHostname("localhost")
	r := require.New(t)

	in := map[string]interface{}{
		"role_name":       "joe",
		"privilege":       "USAGE",
		"connection_name": "conn",
		"schema_name":     "schema",
		"database_name":   "database",
	}
	d := schema.TestResourceDataRaw(t, GrantConnection().Schema, in)
	r.NotNil(d)

	testhelpers.WithMockProviderMeta(t, func(db *utils.ProviderMeta, mock sqlmock.Sqlmock) {
		// Create
		mock.ExpectExec(
			`GRANT USAGE ON CONNECTION "database"."schema"."conn" TO "joe";`,
		).WillReturnResult(sqlmock.NewResult(1, 1))

		// Query Role Id
		rp := `WHERE mz_roles.name = 'joe'`
		testhelpers.MockRoleScan(mock, rp)

		// Query Grant Id
		gp := `WHERE mz_connections.name = 'conn' AND mz_databases.name = 'database' AND mz_schemas.name = 'schema'`
		testhelpers.MockConnectionScan(mock, gp)

		// Query Params
		pp := `WHERE mz_connections.id = 'u1'`
		testhelpers.MockConnectionScan(mock, pp)

		if err := grantConnectionCreate(context.TODO(), d, db); err != nil {
			t.Fatal(err)
		}

		if d.Id() != "aws/us-east-1:GRANT|CONNECTION|u1|u1|USAGE" {
			t.Fatalf("unexpected id of %s", d.Id())
		}
	})
}

func TestResourceGrantConnectionDelete(t *testing.T) {
	r := require.New(t)

	in := map[string]interface{}{
		"role_name":       "joe",
		"privilege":       "USAGE",
		"connection_name": "conn",
		"schema_name":     "schema",
		"database_name":   "database",
	}
	d := schema.TestResourceDataRaw(t, GrantConnection().Schema, in)
	r.NotNil(d)

	testhelpers.WithMockProviderMeta(t, func(db *utils.ProviderMeta, mock sqlmock.Sqlmock) {
		mock.ExpectExec(`REVOKE USAGE ON CONNECTION "database"."schema"."conn" FROM "joe";`).WillReturnResult(sqlmock.NewResult(1, 1))

		if err := grantConnectionDelete(context.TODO(), d, db); err != nil {
			t.Fatal(err)
		}
	})
}
