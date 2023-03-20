package resources

import (
	"context"
	"terraform-materialize/pkg/testhelpers"
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/require"
)

func TestResourceViewCreate(t *testing.T) {
	r := require.New(t)

	in := map[string]interface{}{
		"name":          "view",
		"schema_name":   "schema",
		"database_name": "database",
		"select_stmt":   "SELECT 1 FROM 1",
	}
	d := schema.TestResourceDataRaw(t, View().Schema, in)
	r.NotNil(d)

	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		// Create
		mock.ExpectExec(`CREATE VIEW "database"."schema"."view" AS SELECT 1 FROM 1;`).WillReturnResult(sqlmock.NewResult(1, 1))

		// Query Id
		ir := mock.NewRows([]string{"id"}).AddRow("u1")
		mock.ExpectQuery(`
			SELECT mz_views.id
			FROM mz_views
			JOIN mz_schemas
				ON mz_views.schema_id = mz_schemas.id
			JOIN mz_databases
				ON mz_schemas.database_id = mz_databases.id
			WHERE mz_views.name = 'view'
			AND mz_schemas.name = 'schema'
			AND mz_databases.name = 'database';
		`).WillReturnRows(ir)

		// Query Params
		ip := sqlmock.NewRows([]string{"name", "schema", "database"}).AddRow("view", "schema", "database")
		mock.ExpectQuery(`
			SELECT
				mz_views.name,
				mz_schemas.name,
				mz_databases.name
			FROM mz_views
			JOIN mz_schemas
				ON mz_views.schema_id = mz_schemas.id
			JOIN mz_databases
				ON mz_schemas.database_id = mz_databases.id
			WHERE mz_views.id = 'u1';
		`).WillReturnRows(ip)

		if err := viewCreate(context.TODO(), d, db); err != nil {
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
