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

var inView = map[string]interface{}{
	"name":          "view",
	"schema_name":   "schema",
	"database_name": "database",
	"statement":     "SELECT 1 FROM 1",
}

func TestResourceViewCreate(t *testing.T) {
	r := require.New(t)
	d := schema.TestResourceDataRaw(t, View().Schema, inView)
	r.NotNil(d)

	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		// Create
		mock.ExpectExec(`CREATE VIEW "database"."schema"."view" AS SELECT 1 FROM 1;`).WillReturnResult(sqlmock.NewResult(1, 1))

		// Query Id
		ip := `WHERE mz_databases.name = 'database' AND mz_schemas.name = 'schema' AND mz_views.name = 'view'`
		testhelpers.MockViewScan(mock, ip)

		// Query Params
		pp := `WHERE mz_views.id = 'u1'`
		testhelpers.MockViewScan(mock, pp)

		if err := viewCreate(context.TODO(), d, db); err != nil {
			t.Fatal(err)
		}
	})
}

func TestResourceViewUpdate(t *testing.T) {
	r := require.New(t)
	d := schema.TestResourceDataRaw(t, View().Schema, inView)

	// Set current state
	d.SetId("u1")
	d.Set("name", "old_view")
	r.NotNil(d)

	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		mock.ExpectExec(`ALTER VIEW "database"."schema"."old_view" RENAME TO "database"."schema"."view";`).WillReturnResult(sqlmock.NewResult(1, 1))

		// Query Params
		pp := `WHERE mz_views.id = 'u1'`
		testhelpers.MockViewScan(mock, pp)

		if err := viewUpdate(context.TODO(), d, db); err != nil {
			t.Fatal(err)
		}
	})
}

func TestResourceViewDelete(t *testing.T) {
	r := require.New(t)

	in := map[string]interface{}{
		"name":          "view",
		"schema_name":   "schema",
		"database_name": "database",
	}
	d := schema.TestResourceDataRaw(t, View().Schema, in)
	r.NotNil(d)

	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		mock.ExpectExec(`DROP VIEW "database"."schema"."view";`).WillReturnResult(sqlmock.NewResult(1, 1))

		if err := viewDelete(context.TODO(), d, db); err != nil {
			t.Fatal(err)
		}
	})
}
