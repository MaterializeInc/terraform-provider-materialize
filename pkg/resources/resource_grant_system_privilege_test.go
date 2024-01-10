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

func TestResourceGrantSystemPrivilegeCreate(t *testing.T) {
	utils.SetDefaultRegion("aws/us-east-1")
	r := require.New(t)

	in := map[string]interface{}{
		"role_name": "role",
		"privilege": "CREATEDB",
	}
	d := schema.TestResourceDataRaw(t, GrantSystemPrivilege().Schema, in)
	r.NotNil(d)

	testhelpers.WithMockProviderMeta(t, func(db *utils.ProviderMeta, mock sqlmock.Sqlmock) {
		// Create
		mock.ExpectExec(
			`GRANT CREATEDB ON SYSTEM TO "role";`,
		).WillReturnResult(sqlmock.NewResult(1, 1))

		// SystemPrivilegeId - Query role
		gp := `WHERE mz_roles.name = 'role'`
		testhelpers.MockRoleScan(mock, gp)

		// Query Params
		testhelpers.MockSystemGrantScan(mock)

		if err := grantSystemPrivilegeCreate(context.TODO(), d, db); err != nil {
			t.Fatal(err)
		}

		if d.Id() != "aws/us-east-1:GRANT SYSTEM|u1|CREATEDB" {
			t.Fatalf("unexpected id of %s", d.Id())
		}
	})
}

// Confirm id is updated with region for 0.4.0
func TestResourceGrantSystemPrivilegeReadIdMigration(t *testing.T) {
	utils.SetDefaultRegion("aws/us-east-1")
	r := require.New(t)

	in := map[string]interface{}{
		"role_name": "role",
		"privilege": "CREATEDB",
	}
	d := schema.TestResourceDataRaw(t, GrantSystemPrivilege().Schema, in)
	r.NotNil(d)

	// Set id before migration
	d.SetId("GRANT SYSTEM|u1|CREATEDB")

	testhelpers.WithMockProviderMeta(t, func(db *utils.ProviderMeta, mock sqlmock.Sqlmock) {
		// Query Params
		testhelpers.MockSystemGrantScan(mock)

		if err := grantSystemPrivilegeRead(context.TODO(), d, db); err != nil {
			t.Fatal(err)
		}

		if d.Id() != "aws/us-east-1:GRANT SYSTEM|u1|CREATEDB" {
			t.Fatalf("unexpected id of %s", d.Id())
		}
	})
}

func TestResourceGrantSystemPrivilegeDelete(t *testing.T) {
	r := require.New(t)

	in := map[string]interface{}{
		"role_name": "role",
		"privilege": "CREATEDB",
	}
	d := schema.TestResourceDataRaw(t, GrantSystemPrivilege().Schema, in)
	r.NotNil(d)

	testhelpers.WithMockProviderMeta(t, func(db *utils.ProviderMeta, mock sqlmock.Sqlmock) {
		mock.ExpectExec(`REVOKE CREATEDB ON SYSTEM FROM "role";`).WillReturnResult(sqlmock.NewResult(1, 1))

		if err := grantSystemPrivilegeDelete(context.TODO(), d, db); err != nil {
			t.Fatal(err)
		}
	})
}
