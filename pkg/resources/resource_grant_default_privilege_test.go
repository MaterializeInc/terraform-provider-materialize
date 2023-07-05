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

	i, err := parseDefaultPrivilegeId("GRANT DEFAULT|TABLE|u1||||SELECT")
	if err != nil {
		t.Fatal(err)
	}
	if i.objectType != "TABLE" {
		t.Fatalf("%s: does not match expected TABLE", i.objectType)
	}
	if i.granteeId != "u1" {
		t.Fatalf("%s: does not match expected u1", i.granteeId)
	}
	if i.targetRoleId != "" {
		t.Fatalf("%s: expected to be empty string", i.targetRoleId)
	}
	if i.schemaId != "" {
		t.Fatalf("%s: expected to be empty string", i.schemaId)
	}
	if i.databaseId != "" {
		t.Fatalf("%s: expected to be empty string", i.databaseId)
	}

	k, err := parseDefaultPrivilegeId("GRANT DEFAULT|TABLE|u1|u2|u3|u4|SELECT")
	if err != nil {
		t.Fatal(err)
	}
	if k.objectType != "TABLE" {
		t.Fatalf("%s: does not match expected TABLE", k.objectType)
	}
	if k.granteeId != "u1" {
		t.Fatalf("%s: does not match expected u1", k.granteeId)
	}
	if k.targetRoleId != "u2" {
		t.Fatalf("%s: expected to be empty string", k.targetRoleId)
	}
	if k.schemaId != "u3" {
		t.Fatalf("%s: expected to be empty string", k.schemaId)
	}
	if k.databaseId != "u4" {
		t.Fatalf("%s: expected to be empty string", k.databaseId)
	}
}

func TestResourceGrantDefaultPrivilegeCreate(t *testing.T) {
	r := require.New(t)

	in := map[string]interface{}{
		"grantee_name":     "project_managers",
		"target_role_name": "developers",
		"object_type":      "SECRET",
		"privilege":        "USAGE",
		"database_name":    "database",
		"schema_name":      "schema",
	}
	d := schema.TestResourceDataRaw(t, GrantDefaultPrivilege().Schema, in)
	r.NotNil(d)

	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		// Create
		mock.ExpectExec(
			`ALTER DEFAULT PRIVILEGES FOR ROLE developers IN SCHEMA "database"."schema" GRANT USAGE ON SECRETS TO project_managers;`,
		).WillReturnResult(sqlmock.NewResult(1, 1))

		// DefaultPrivilegeId - Query grantee role
		gp := `WHERE mz_roles.name = 'project_managers'`
		testhelpers.MockRoleScan(mock, gp)

		// DefaultPrivilegeId - Query target role
		tp := `WHERE mz_roles.name = 'developers'`
		testhelpers.MockRoleScan(mock, tp)

		// DefaultPrivilegeId - Query database
		dp := `WHERE mz_databases.name = 'database'`
		testhelpers.MockDatabaseScan(mock, dp)

		// DefaultPrivilegeId - Query schema
		sp := `WHERE mz_databases.name = 'database' AND mz_schemas.name = 'schema'`
		testhelpers.MockSchemaScan(mock, sp)

		// Query Params
		qp := `
			WHERE mz_default_privileges.database_id = 'u1'
			AND mz_default_privileges.grantee = 'u1'
			AND mz_default_privileges.object_type = 'SECRET'
			AND mz_default_privileges.role_id = 'u1'
			AND mz_default_privileges.schema_id = 'u1'`
		testhelpers.MockDefaultPrivilegeScan(mock, qp)

		if err := grantDefaultPrivilegeCreate(context.TODO(), d, db); err != nil {
			t.Fatal(err)
		}

		if d.Id() != "GRANT DEFAULT|SECRET|u1|u1|u1|u1|USAGE" {
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
