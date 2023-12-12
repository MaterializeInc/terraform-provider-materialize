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

func TestResourceGrantClusterDefaultPrivilegeCreate(t *testing.T) {
	utils.SetDefaultRegion("aws/us-east-1")
	r := require.New(t)

	in := map[string]interface{}{
		"grantee_name":     "project_managers",
		"target_role_name": "developers",
		"privilege":        "USAGE",
	}
	d := schema.TestResourceDataRaw(t, GrantClusterDefaultPrivilege().Schema, in)
	r.NotNil(d)

	testhelpers.WithMockProviderMeta(t, func(db *utils.ProviderMeta, mock sqlmock.Sqlmock) {
		// Create
		mock.ExpectExec(
			`ALTER DEFAULT PRIVILEGES FOR ROLE "developers" GRANT USAGE ON CLUSTERS TO "project_managers";`,
		).WillReturnResult(sqlmock.NewResult(1, 1))

		// DefaultPrivilegeId - Query grantee role
		gp := `WHERE mz_roles.name = 'project_managers'`
		testhelpers.MockRoleScan(mock, gp)

		// DefaultPrivilegeId - Query target role
		tp := `WHERE mz_roles.name = 'developers'`
		testhelpers.MockRoleScan(mock, tp)

		// Query Params
		qp := `
			WHERE mz_default_privileges.grantee = 'u1'
			AND mz_default_privileges.object_type = 'cluster'
			AND mz_default_privileges.role_id = 'u1'`
		testhelpers.MockDefaultPrivilegeScan(mock, qp, "cluster")

		if err := grantClusterDefaultPrivilegeCreate(context.TODO(), d, db); err != nil {
			t.Fatal(err)
		}

		if d.Id() != "aws/us-east-1:GRANT DEFAULT|CLUSTER|u1|u1|||USAGE" {
			t.Fatalf("unexpected id of %s", d.Id())
		}
	})
}

func TestResourceGrantClusterDefaultPrivilegeDelete(t *testing.T) {
	r := require.New(t)

	in := map[string]interface{}{
		"grantee_name":     "project_managers",
		"target_role_name": "developers",
		"privilege":        "USAGE",
	}
	d := schema.TestResourceDataRaw(t, GrantClusterDefaultPrivilege().Schema, in)
	r.NotNil(d)

	testhelpers.WithMockProviderMeta(t, func(db *utils.ProviderMeta, mock sqlmock.Sqlmock) {
		mock.ExpectExec(`ALTER DEFAULT PRIVILEGES FOR ROLE "developers" REVOKE USAGE ON CLUSTERS FROM "project_managers";`).WillReturnResult(sqlmock.NewResult(1, 1))

		if err := grantClusterDefaultPrivilegeDelete(context.TODO(), d, db); err != nil {
			t.Fatal(err)
		}
	})
}
