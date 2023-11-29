package resources

import (
	"context"
	"testing"

	"github.com/MaterializeInc/terraform-provider-materialize/pkg/testhelpers"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/require"
)

var inRole = map[string]interface{}{
	"name":    "role",
	"inherit": true,
}

func TestResourceRoleCreate(t *testing.T) {
	r := require.New(t)
	d := schema.TestResourceDataRaw(t, Role().Schema, inRole)
	r.NotNil(d)

	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
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

func TestResourceRoleDelete(t *testing.T) {
	r := require.New(t)

	d := schema.TestResourceDataRaw(t, Role().Schema, inRole)
	r.NotNil(d)

	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		mock.ExpectExec(`DROP ROLE "role";`).WillReturnResult(sqlmock.NewResult(1, 1))

		if err := roleDelete(context.TODO(), d, db); err != nil {
			t.Fatal(err)
		}
	})
}
