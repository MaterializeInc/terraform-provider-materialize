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

var inRoleWithPasswordWo = map[string]interface{}{
	"name":                "role",
	"inherit":             true,
	"password_wo":         "ephemeral_password_value",
	"password_wo_version": 1,
}

var inRoleAdoptExisting = map[string]interface{}{
	"name":                 "role",
	"login":                true,
	"password":             "password123",
	"create_if_not_exists": true,
}

var inRoleCreateIfNotExistsNew = map[string]interface{}{
	"name":                 "role",
	"inherit":              true,
	"create_if_not_exists": true,
}

func TestResourceRoleCreate(t *testing.T) {
	r := require.New(t)
	d := schema.TestResourceDataRaw(t, Role().Schema, inRole)
	r.NotNil(d)

	testhelpers.WithMockProviderMeta(t, func(db *utils.ProviderMeta, mock sqlmock.Sqlmock) {
		// Create
		mock.ExpectExec(
			`CREATE ROLE "role" WITH INHERIT;`,
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
			`CREATE ROLE "role" WITH INHERIT LOGIN;`,
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
			`CREATE ROLE "role" WITH INHERIT LOGIN PASSWORD 'password123';`,
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
			`CREATE ROLE "role" WITH INHERIT PASSWORD 'password123';`,
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

// When create_if_not_exists is set and the role already exists (e.g. an
// SSO/OIDC role auto-provisioned on first login), roleCreate must adopt it
// instead of issuing CREATE ROLE, then converge the configured attributes via
// ALTER ROLE.
func TestResourceRoleCreateIfNotExistsAdoptsExisting(t *testing.T) {
	r := require.New(t)
	d := schema.TestResourceDataRaw(t, Role().Schema, inRoleAdoptExisting)
	r.NotNil(d)

	testhelpers.WithMockProviderMeta(t, func(db *utils.ProviderMeta, mock sqlmock.Sqlmock) {
		// Probe: role already exists
		probe := `WHERE mz_roles.name = 'role'`
		testhelpers.MockRoleScan(mock, probe)

		// Adopt: converge configured attributes via ALTER (no CREATE ROLE)
		mock.ExpectExec(`ALTER ROLE "role" WITH PASSWORD 'password123';`).
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectExec(`ALTER ROLE "role" LOGIN;`).
			WillReturnResult(sqlmock.NewResult(1, 1))

		// Final read by id
		pp := `WHERE mz_roles.id = 'u1'`
		testhelpers.MockRoleScan(mock, pp)

		if err := roleCreate(context.TODO(), d, db); err != nil {
			t.Fatal(err)
		}

		r.Equal("aws/us-east-1:u1", d.Id())
	})
}

// When create_if_not_exists is set but the role does not yet exist, roleCreate
// must fall through to the normal CREATE ROLE path.
func TestResourceRoleCreateIfNotExistsCreatesWhenAbsent(t *testing.T) {
	r := require.New(t)
	d := schema.TestResourceDataRaw(t, Role().Schema, inRoleCreateIfNotExistsNew)
	r.NotNil(d)

	testhelpers.WithMockProviderMeta(t, func(db *utils.ProviderMeta, mock sqlmock.Sqlmock) {
		// Probe: role does not exist
		probe := `WHERE mz_roles.name = 'role'`
		testhelpers.MockRoleScanNoRows(mock, probe)

		// Falls through to normal create
		mock.ExpectExec(`CREATE ROLE "role" WITH INHERIT;`).
			WillReturnResult(sqlmock.NewResult(1, 1))

		// Query Id after create
		testhelpers.MockRoleScan(mock, probe)

		// Final read by id
		pp := `WHERE mz_roles.id = 'u1'`
		testhelpers.MockRoleScan(mock, pp)

		if err := roleCreate(context.TODO(), d, db); err != nil {
			t.Fatal(err)
		}
	})
}

func TestResourceRoleCreateIfNotExistsSchema(t *testing.T) {
	r := require.New(t)

	f, ok := Role().Schema["create_if_not_exists"]
	r.True(ok, "create_if_not_exists should be set")
	r.Equal(schema.TypeBool, f.Type)
	r.True(f.Optional)
	r.False(f.Computed)
	r.Equal(false, f.Default)
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

func TestResourceRoleUpdateWithPasswordWo(t *testing.T) {
	r := require.New(t)
	role := Role().Schema

	passwordWo, ok := role["password_wo"]
	r.True(ok, "password_wo should be set")
	r.Equal(schema.TypeString, passwordWo.Type)
	r.True(passwordWo.Optional)
	r.True(passwordWo.Sensitive)
	r.True(passwordWo.WriteOnly, "password_wo should be WriteOnly")

	passwordWoVersion, ok := roleSchema["password_wo_version"]
	r.True(ok, "password_wo_version should be set")
	r.Equal(schema.TypeInt, passwordWoVersion.Type)
	r.True(passwordWoVersion.Optional)

	password, ok := role["password"]
	r.True(ok, "password should be set")
	r.Equal(schema.TypeString, password.Type)
	r.True(password.Optional)
	r.True(password.Sensitive)
	r.False(password.WriteOnly, "password should not be WriteOnly")
}

func TestResourceRoleSchema_ExactlyOneOf(t *testing.T) {
	d := schema.TestResourceDataRaw(t, Role().Schema, inRoleWithPasswordWo)
	require.NotNil(t, d)

	passwordWoField := Role().Schema["password_wo"]
	require.Contains(t, passwordWoField.RequiredWith, "password_wo_version")

	passwordWoVersionField := Role().Schema["password_wo_version"]
	require.Contains(t, passwordWoVersionField.RequiredWith, "password_wo")
}
