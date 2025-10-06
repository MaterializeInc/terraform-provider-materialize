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

var inRole = map[string]interface{}{
	"name":    "role",
	"inherit": true,
}

var inRoleWithLogin = map[string]interface{}{
	"name":    "role",
	"inherit": true,
	"login":   true,
}

var inRoleWithPasswordAndLogin = map[string]interface{}{
	"name":     "role",
	"inherit":  true,
	"password": "password123",
	"login":    true,
}

var inRoleWithPasswordNoLogin = map[string]interface{}{
	"name":     "role",
	"inherit":  true,
	"password": "password123",
	"login":    false,
}

func TestResourceRoleCreate(t *testing.T) {
	r := require.New(t)
	d := schema.TestResourceDataRaw(t, Role().Schema, inRole)
	r.NotNil(d)

	testhelpers.WithMockProviderMeta(t, func(db *utils.ProviderMeta, mock sqlmock.Sqlmock) {
		// Create
		mock.ExpectExec(
			`CREATE ROLE "role" INHERIT;`,
		).WillReturnResult(sqlmock.NewResult(1, 1))

		// Query Id
		ip := `WHERE mz_roles.name = 'role'`
		testhelpers.MockRoleScan(mock, ip)

		// Query Params
		pp := `WHERE mz_roles.id = 'u1'`
		testhelpers.MockRoleScan(mock, pp)

		if err := roleCreate(context.TODO(), d, db); err != nil {
			t.Fatal(err)
		}
	})
}

func TestResourceRoleCreateWithLogin(t *testing.T) {
	r := require.New(t)
	d := schema.TestResourceDataRaw(t, Role().Schema, inRoleWithLogin)
	r.NotNil(d)

	testhelpers.WithMockProviderMeta(t, func(db *utils.ProviderMeta, mock sqlmock.Sqlmock) {
		// Create
		mock.ExpectExec(
			`CREATE ROLE "role" INHERIT WITH LOGIN;`,
		).WillReturnResult(sqlmock.NewResult(1, 1))

		// Query Id
		ip := `WHERE mz_roles.name = 'role'`
		testhelpers.MockRoleScan(mock, ip)

		// Query Params
		pp := `WHERE mz_roles.id = 'u1'`
		testhelpers.MockRoleScan(mock, pp)

		if err := roleCreate(context.TODO(), d, db); err != nil {
			t.Fatal(err)
		}
	})
}

func TestResourceRoleCreateWithPasswordAndLogin(t *testing.T) {
	r := require.New(t)
	d := schema.TestResourceDataRaw(t, Role().Schema, inRoleWithPasswordAndLogin)
	r.NotNil(d)

	testhelpers.WithMockProviderMeta(t, func(db *utils.ProviderMeta, mock sqlmock.Sqlmock) {
		// Create
		mock.ExpectExec(
			`CREATE ROLE "role" INHERIT WITH LOGIN PASSWORD 'password123';`,
		).WillReturnResult(sqlmock.NewResult(1, 1))

		// Query Id
		ip := `WHERE mz_roles.name = 'role'`
		testhelpers.MockRoleScan(mock, ip)

		// Query Params
		pp := `WHERE mz_roles.id = 'u1'`
		testhelpers.MockRoleScan(mock, pp)

		if err := roleCreate(context.TODO(), d, db); err != nil {
			t.Fatal(err)
		}
	})
}

func TestResourceRoleCreateWithPasswordNoLogin(t *testing.T) {
	r := require.New(t)
	d := schema.TestResourceDataRaw(t, Role().Schema, inRoleWithPasswordNoLogin)
	r.NotNil(d)

	testhelpers.WithMockProviderMeta(t, func(db *utils.ProviderMeta, mock sqlmock.Sqlmock) {
		// Create
		mock.ExpectExec(
			`CREATE ROLE "role" INHERIT WITH PASSWORD 'password123';`,
		).WillReturnResult(sqlmock.NewResult(1, 1))

		// Query Id
		ip := `WHERE mz_roles.name = 'role'`
		testhelpers.MockRoleScan(mock, ip)

		// Query Params
		pp := `WHERE mz_roles.id = 'u1'`
		testhelpers.MockRoleScan(mock, pp)

		if err := roleCreate(context.TODO(), d, db); err != nil {
			t.Fatal(err)
		}
	})
}

// Confirm id is updated with region for 0.4.0
func TestResourceRoleReadIdMigration(t *testing.T) {
	r := require.New(t)
	d := schema.TestResourceDataRaw(t, Role().Schema, inRole)
	r.NotNil(d)

	// Set id before migration
	d.SetId("u1")

	testhelpers.WithMockProviderMeta(t, func(db *utils.ProviderMeta, mock sqlmock.Sqlmock) {
		// Query Params
		pp := `WHERE mz_roles.id = 'u1'`
		testhelpers.MockRoleScan(mock, pp)

		if err := roleRead(context.TODO(), d, db); err != nil {
			t.Fatal(err)
		}

		if d.Id() != "aws/us-east-1:u1" {
			t.Fatalf("unexpected id of %s", d.Id())
		}
	})
}

func TestResourceRoleDelete(t *testing.T) {
	r := require.New(t)

	d := schema.TestResourceDataRaw(t, Role().Schema, inRole)
	r.NotNil(d)

	testhelpers.WithMockProviderMeta(t, func(db *utils.ProviderMeta, mock sqlmock.Sqlmock) {
		mock.ExpectExec(`DROP ROLE "role";`).WillReturnResult(sqlmock.NewResult(1, 1))

		if err := roleDelete(context.TODO(), d, db); err != nil {
			t.Fatal(err)
		}
	})
}
