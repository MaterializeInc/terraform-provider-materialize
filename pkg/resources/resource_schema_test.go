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

func TestResourceSchemaCreate(t *testing.T) {
	r := require.New(t)

	in := map[string]interface{}{
		"name":          "schema",
		"database_name": "database",
	}
	d := schema.TestResourceDataRaw(t, Schema().Schema, in)
	r.NotNil(d)

	testhelpers.WithMockProviderMeta(t, func(db *utils.ProviderMeta, mock sqlmock.Sqlmock) {
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

// Confirm id is updated with region for 0.4.0
func TestResourceSchemaReadIdMigration(t *testing.T) {
	r := require.New(t)

	in := map[string]interface{}{
		"name": "schema",
	}
	d := schema.TestResourceDataRaw(t, Schema().Schema, in)
	r.NotNil(d)

	// Set id before migration
	d.SetId("u1")

	testhelpers.WithMockProviderMeta(t, func(db *utils.ProviderMeta, mock sqlmock.Sqlmock) {
		// Query Params
		pp := `WHERE mz_schemas.id = 'u1'`
		testhelpers.MockSchemaScan(mock, pp)

		if err := schemaRead(context.TODO(), d, db); err != nil {
			t.Fatal(err)
		}

		if d.Id() != "aws/us-east-1:u1" {
			t.Fatalf("unexpected id of %s", d.Id())
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

	testhelpers.WithMockProviderMeta(t, func(db *utils.ProviderMeta, mock sqlmock.Sqlmock) {
		mock.ExpectExec(`DROP SCHEMA "database"."schema";`).WillReturnResult(sqlmock.NewResult(1, 1))

		if err := schemaDelete(context.TODO(), d, db); err != nil {
			t.Fatal(err)
		}
	})
}
