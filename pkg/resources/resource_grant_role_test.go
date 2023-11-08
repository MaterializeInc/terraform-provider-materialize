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

func TestResourceGrantRolePrivilegeCreate(t *testing.T) {
	utils.SetRegionFromHostname("localhost")
	r := require.New(t)

	in := map[string]interface{}{
		"role_name":   "role",
		"member_name": "member",
	}
	d := schema.TestResourceDataRaw(t, GrantRole().Schema, in)
	r.NotNil(d)

	testhelpers.WithMockProviderMeta(t, func(db *utils.ProviderMeta, mock sqlmock.Sqlmock) {
		// Create
		mock.ExpectExec(
			`GRANT "role" TO "member";`,
		).WillReturnResult(sqlmock.NewResult(1, 1))

		// RolePrivilegeId - Query role
		gp := `WHERE mz_roles.name = 'role'`
		testhelpers.MockRoleScan(mock, gp)

		// RolePrivilegeId - Query member role
		tp := `WHERE mz_roles.name = 'member'`
		testhelpers.MockRoleScan(mock, tp)

		// Query Params
		testhelpers.MockRoleGrantScan(mock)

		if err := grantRoleCreate(context.TODO(), d, db); err != nil {
			t.Fatal(err)
		}

		if d.Id() != "aws/us-east-1:ROLE MEMBER|u1|u1" {
			t.Fatalf("unexpected id of %s", d.Id())
		}
	})
}

// Confirm id is updated with region for 0.4.0
func TestResourceGrantRolePrivilegeReadIdMigration(t *testing.T) {
	utils.SetRegionFromHostname("localhost")
	r := require.New(t)

	in := map[string]interface{}{
		"role_name":   "role",
		"member_name": "member",
	}
	d := schema.TestResourceDataRaw(t, GrantRole().Schema, in)
	r.NotNil(d)

	// Set id before migration
	d.SetId("ROLE MEMBER|u1|u1")

	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		// Query Params
		testhelpers.MockRoleGrantScan(mock)

		if err := grantRoleRead(context.TODO(), d, db); err != nil {
			t.Fatal(err)
		}

		if d.Id() != "aws/us-east-1:ROLE MEMBER|u1|u1" {
			t.Fatalf("unexpected id of %s", d.Id())
		}
	})
}

func TestResourceGrantRolePrivilegeDelete(t *testing.T) {
	r := require.New(t)

	in := map[string]interface{}{
		"role_name":   "role",
		"member_name": "member",
	}
	d := schema.TestResourceDataRaw(t, GrantRole().Schema, in)
	r.NotNil(d)

	testhelpers.WithMockProviderMeta(t, func(db *utils.ProviderMeta, mock sqlmock.Sqlmock) {
		mock.ExpectExec(`REVOKE "role" FROM "member";`).WillReturnResult(sqlmock.NewResult(1, 1))

		if err := grantRoleDelete(context.TODO(), d, db); err != nil {
			t.Fatal(err)
		}
	})
}
