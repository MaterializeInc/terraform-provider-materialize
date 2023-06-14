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
		ip := `WHERE mz_databases.name = 'database' AND mz_schemas.name = 'schema' AND mz_secrets.name = 'secret'`
		testhelpers.MockSecretScan(mock, ip)

		// Query Params
		pp := `WHERE mz_secrets.id = 'u1'`
		testhelpers.MockSecretScan(mock, pp)

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
		pp := `WHERE mz_secrets.id = 'u1'`
		testhelpers.MockSecretScan(mock, pp)

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
