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

func TestResourceGrantDatabaseCreate(t *testing.T) {
	utils.SetDefaultRegion("aws/us-east-1")
	r := require.New(t)

	in := map[string]interface{}{
		"role_name":     "joe",
		"privilege":     "CREATE",
		"database_name": "materialize",
	}
	d := schema.TestResourceDataRaw(t, GrantDatabase().Schema, in)
	r.NotNil(d)

	testhelpers.WithMockProviderMeta(t, func(db *utils.ProviderMeta, mock sqlmock.Sqlmock) {
		// Create
		mock.ExpectExec(
			`GRANT CREATE ON DATABASE "materialize" TO "joe";`,
		).WillReturnResult(sqlmock.NewResult(1, 1))

		// Query Role Id
		rp := `WHERE mz_roles.name = 'joe'`
		testhelpers.MockRoleScan(mock, rp)

		// Query Grant Id
		gp := `WHERE mz_databases.name = 'materialize'`
		testhelpers.MockDatabaseScan(mock, gp)

		// Query Params
		pp := `WHERE mz_databases.id = 'u1'`
		testhelpers.MockDatabaseScan(mock, pp)

		if err := grantDatabaseCreate(context.TODO(), d, db); err != nil {
			t.Fatal(err)
		}

		if d.Id() != "aws/us-east-1:GRANT|DATABASE|u1|u1|CREATE" {
			t.Fatalf("unexpected id of %s", d.Id())
		}
	})
}

func TestResourceGrantDatabaseCreateEmail(t *testing.T) {
	r := require.New(t)

	in := map[string]interface{}{
		"role_name":     "joe@materialize.com",
		"privilege":     "CREATE",
		"database_name": "materialize",
	}
	d := schema.TestResourceDataRaw(t, GrantDatabase().Schema, in)
	r.NotNil(d)

	testhelpers.WithMockProviderMeta(t, func(db *utils.ProviderMeta, mock sqlmock.Sqlmock) {
		// Create
		mock.ExpectExec(
			`GRANT CREATE ON DATABASE "materialize" TO "joe@materialize.com";`,
		).WillReturnResult(sqlmock.NewResult(1, 1))
	})
}

func TestResourceGrantDatabaseDelete(t *testing.T) {
	r := require.New(t)

	in := map[string]interface{}{
		"role_name":     "joe",
		"privilege":     "CREATE",
		"database_name": "materialize",
	}
	d := schema.TestResourceDataRaw(t, GrantDatabase().Schema, in)
	r.NotNil(d)

	testhelpers.WithMockProviderMeta(t, func(db *utils.ProviderMeta, mock sqlmock.Sqlmock) {
		mock.ExpectExec(`REVOKE CREATE ON DATABASE "materialize" FROM "joe";`).WillReturnResult(sqlmock.NewResult(1, 1))

		if err := grantDatabaseDelete(context.TODO(), d, db); err != nil {
			t.Fatal(err)
		}
	})
}
