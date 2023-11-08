package resources

import (
	"context"
	"testing"

	"github.com/MaterializeInc/terraform-provider-materialize/pkg/testhelpers"
	"github.com/MaterializeInc/terraform-provider-materialize/pkg/utils"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/require"
)

var inType = map[string]interface{}{
	"name":            "type",
	"schema_name":     "schema",
	"database_name":   "database",
	"list_properties": []interface{}{map[string]interface{}{"element_type": "int4"}},
}

func TestResourceTypeCreate(t *testing.T) {
	r := require.New(t)
	d := schema.TestResourceDataRaw(t, Type().Schema, inType)
	r.NotNil(d)

	testhelpers.WithMockProviderMeta(t, func(db *utils.ProviderMeta, mock sqlmock.Sqlmock) {
		// Create
		mock.ExpectExec(`CREATE TYPE "database"."schema"."type" AS LIST \(ELEMENT TYPE = int4\);`).WillReturnResult(sqlmock.NewResult(1, 1))

		// Query Id
		ip := `WHERE mz_databases.name = 'database' AND mz_schemas.name = 'schema' AND mz_types.name = 'type'`
		testhelpers.MockTypeScan(mock, ip)

		// Query Params
		pp := `WHERE mz_types.id = 'u1'`
		testhelpers.MockTypeScan(mock, pp)

		if err := typeCreate(context.TODO(), d, db); err != nil {
			t.Fatal(err)
		}
	})
}

// Confirm id is updated with region for 0.4.0
func TestResourceTypeReadIdMigration(t *testing.T) {
	r := require.New(t)
	d := schema.TestResourceDataRaw(t, Type().Schema, inType)
	r.NotNil(d)

	// Set id before migration
	d.SetId("u1")

	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		// Query Params
		pp := `WHERE mz_types.id = 'u1'`
		testhelpers.MockTypeScan(mock, pp)

		if err := typeRead(context.TODO(), d, db); err != nil {
			t.Fatal(err)
		}

		if d.Id() != "aws/us-east-1:u1" {
			t.Fatalf("unexpected id of %s", d.Id())
		}
	})
}

func TestResourceTypeDelete(t *testing.T) {
	r := require.New(t)

	d := schema.TestResourceDataRaw(t, Type().Schema, inType)
	r.NotNil(d)

	testhelpers.WithMockProviderMeta(t, func(db *utils.ProviderMeta, mock sqlmock.Sqlmock) {
		mock.ExpectExec(`DROP TYPE "database"."schema"."type";`).WillReturnResult(sqlmock.NewResult(1, 1))

		if err := typeDelete(context.TODO(), d, db); err != nil {
			t.Fatal(err)
		}
	})
}
