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

func TestResourceGrantViewCreate(t *testing.T) {
	r := require.New(t)

	in := map[string]interface{}{
		"role_name":     "joe",
		"privilege":     "SELECT",
		"view_name":     "view",
		"schema_name":   "schema",
		"database_name": "database",
	}
	d := schema.TestResourceDataRaw(t, GrantView().Schema, in)
	r.NotNil(d)

	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		// Create
		mock.ExpectExec(
			`GRANT SELECT ON TABLE "database"."schema"."view" TO "joe";`,
		).WillReturnResult(sqlmock.NewResult(1, 1))

		// Query Role Id
		rp := `WHERE mz_roles.name = 'joe'`
		testhelpers.MockRoleScan(mock, rp)

		// Query Grant Id
		gp := `WHERE mz_databases.name = 'database' AND mz_schemas.name = 'schema' AND mz_views.name = 'view'`
		testhelpers.MockViewScan(mock, gp)

		// Query Params
		pp := `WHERE mz_views.id = 'u1'`
		testhelpers.MockViewScan(mock, pp)

		if err := grantViewCreate(context.TODO(), d, db); err != nil {
			t.Fatal(err)
		}

		if d.Id() != "GRANT|VIEW|u1|u1|SELECT" {
			t.Fatalf("unexpected id of %s", d.Id())
		}
	})
}

func TestResourceGrantViewDelete(t *testing.T) {
	r := require.New(t)

	in := map[string]interface{}{
		"role_name":     "joe",
		"privilege":     "SELECT",
		"view_name":     "view",
		"schema_name":   "schema",
		"database_name": "database",
	}
	d := schema.TestResourceDataRaw(t, GrantView().Schema, in)
	r.NotNil(d)

	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		mock.ExpectExec(`REVOKE SELECT ON TABLE "database"."schema"."view" FROM "joe";`).WillReturnResult(sqlmock.NewResult(1, 1))

		if err := grantViewDelete(context.TODO(), d, db); err != nil {
			t.Fatal(err)
		}
	})
}
