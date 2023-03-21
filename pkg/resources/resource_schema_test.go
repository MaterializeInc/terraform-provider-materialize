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

func TestResourceSchemaCreate(t *testing.T) {
	r := require.New(t)

	in := map[string]interface{}{
		"name":          "schema",
		"database_name": "database",
	}
	d := schema.TestResourceDataRaw(t, Schema().Schema, in)
	r.NotNil(d)

	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		// Create
		mock.ExpectExec(
			`CREATE SCHEMA "database"."schema";`,
		).WillReturnResult(sqlmock.NewResult(1, 1))

		// Query Id
		ir := mock.NewRows([]string{"id"}).AddRow("u1")
		mock.ExpectQuery(`
			SELECT mz_schemas.id
			FROM mz_schemas JOIN mz_databases
				ON mz_schemas.database_id = mz_databases.id
			WHERE mz_schemas.name = 'schema'
			AND mz_databases.name = 'database';
		`).WillReturnRows(ir)

		// Query Params
		ip := sqlmock.NewRows([]string{"name", "database_name"}).
			AddRow("schema", "database")
		mock.ExpectQuery(`
			SELECT
				mz_schemas.name,
				mz_databases.name
			FROM mz_schemas JOIN mz_databases
				ON mz_schemas.database_id = mz_databases.id
			WHERE mz_schemas.id = 'u1';		
		`).WillReturnRows(ip)

		if err := schemaCreate(context.TODO(), d, db); err != nil {
			t.Fatal(err)
		}
	})

}

func TestResourceSchemaDelete(t *testing.T) {
	r := require.New(t)

	in := map[string]interface{}{
		"name":          "schema",
		"database_name": "database",
	}
	d := schema.TestResourceDataRaw(t, Schema().Schema, in)
	r.NotNil(d)

	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		mock.ExpectExec(`DROP SCHEMA "database"."schema";`).WillReturnResult(sqlmock.NewResult(1, 1))

		if err := schemaDelete(context.TODO(), d, db); err != nil {
			t.Fatal(err)
		}
	})

}
