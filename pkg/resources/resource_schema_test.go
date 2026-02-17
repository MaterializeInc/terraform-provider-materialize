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

func TestResourceSchemaCreateWithIdentifyByName(t *testing.T) {
	r := require.New(t)

	in := map[string]interface{}{
		"name":             "schema",
		"database_name":    "database",
		"identify_by_name": true,
	}
	d := schema.TestResourceDataRaw(t, Schema().Schema, in)
	r.NotNil(d)

	testhelpers.WithMockProviderMeta(t, func(db *utils.ProviderMeta, mock sqlmock.Sqlmock) {
		mock.ExpectExec(`CREATE SCHEMA "database"."schema";`).WillReturnResult(sqlmock.NewResult(1, 1))

		// Create path with identify_by_name does not call SchemaId; schemaRead uses name lookup
		pp := `WHERE mz_databases.name = 'database' AND mz_schemas.name = 'schema'`
		testhelpers.MockSchemaScan(mock, pp)

		if err := schemaCreate(context.TODO(), d, db); err != nil {
			t.Fatal(err)
		}

		r.Equal("aws/us-east-1:name:database|schema", d.Id())
		r.True(d.Get("identify_by_name").(bool))
	})
}

// Confirm id is updated with region and identify_by_name (by-id path keeps 2-part format)
func TestResourceSchemaReadIdMigration(t *testing.T) {
	utils.SetDefaultRegion("aws/us-east-1")
	r := require.New(t)

	testCases := []struct {
		name           string
		identifyByName bool
		initialId      string
		expectedId     string
		mockPredicate  string
		expectName     string
		expectDatabase string
	}{
		{
			name:           "Migrate to ID-based identifier (2-part format unchanged)",
			identifyByName: false,
			initialId:      "u1",
			expectedId:     "aws/us-east-1:u1",
			mockPredicate:  `WHERE mz_schemas.id = 'u1'`,
			expectName:     "schema",
			expectDatabase: "database",
		},
		{
			name:           "Migrate to name-based identifier",
			identifyByName: true,
			initialId:      "u1",
			expectedId:     "aws/us-east-1:name:database|schema",
			mockPredicate:  `WHERE mz_schemas.id = 'u1'`,
			expectName:     "schema",
			expectDatabase: "database",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			in := map[string]interface{}{
				"name":             "schema",
				"database_name":    "database",
				"identify_by_name": tc.identifyByName,
			}
			d := schema.TestResourceDataRaw(t, Schema().Schema, in)
			r.NotNil(d)

			d.SetId(tc.initialId)

			testhelpers.WithMockProviderMeta(t, func(db *utils.ProviderMeta, mock sqlmock.Sqlmock) {
				testhelpers.MockSchemaScan(mock, tc.mockPredicate)

				if err := schemaRead(context.TODO(), d, db); err != nil {
					t.Fatal(err)
				}

				if d.Id() != tc.expectedId {
					t.Fatalf("unexpected id of %s, expected %s", d.Id(), tc.expectedId)
				}
				if name := d.Get("name").(string); name != tc.expectName {
					t.Fatalf("unexpected name %s, expected %s", name, tc.expectName)
				}
				if dbName := d.Get("database_name").(string); dbName != tc.expectDatabase {
					t.Fatalf("unexpected database_name %s, expected %s", dbName, tc.expectDatabase)
				}
			})
		})
	}
}

// Confirm read with identify_by_name preserves region:name:database|schema
func TestResourceSchemaReadIdentifyByName(t *testing.T) {
	r := require.New(t)

	in := map[string]interface{}{
		"name":             "schema",
		"database_name":    "database",
		"identify_by_name": true,
	}
	d := schema.TestResourceDataRaw(t, Schema().Schema, in)
	r.NotNil(d)
	d.SetId("aws/us-east-1:name:database|schema")

	testhelpers.WithMockProviderMeta(t, func(db *utils.ProviderMeta, mock sqlmock.Sqlmock) {
		pp := `WHERE mz_databases.name = 'database' AND mz_schemas.name = 'schema'`
		testhelpers.MockSchemaScan(mock, pp)

		if err := schemaRead(context.TODO(), d, db); err != nil {
			t.Fatal(err)
		}

		r.Equal("aws/us-east-1:name:database|schema", d.Id())
		r.True(d.Get("identify_by_name").(bool))
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
