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

func TestResourceGrantSystemPrivilegeCreate(t *testing.T) {
	r := require.New(t)

	in := map[string]interface{}{
		"role_name": "role",
		"privilege": "CREATEDB",
	}
	d := schema.TestResourceDataRaw(t, GrantSystem().Schema, in)
	r.NotNil(d)

	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		// Create
		mock.ExpectExec(
			`GRANT CREATEDB ON SYSTEM TO role;`,
		).WillReturnResult(sqlmock.NewResult(1, 1))

		// SystemPrivilegeId - Query role
		gp := `WHERE mz_roles.name = 'role'`
		testhelpers.MockRoleScan(mock, gp)

		// Query Params
		testhelpers.MockSystemGrantScan(mock)

		if err := grantSystemCreate(context.TODO(), d, db); err != nil {
			t.Fatal(err)
		}

		if d.Id() != "GRANT SYSTEM|u1|CREATEDB" {
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
	d := schema.TestResourceDataRaw(t, GrantSystem().Schema, in)
	r.NotNil(d)

	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		mock.ExpectExec(`REVOKE CREATEDB ON SYSTEM FROM role;`).WillReturnResult(sqlmock.NewResult(1, 1))

		if err := grantSystemDelete(context.TODO(), d, db); err != nil {
			t.Fatal(err)
		}
	})
}
