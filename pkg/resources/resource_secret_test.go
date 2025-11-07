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

	testhelpers.WithMockProviderMeta(t, func(db *utils.ProviderMeta, mock sqlmock.Sqlmock) {

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

// Confirm id is updated with region for 0.4.0
func TestResourceSecretReadIdMigration(t *testing.T) {
	r := require.New(t)
	d := schema.TestResourceDataRaw(t, Secret().Schema, inSecret)
	r.NotNil(d)

	// Set id before migration
	d.SetId("u1")

	testhelpers.WithMockProviderMeta(t, func(db *utils.ProviderMeta, mock sqlmock.Sqlmock) {
		// Query Params
		pp := `WHERE mz_secrets.id = 'u1'`
		testhelpers.MockSecretScan(mock, pp)

		if err := secretRead(context.TODO(), d, db); err != nil {
			t.Fatal(err)
		}

		if d.Id() != "aws/us-east-1:u1" {
			t.Fatalf("unexpected id of %s", d.Id())
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

	testhelpers.WithMockProviderMeta(t, func(db *utils.ProviderMeta, mock sqlmock.Sqlmock) {
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

	testhelpers.WithMockProviderMeta(t, func(db *utils.ProviderMeta, mock sqlmock.Sqlmock) {
		mock.ExpectExec(`DROP SECRET "database"."schema"."secret";`).WillReturnResult(sqlmock.NewResult(1, 1))

		if err := secretDelete(context.TODO(), d, db); err != nil {
			t.Fatal(err)
		}
	})
}

func TestResourceSecretSchema_WriteOnly(t *testing.T) {
	r := require.New(t)

	// Test that the schema has the write-only fields
	s := Secret().Schema

	// Check value_wo field exists and is configured correctly
	valueWo, ok := s["value_wo"]
	r.True(ok, "value_wo field should exist")
	r.Equal(schema.TypeString, valueWo.Type)
	r.True(valueWo.Optional)
	r.True(valueWo.Sensitive)
	r.True(valueWo.WriteOnly, "value_wo should be WriteOnly")

	// Check value_wo_version field exists
	valueWoVersion, ok := s["value_wo_version"]
	r.True(ok, "value_wo_version field should exist")
	r.Equal(schema.TypeInt, valueWoVersion.Type)
	r.True(valueWoVersion.Optional)

	// Check value field is now optional (not required)
	value, ok := s["value"]
	r.True(ok, "value field should exist")
	r.True(value.Optional, "value should be optional")
	r.False(value.Required, "value should not be required")
}

func TestResourceSecretSchema_ExactlyOneOf(t *testing.T) {
	// Test that value and value_wo cannot both be set
	in := map[string]interface{}{
		"name":          "secret",
		"schema_name":   "schema",
		"database_name": "database",
		"value":         "regular_value",
		"value_wo":      "write_only_value",
	}

	// This should fail validation due to ExactlyOneOf constraint
	d := schema.TestResourceDataRaw(t, Secret().Schema, in)

	valueField := Secret().Schema["value"]
	require.Contains(t, valueField.ExactlyOneOf, "value")
	require.Contains(t, valueField.ExactlyOneOf, "value_wo")

	valueWoField := Secret().Schema["value_wo"]
	require.Contains(t, valueWoField.ExactlyOneOf, "value")
	require.Contains(t, valueWoField.ExactlyOneOf, "value_wo")

	// Verify RequiredWith constraint
	require.Contains(t, valueWoField.RequiredWith, "value_wo_version")
	require.Contains(t, Secret().Schema["value_wo_version"].RequiredWith, "value_wo")

	// Clean up - this validates the data was created even though it would fail in real usage
	require.NotNil(t, d)
}
