package resources

import (
	"context"
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/MaterializeInc/terraform-provider-materialize/pkg/testhelpers"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/require"
)

func TestParseDefaultPrivilegeId(t *testing.T) {
	i, err := parseDefaultPrivilegeKey("GRANT DEFAULT|TABLE|u1||||SELECT")
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
	i, err := parseDefaultPrivilegeKey("GRANT DEFAULT|TABLE|u1|u2|u3|u4|SELECT")
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

func TestResourceGrantDefaultPrivilegeCreate(t *testing.T) {
	r := require.New(t)

	in := map[string]interface{}{
		"grantee_name":     "project_managers",
		"target_role_name": "developers",
		"object_type":      "SECRET",
		"privilege":        "USAGE",
	}
	d := schema.TestResourceDataRaw(t, GrantDefaultPrivilege().Schema, in)
	r.NotNil(d)

	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		// Create
		mock.ExpectExec(
			`ALTER DEFAULT PRIVILEGES FOR ROLE developers GRANT USAGE ON SECRETS TO project_managers;`,
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
			AND mz_default_privileges.object_type = 'SECRET'
			AND mz_default_privileges.role_id = 'u1'`
		testhelpers.MockDefaultPrivilegeScan(mock, qp)

		if err := grantDefaultPrivilegeCreate(context.TODO(), d, db); err != nil {
			t.Fatal(err)
		}

		if d.Id() != "GRANT DEFAULT|SECRET|u1|u1|||USAGE" {
			t.Fatalf("unexpected id of %s", d.Id())
		}
	})
}

func TestResourceGrantDefaultPrivilegeDelete(t *testing.T) {
	r := require.New(t)

	in := map[string]interface{}{
		"grantee_name":     "project_managers",
		"target_role_name": "developers",
		"object_type":      "SECRET",
		"privilege":        "USAGE",
	}
	d := schema.TestResourceDataRaw(t, GrantDefaultPrivilege().Schema, in)
	r.NotNil(d)

	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		mock.ExpectExec(`ALTER DEFAULT PRIVILEGES FOR ROLE developers REVOKE USAGE ON SECRETS FROM project_managers;`).WillReturnResult(sqlmock.NewResult(1, 1))

		if err := grantDefaultPrivilegeDelete(context.TODO(), d, db); err != nil {
			t.Fatal(err)
		}
	})
}
