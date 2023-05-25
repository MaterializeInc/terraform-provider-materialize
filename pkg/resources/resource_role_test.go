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
	"name":           "role",
	"inherit":        true,
	"create_role":    true,
	"create_db":      false,
	"create_cluster": true,
}

var readRole string = `
SELECT
	id,
	name AS role_name,
	inherit,
	create_role,
	create_db,
	create_cluster
FROM mz_roles
WHERE mz_roles.id = 'u1';`

func TestResourceRoleCreate(t *testing.T) {
	r := require.New(t)
	d := schema.TestResourceDataRaw(t, Role().Schema, inRole)
	r.NotNil(d)

	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		// Create
		mock.ExpectExec(
			`CREATE ROLE "role" INHERIT CREATEROLE CREATECLUSTER;`,
		).WillReturnResult(sqlmock.NewResult(1, 1))

		// Query Id
		ir := mock.NewRows([]string{"id"}).
			AddRow("u1")
		mock.ExpectQuery(`
			SELECT
				id,
				name AS role_name,
				inherit,
				create_role,
				create_db,
				create_cluster
			FROM mz_roles
			WHERE mz_roles.name = 'role'`).WillReturnRows(ir)

		// Query Params
		ip := sqlmock.NewRows([]string{"role_name", "inherit", "create_role", "create_db", "create_cluster"}).
			AddRow("role", true, true, false, true)
		mock.ExpectQuery(readRole).WillReturnRows(ip)

		if err := roleCreate(context.TODO(), d, db); err != nil {
			t.Fatal(err)
		}
	})

}

func TestResourceRoleUpdate(t *testing.T) {
	r := require.New(t)
	d := schema.TestResourceDataRaw(t, Role().Schema, inRole)

	// Set current state
	d.SetId("u1")
	d.Set("inherit", true)
	d.Set("create_role", false)
	d.Set("create_db", false)
	d.Set("create_cluster", false)
	r.NotNil(d)

	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		mock.ExpectExec(`ALTER ROLE "role" CREATEROLE;`).WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectExec(`ALTER ROLE "role" CREATECLUSTER;`).WillReturnResult(sqlmock.NewResult(1, 1))

		// Query Params
		ip := sqlmock.NewRows([]string{"role_name", "inherit", "create_role", "create_db", "create_cluster"}).
			AddRow("role", true, true, false, true)
		mock.ExpectQuery(readRole).WillReturnRows(ip)

		if err := roleUpdate(context.TODO(), d, db); err != nil {
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
