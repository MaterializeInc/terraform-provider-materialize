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

var inSecret = map[string]interface{}{
	"name":          "secret",
	"schema_name":   "schema",
	"database_name": "database",
	"value":         "value",
}

var readSecret string = `
SELECT
	mz_secrets.id,
	mz_secrets.name,
	mz_schemas.name AS schema_name,
	mz_databases.name AS database_name
FROM mz_secrets
JOIN mz_schemas
	ON mz_secrets.schema_id = mz_schemas.id
JOIN mz_databases
	ON mz_schemas.database_id = mz_databases.id
WHERE mz_secrets.id = 'u1';`

func TestResourceSecretCreate(t *testing.T) {
	r := require.New(t)
	d := schema.TestResourceDataRaw(t, Secret().Schema, inSecret)
	r.NotNil(d)

	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		// Create
		mock.ExpectExec(
			`CREATE SECRET "database"."schema"."secret" AS 'value';`,
		).WillReturnResult(sqlmock.NewResult(1, 1))

		// Query Id
		ir := mock.NewRows([]string{"id", "name", "schema_name", "database_name"}).
			AddRow("u1", "secret", "schema", "database")
		mock.ExpectQuery(`
			SELECT
				mz_secrets.id,
				mz_secrets.name,
				mz_schemas.name AS schema_name,
				mz_databases.name AS database_name
			FROM mz_secrets
			JOIN mz_schemas
				ON mz_secrets.schema_id = mz_schemas.id
			JOIN mz_databases
				ON mz_schemas.database_id = mz_databases.id
			WHERE mz_databases.name = 'database'
			AND mz_schemas.name = 'schema'
			AND mz_secrets.name = 'secret';`).WillReturnRows(ir)

		// Query Params
		ip := mock.NewRows([]string{"id", "name", "schema_name", "database_name"}).
			AddRow("u1", "secret", "schema", "database")
		mock.ExpectQuery(readSecret).WillReturnRows(ip)

		if err := secretCreate(context.TODO(), d, db); err != nil {
			t.Fatal(err)
		}
	})

}

func TestResourceSecretUpdate(t *testing.T) {
	r := require.New(t)
	d := schema.TestResourceDataRaw(t, Secret().Schema, inSecret)

	// Set current state
	d.SetId("u1")
	d.Set("name", "old_secret")
	d.Set("value", "old_value")
	r.NotNil(d)

	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		mock.ExpectExec(`ALTER SECRET "database"."schema"."" RENAME TO "secret";`).WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectExec(`ALTER SECRET "database"."schema"."old_secret" AS 'value';`).WillReturnResult(sqlmock.NewResult(1, 1))

		// Query Params
		ip := sqlmock.NewRows([]string{"name", "schema_name", "database_name"}).
			AddRow("secret", "schema", "database")
		mock.ExpectQuery(readSecret).WillReturnRows(ip)

		if err := secretUpdate(context.TODO(), d, db); err != nil {
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
