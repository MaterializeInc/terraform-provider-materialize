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

func TestResourceGrantMaterializedViewCreate(t *testing.T) {
	r := require.New(t)

	in := map[string]interface{}{
		"role_name":              "joe",
		"privilege":              "SELECT",
		"materialized_view_name": "mview",
		"schema_name":            "schema",
		"database_name":          "database",
	}
	d := schema.TestResourceDataRaw(t, GrantMaterializedView().Schema, in)
	r.NotNil(d)

	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		// Create
		mock.ExpectExec(
			`GRANT SELECT ON TABLE "database"."schema"."mview" TO joe;`,
		).WillReturnResult(sqlmock.NewResult(1, 1))

		// Query Role Id
		rp := `WHERE mz_roles.name = 'joe'`
		testhelpers.MockRoleScan(mock, rp)

		// Query Grant Id
		gp := `WHERE mz_databases.name = 'database' AND mz_materialized_views.name = 'mview' AND mz_schemas.name = 'schema'`
		testhelpers.MockMaterailizeViewScan(mock, gp)

		// Query Params
		pp := `WHERE mz_materialized_views.id = 'u1'`
		testhelpers.MockMaterailizeViewScan(mock, pp)

		if err := grantMaterializedViewCreate(context.TODO(), d, db); err != nil {
			t.Fatal(err)
		}

		if d.Id() != "GRANT|MATERIALIZED VIEW|u1|u1|SELECT" {
			t.Fatalf("unexpected id of %s", d.Id())
		}
	})
}

func TestResourceGrantMaterializedViewDelete(t *testing.T) {
	r := require.New(t)

	in := map[string]interface{}{
		"role_name":              "joe",
		"privilege":              "SELECT",
		"materialized_view_name": "mview",
		"schema_name":            "schema",
		"database_name":          "database",
	}
	d := schema.TestResourceDataRaw(t, GrantMaterializedView().Schema, in)
	r.NotNil(d)

	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		mock.ExpectExec(`REVOKE SELECT ON TABLE "database"."schema"."mview" FROM joe;`).WillReturnResult(sqlmock.NewResult(1, 1))

		if err := grantMaterializedViewDelete(context.TODO(), d, db); err != nil {
			t.Fatal(err)
		}
	})
}
