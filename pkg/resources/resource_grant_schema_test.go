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

func TestResourceGrantSchemaCreate(t *testing.T) {
	r := require.New(t)

	in := map[string]interface{}{
		"role_name":     "joe",
		"privilege":     "CREATE",
		"schema_name":   "schema",
		"database_name": "database",
	}
	d := schema.TestResourceDataRaw(t, GrantSchema().Schema, in)
	r.NotNil(d)

	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		// Create
		mock.ExpectExec(
			`GRANT CREATE ON SCHEMA "database"."schema" TO joe;`,
		).WillReturnResult(sqlmock.NewResult(1, 1))

		// Query Role Id
		rp := `WHERE mz_roles.name = 'joe'`
		testhelpers.MockRoleScan(mock, rp)

		// Query Grant Id
		gp := `WHERE mz_databases.name = 'database' AND mz_schemas.name = 'schema'`
		testhelpers.MockSchemaScan(mock, gp)

		// Query Params
		pp := `WHERE mz_schemas.id = 'u1'`
		testhelpers.MockSchemaScan(mock, pp)

		if err := grantSchemaCreate(context.TODO(), d, db); err != nil {
			t.Fatal(err)
		}

		if d.Id() != "GRANT|SCHEMA|u1|u1|CREATE" {
			t.Fatalf("unexpected id of %s", d.Id())
		}
	})
}

func TestResourceGrantSchemaDelete(t *testing.T) {
	r := require.New(t)

	in := map[string]interface{}{
		"role_name":     "joe",
		"privilege":     "CREATE",
		"schema_name":   "schema",
		"database_name": "database",
	}
	d := schema.TestResourceDataRaw(t, GrantSchema().Schema, in)
	r.NotNil(d)

	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		mock.ExpectExec(`REVOKE CREATE ON SCHEMA "database"."schema" FROM joe;`).WillReturnResult(sqlmock.NewResult(1, 1))

		if err := grantSchemaDelete(context.TODO(), d, db); err != nil {
			t.Fatal(err)
		}
	})
}
