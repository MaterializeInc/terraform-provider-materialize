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
		ip := `WHERE mz_databases.name = 'database' AND mz_schemas.name = 'schema'`
		testhelpers.MockSchemaScan(mock, ip)

		// Query Params
		pp := `WHERE mz_schemas.id = 'u1'`
		testhelpers.MockSchemaScan(mock, pp)

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
