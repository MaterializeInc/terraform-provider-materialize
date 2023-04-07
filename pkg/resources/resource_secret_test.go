package resources

import (
	"context"
	"testing"

	"github.com/MaterializeInc/terraform-materialize-provider/pkg/testhelpers"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/require"
)

func TestResourceSecretCreate(t *testing.T) {
	r := require.New(t)

	in := map[string]interface{}{
		"name":          "secret",
		"schema_name":   "schema",
		"database_name": "database",
		"value":         "value",
	}
	d := schema.TestResourceDataRaw(t, Secret().Schema, in)
	r.NotNil(d)

	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		// Create
		mock.ExpectExec(
			`CREATE SECRET "database"."schema"."secret" AS 'value';`,
		).WillReturnResult(sqlmock.NewResult(1, 1))

		// Query Id
		ir := mock.NewRows([]string{"id"}).
			AddRow("u1")
		mock.ExpectQuery(`
			SELECT mz_secrets.id
			FROM mz_secrets
			JOIN mz_schemas
				ON mz_secrets.schema_id = mz_schemas.id
			JOIN mz_databases
				ON mz_schemas.database_id = mz_databases.id
			WHERE mz_secrets.name = 'secret'
			AND mz_schemas.name = 'schema'
			AND mz_databases.name = 'database';`).WillReturnRows(ir)

		// Query Params
		ip := sqlmock.NewRows([]string{"name", "schema_name", "database_name"}).
			AddRow("secret", "schema", "database")
		mock.ExpectQuery(`
			SELECT
				mz_secrets.name AS name,
				mz_schemas.name AS schema_name,
				mz_databases.name AS database_name
			FROM mz_secrets
			JOIN mz_schemas
				ON mz_secrets.schema_id = mz_schemas.id
			JOIN mz_databases
				ON mz_schemas.database_id = mz_databases.id
			WHERE mz_secrets.id = 'u1';`).WillReturnRows(ip)

		if err := secretCreate(context.TODO(), d, db); err != nil {
			t.Fatal(err)
		}
	})

}

func TestResourceSecretDelete(t *testing.T) {
	r := require.New(t)

	in := map[string]interface{}{
		"name":          "secret",
		"schema_name":   "schema",
		"database_name": "database",
		"value":         "value",
	}
	d := schema.TestResourceDataRaw(t, Secret().Schema, in)
	r.NotNil(d)

	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		mock.ExpectExec(`DROP SECRET "database"."schema"."secret";`).WillReturnResult(sqlmock.NewResult(1, 1))

		if err := secretDelete(context.TODO(), d, db); err != nil {
			t.Fatal(err)
		}
	})

}
