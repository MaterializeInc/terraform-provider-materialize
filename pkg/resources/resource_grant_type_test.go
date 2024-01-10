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

func TestResourceGrantTypeCreate(t *testing.T) {
	utils.SetDefaultRegion("aws/us-east-1")
	r := require.New(t)

	in := map[string]interface{}{
		"role_name":     "joe",
		"privilege":     "USAGE",
		"type_name":     "type",
		"schema_name":   "schema",
		"database_name": "database",
	}
	d := schema.TestResourceDataRaw(t, GrantType().Schema, in)
	r.NotNil(d)

	testhelpers.WithMockProviderMeta(t, func(db *utils.ProviderMeta, mock sqlmock.Sqlmock) {
		// Create
		mock.ExpectExec(
			`GRANT USAGE ON TYPE "database"."schema"."type" TO "joe";`,
		).WillReturnResult(sqlmock.NewResult(1, 1))

		// Query Role Id
		rp := `WHERE mz_roles.name = 'joe'`
		testhelpers.MockRoleScan(mock, rp)

		// Query Grant Id
		gp := `WHERE mz_databases.name = 'database' AND mz_schemas.name = 'schema' AND mz_types.name = 'type'`
		testhelpers.MockTypeScan(mock, gp)

		// Query Params
		pp := `WHERE mz_types.id = 'u1'`
		testhelpers.MockTypeScan(mock, pp)

		if err := grantTypeCreate(context.TODO(), d, db); err != nil {
			t.Fatal(err)
		}

		if d.Id() != "aws/us-east-1:GRANT|TYPE|u1|u1|USAGE" {
			t.Fatalf("unexpected id of %s", d.Id())
		}
	})
}

func TestResourceGrantTypeDelete(t *testing.T) {
	r := require.New(t)

	in := map[string]interface{}{
		"role_name":     "joe",
		"privilege":     "USAGE",
		"type_name":     "type",
		"schema_name":   "schema",
		"database_name": "database",
	}
	d := schema.TestResourceDataRaw(t, GrantType().Schema, in)
	r.NotNil(d)

	testhelpers.WithMockProviderMeta(t, func(db *utils.ProviderMeta, mock sqlmock.Sqlmock) {
		mock.ExpectExec(`REVOKE USAGE ON TYPE "database"."schema"."type" FROM "joe";`).WillReturnResult(sqlmock.NewResult(1, 1))

		if err := grantTypeDelete(context.TODO(), d, db); err != nil {
			t.Fatal(err)
		}
	})
}
