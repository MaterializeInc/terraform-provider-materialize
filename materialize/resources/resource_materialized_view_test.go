package resources

import (
	"context"
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/require"
)

func TestResourceMaterializedViewCreate(t *testing.T) {
	r := require.New(t)

	in := map[string]interface{}{
		"name":          "materialized_view",
		"schema_name":   "schema",
		"database_name": "database",
		"select_stmt":   "SELECT 1 FROM 1",
	}
	d := schema.TestResourceDataRaw(t, MaterializedView().Schema, in)
	r.NotNil(d)

	WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		// Create
		mock.ExpectExec(`CREATE MATERIALIZED VIEW "database"."schema"."materialized_view" AS SELECT 1 FROM 1;`).WillReturnResult(sqlmock.NewResult(1, 1))

		// Query Id
		ir := mock.NewRows([]string{"id"}).AddRow("u1")
		mock.ExpectQuery(`
			SELECT mz_materialized_views.id
			FROM mz_materialized_views
			JOIN mz_schemas
				ON mz_materialized_views.schema_id = mz_schemas.id
			JOIN mz_databases
				ON mz_schemas.database_id = mz_databases.id
			WHERE mz_materialized_views.name = 'materialized_view'
			AND mz_schemas.name = 'schema'
			AND mz_databases.name = 'database';
		`).WillReturnRows(ir)

		// Query Params
		ip := sqlmock.NewRows([]string{"name", "schema", "database"}).AddRow("materialized_view", "schema", "database")
		mock.ExpectQuery(`
			SELECT
				mz_materialized_views.name,
				mz_schemas.name,
				mz_databases.name
			FROM mz_materialized_views
			JOIN mz_schemas
				ON mz_materialized_views.schema_id = mz_schemas.id
			JOIN mz_databases
				ON mz_schemas.database_id = mz_databases.id
			WHERE mz_materialized_views.id = 'u1';
		`).WillReturnRows(ip)

		if err := materializedViewCreate(context.TODO(), d, db); err != nil {
			t.Fatal(err)
		}
	})

}

func TestResourceMaterializedViewDelete(t *testing.T) {
	r := require.New(t)

	in := map[string]interface{}{
		"name":          "materialized_view",
		"schema_name":   "schema",
		"database_name": "database",
	}
	d := schema.TestResourceDataRaw(t, MaterializedView().Schema, in)
	r.NotNil(d)

	WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		mock.ExpectExec(`DROP MATERIALIZED VIEW "database"."schema"."materialized_view";`).WillReturnResult(sqlmock.NewResult(1, 1))

		if err := materializedViewDelete(context.TODO(), d, db); err != nil {
			t.Fatal(err)
		}
	})

}

func TestMaterializedViewCreateQuery(t *testing.T) {
	r := require.New(t)
	b := newMaterializedViewBuilder("materialized_view", "schema", "database")
	b.SelectStmt("SELECT 1 FROM t1")
	r.Equal(`CREATE MATERIALIZED VIEW "database"."schema"."materialized_view" AS SELECT 1 FROM t1;`, b.Create())
}

func TestMaterializedViewCreateQueryIfNotExist(t *testing.T) {
	r := require.New(t)
	b := newMaterializedViewBuilder("materialized_view", "schema", "database")
	b.SelectStmt("SELECT 1 FROM t1")
	r.Equal(`CREATE MATERIALIZED VIEW "database"."schema"."materialized_view" AS SELECT 1 FROM t1;`, b.Create())
}

func TestMaterializedViewRenameQuery(t *testing.T) {
	r := require.New(t)
	b := newMaterializedViewBuilder("materialized_view", "schema", "database")
	r.Equal(`ALTER MATERIALIZED VIEW "database"."schema"."materialized_view" RENAME TO "database"."schema"."new_view";`, b.Rename("new_view"))
}

func TestMaterializedViewDropQuery(t *testing.T) {
	r := require.New(t)
	b := newMaterializedViewBuilder("materialized_view", "schema", "database")
	r.Equal(`DROP MATERIALIZED VIEW "database"."schema"."materialized_view";`, b.Drop())
}
