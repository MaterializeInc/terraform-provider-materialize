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

var inOwnership = map[string]interface{}{
	"object":      []interface{}{map[string]interface{}{"name": "item", "schema_name": "public", "database_name": "database"}},
	"object_type": "TABLE",
	"role_name":   "my_role",
}

// func TestResourceOwnershipCreate(t *testing.T) {
// 	r := require.New(t)
// 	d := schema.TestResourceDataRaw(t, Ownership().Schema, inOwnership)
// 	r.NotNil(d)

// 	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
// 		// Create
// 		mock.ExpectExec(
// 			`ALTER TABLE "database"."schema"."table" OWNER TO my_role;`,
// 		).WillReturnResult(sqlmock.NewResult(1, 1))

// 		// Query Id
// 		ir := mock.NewRows([]string{"id"}).
// 			AddRow("u1")
// 		mock.ExpectQuery(`
// 			SELECT o.id
// 			FROM mz_tables o
// 			JOIN mz_schemas
// 				ON o.schema_id = mz_schemas.id
// 			JOIN mz_databases
// 				ON mz_schemas.database_id = mz_databases.id
// 			WHERE o.name = 'table'
// 			AND mz_databases.name = 'database'
// 			AND mz_schemas.name = 'schema'
// 		`).WillReturnRows(ir)

// 		// Query Params
// 		ip := sqlmock.NewRows([]string{"owner_id", "name"}).AddRow("u1", "my_role")
// 		mock.ExpectQuery(`
// 			SELECT
// 				o.owner_id,
// 				r.name
// 			FROM mz_tables o
// 			JOIN mz_roles r
// 				ON o.owner_id = r.id
// 			WHERE o.id = 'u1'
// 		`).WillReturnRows(ip)

// 		if err := ownershipCreate(context.TODO(), d, db); err != nil {
// 			t.Fatal(err)
// 		}
// 	})
// }

// func TestResourceOwnershipUpdate(t *testing.T) {
// 	r := require.New(t)
// 	d := schema.TestResourceDataRaw(t, Ownership().Schema, inOwnership)

// 	// Set current state
// 	d.SetId("ownership|table|u1")
// 	d.Set("role_name", "my_old_role")
// 	d.Set("object", []interface{}{map[string]interface{}{"name": "item", "schema_name": "public", "database_name": "database"}})
// 	r.NotNil(d)

// 	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
// 		mock.ExpectExec(
// 			`ALTER TABLE "database"."schema"."table" OWNER TO my_role;`,
// 		).WillReturnResult(sqlmock.NewResult(1, 1))

// 		// Query Params
// 		ip := sqlmock.NewRows([]string{"owner_id", "name"}).AddRow("u1", "my_role")
// 		mock.ExpectQuery(`
// 			SELECT
// 				o.owner_id,
// 				r.name
// 			FROM mz_tables o
// 			JOIN mz_roles r
// 				ON o.owner_id = r.id
// 			WHERE o.id = 'u1'
// 		`).WillReturnRows(ip)

// 		if err := ownershipCreate(context.TODO(), d, db); err != nil {
// 			t.Fatal(err)
// 		}
// 	})
// }

func TestResourceOwnershipDelete(t *testing.T) {
	r := require.New(t)

	d := schema.TestResourceDataRaw(t, Ownership().Schema, inOwnership)
	r.NotNil(d)

	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		if err := ownershipDelete(context.TODO(), d, db); err != nil {
			t.Fatal(err)
		}
	})

	if d.Id() != "" {
		t.Errorf("State id not set to empty string")
	}
}
