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

func TestParseDefaultPrivilegeId(t *testing.T) {
	i, err := parseDefaultPrivilegeKey("aws/us-east-1:GRANT DEFAULT|TABLE|u1||||SELECT")
	if err != nil {
		t.Fatal(err)
	}
	if i.objectType != "TABLE" {
		t.Fatalf("object type %s: does not match expected TABLE", i.objectType)
	}
	if i.granteeId != "u1" {
		t.Fatalf("grantee id %s: does not match expected u1", i.granteeId)
	}
	if i.targetRoleId != "" {
		t.Fatalf("role id %s: expected to be empty string", i.targetRoleId)
	}
	if i.databaseId != "" {
		t.Fatalf("database id %s: expected to be empty string", i.databaseId)
	}
	if i.schemaId != "" {
		t.Fatalf("schema id %s: expected to be empty string", i.schemaId)
	}
}

func TestParseDefaultPrivilegeIdComplex(t *testing.T) {
	i, err := parseDefaultPrivilegeKey("aws/us-east-1:GRANT DEFAULT|TABLE|u1|u2|u3|u4|SELECT")
	if err != nil {
		t.Fatal(err)
	}
	if i.objectType != "TABLE" {
		t.Fatalf("object type %s: does not match expected TABLE", i.objectType)
	}
	if i.granteeId != "u1" {
		t.Fatalf("grantee id %s: does not match expected u1", i.granteeId)
	}
	if i.targetRoleId != "u2" {
		t.Fatalf("role id %s: expected u2", i.targetRoleId)
	}
	if i.databaseId != "u3" {
		t.Fatalf("database id %s: expected u3", i.databaseId)
	}
	if i.schemaId != "u4" {
		t.Fatalf("schema id %s: expected to u4", i.schemaId)
	}
}

func TestParseDefaultPrivilegeIdError(t *testing.T) {
	_, err := parseDefaultPrivilegeKey("incorrect|key|string")
	if err.Error() != "incorrect|key|string: cannot be parsed correctly" {
		t.Fatalf("error not handled")
	}
}

func TestParseDefaultPrivilegeIdErrorEmpty(t *testing.T) {
	_, err := parseDefaultPrivilegeKey("")
	if err.Error() != ": cannot be parsed correctly" {
		t.Fatal(err)
	}
}

// Confirm id is updated with region for 0.4.0
// All resources share the same read function
func TestResourceGrantDefaultPrivilegeReadIdMigration(t *testing.T) {
	utils.SetDefaultRegion("aws/us-east-1")
	r := require.New(t)

	in := map[string]interface{}{
		"grantee_name":     "project_managers",
		"target_role_name": "developers",
		"privilege":        "USAGE",
	}
	d := schema.TestResourceDataRaw(t, GrantSystemPrivilege().Schema, in)
	r.NotNil(d)

	// Set id before migration
	d.SetId("GRANT DEFAULT|CLUSTER|u1|u1|||USAGE")

	testhelpers.WithMockProviderMeta(t, func(db *utils.ProviderMeta, mock sqlmock.Sqlmock) {
		// Query Params
		qp := `
			WHERE mz_default_privileges.grantee = 'u1'
			AND mz_default_privileges.object_type = 'cluster'
			AND mz_default_privileges.role_id = 'u1'`
		testhelpers.MockDefaultPrivilegeScan(mock, qp, "cluster")

		if err := grantDefaultPrivilegeRead(context.TODO(), d, db); err != nil {
			t.Fatal(err)
		}

		if d.Id() != "aws/us-east-1:GRANT DEFAULT|CLUSTER|u1|u1|||USAGE" {
			t.Fatalf("unexpected id of %s", d.Id())
		}
	})
}
