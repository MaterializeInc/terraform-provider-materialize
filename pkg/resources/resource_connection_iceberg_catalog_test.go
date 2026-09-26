package resources

import (
	"context"
	"testing"

	"github.com/MaterializeInc/terraform-provider-materialize/pkg/testhelpers"
	"github.com/MaterializeInc/terraform-provider-materialize/pkg/utils"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/stretchr/testify/require"
)

var inIcebergCatalog = map[string]interface{}{
	"name":           "iceberg_conn",
	"schema_name":    "schema",
	"database_name":  "database",
	"catalog_type":   "s3tablesrest",
	"url":            "https://s3tables.us-east-1.amazonaws.com/iceberg",
	"warehouse":      "arn:aws:s3tables:us-east-1:123456789012:bucket/my-bucket",
	"aws_connection": []interface{}{map[string]interface{}{"name": "aws_conn", "schema_name": "public", "database_name": "materialize"}},
	"validate":       false,
}

func TestResourceConnectionIcebergCatalogCreate(t *testing.T) {
	r := require.New(t)
	d := schema.TestResourceDataRaw(t, ConnectionIcebergCatalog().Schema, inIcebergCatalog)
	r.NotNil(d)

	testhelpers.WithMockProviderMeta(t, func(db *utils.ProviderMeta, mock sqlmock.Sqlmock) {
		// Create
		mock.ExpectExec(
			`CREATE CONNECTION "database"."schema"."iceberg_conn" TO ICEBERG CATALOG \(CATALOG TYPE = 's3tablesrest', URL = 'https://s3tables.us-east-1.amazonaws.com/iceberg', WAREHOUSE = 'arn:aws:s3tables:us-east-1:123456789012:bucket/my-bucket', AWS CONNECTION = "materialize"."public"."aws_conn"\) WITH \(VALIDATE = false\);`,
		).WillReturnResult(sqlmock.NewResult(1, 1))

		// Query Id
		ip := `WHERE mz_connections.name = 'iceberg_conn' AND mz_databases.name = 'database' AND mz_schemas.name = 'schema'`
		testhelpers.MockConnectionScan(mock, ip)

		// Query Params (uses generic connection scan)
		pp := `WHERE mz_connections.id = 'u1'`
		testhelpers.MockConnectionScan(mock, pp)

		if err := connectionIcebergCatalogCreate(context.TODO(), d, db); err != nil {
			t.Fatal(err)
		}
	})
}

func TestResourceConnectionIcebergCatalogCreateRest(t *testing.T) {
	r := require.New(t)
	in := map[string]interface{}{
		"name":          "iceberg_conn",
		"schema_name":   "schema",
		"database_name": "database",
		"catalog_type":  "rest",
		"url":           "https://dbc.cloud.databricks.com/api/2.1/unity-catalog/iceberg-rest",
		"warehouse":     "main",
		"credential": []interface{}{map[string]interface{}{
			"secret": []interface{}{map[string]interface{}{"name": "databricks_oauth", "schema_name": "public", "database_name": "materialize"}},
		}},
		"oauth2_server_url": "https://dbc.cloud.databricks.com/oidc/v1/token",
		"scope":             "all-apis",
		"access_delegation": "vended-credentials",
		"validate":          false,
	}
	d := schema.TestResourceDataRaw(t, ConnectionIcebergCatalog().Schema, in)
	r.NotNil(d)

	testhelpers.WithMockProviderMeta(t, func(db *utils.ProviderMeta, mock sqlmock.Sqlmock) {
		// Create
		mock.ExpectExec(
			`CREATE CONNECTION "database"."schema"."iceberg_conn" TO ICEBERG CATALOG \(CATALOG TYPE = 'rest', URL = 'https://dbc.cloud.databricks.com/api/2.1/unity-catalog/iceberg-rest', WAREHOUSE = 'main', CREDENTIAL = SECRET "materialize"."public"."databricks_oauth", OAUTH2 SERVER URL = 'https://dbc.cloud.databricks.com/oidc/v1/token', SCOPE = 'all-apis', ACCESS DELEGATION = 'vended-credentials'\) WITH \(VALIDATE = false\);`,
		).WillReturnResult(sqlmock.NewResult(1, 1))

		// Query Id
		ip := `WHERE mz_connections.name = 'iceberg_conn' AND mz_databases.name = 'database' AND mz_schemas.name = 'schema'`
		testhelpers.MockConnectionScan(mock, ip)

		// Query Params
		pp := `WHERE mz_connections.id = 'u1'`
		testhelpers.MockConnectionScan(mock, pp)

		if err := connectionIcebergCatalogCreate(context.TODO(), d, db); err != nil {
			t.Fatal(err)
		}
	})
}

func TestResourceConnectionIcebergCatalogRead(t *testing.T) {
	r := require.New(t)
	d := schema.TestResourceDataRaw(t, ConnectionIcebergCatalog().Schema, inIcebergCatalog)
	r.NotNil(d)

	// Set id before read
	d.SetId("u1")

	testhelpers.WithMockProviderMeta(t, func(db *utils.ProviderMeta, mock sqlmock.Sqlmock) {
		// Query Params (uses generic connection scan)
		pp := `WHERE mz_connections.id = 'u1'`
		testhelpers.MockConnectionScan(mock, pp)

		if err := connectionRead(context.TODO(), d, db); err != nil {
			t.Fatal(err)
		}

		if d.Id() != "aws/us-east-1:u1" {
			t.Fatalf("unexpected id of %s", d.Id())
		}

		// Note: catalog_type, url, warehouse, and aws_connection are maintained
		// from Terraform state since mz_internal.mz_iceberg_catalog_connections
		// does not exist yet. We verify base connection fields from the mock.
		if d.Get("name").(string) != "connection" {
			t.Fatalf("unexpected name: %s", d.Get("name").(string))
		}
	})
}

func TestResourceConnectionIcebergCatalogUpdate(t *testing.T) {
	r := require.New(t)
	d := schema.TestResourceDataRaw(t, ConnectionIcebergCatalog().Schema, inIcebergCatalog)

	// Set current state
	d.SetId("u1")
	d.Set("name", "old_conn")
	r.NotNil(d)

	testhelpers.WithMockProviderMeta(t, func(db *utils.ProviderMeta, mock sqlmock.Sqlmock) {
		// TODO: Only name rename is supported via ALTER; catalog_type, url, warehouse,
		// and aws_connection are ForceNew and will recreate the resource.
		// Error: "storage error: cannot be altered in the requested way (SQLSTATE XX000)"
		// Once Materialize supports ALTER for these properties, add tests for in-place updates.
		mock.ExpectExec(`ALTER CONNECTION "database"."schema"."" RENAME TO "iceberg_conn";`).WillReturnResult(sqlmock.NewResult(1, 1))

		// Query Params (uses generic connection scan)
		pp := `WHERE mz_connections.id = 'u1'`
		testhelpers.MockConnectionScan(mock, pp)

		if err := connectionIcebergCatalogUpdate(context.TODO(), d, db); err != nil {
			t.Fatal(err)
		}
	})
}

func TestResourceConnectionIcebergCatalogOptionsPlanCheck(t *testing.T) {
	r := require.New(t)
	res := ConnectionIcebergCatalog()
	aws := []interface{}{map[string]interface{}{"name": "aws_conn"}}
	cred := []interface{}{map[string]interface{}{"secret": []interface{}{map[string]interface{}{"name": "oauth"}}}}
	plan := func(extra map[string]interface{}) error {
		cfg := map[string]interface{}{"name": "iceberg_conn", "url": "https://catalog.example.com/iceberg"}
		for k, v := range extra {
			cfg[k] = v
		}
		_, err := res.Diff(context.TODO(), nil, terraform.NewResourceConfigRaw(cfg), nil)
		return err
	}

	wh := "arn:aws:s3tables:us-east-1:123456789012:bucket/b"
	r.NoError(plan(map[string]interface{}{"catalog_type": "s3tablesrest", "warehouse": wh, "aws_connection": aws}))
	r.NoError(plan(map[string]interface{}{"catalog_type": "rest", "credential": cred, "access_delegation": "vended-credentials"}))
	r.ErrorContains(plan(map[string]interface{}{"catalog_type": "s3tablesrest"}), "aws_connection is required")
	r.ErrorContains(plan(map[string]interface{}{"catalog_type": "s3tablesrest", "aws_connection": aws}), "warehouse is required")
	r.ErrorContains(plan(map[string]interface{}{"catalog_type": "s3tablesrest", "warehouse": wh, "aws_connection": aws, "access_delegation": "vended-credentials"}), "access_delegation is not supported")
	r.ErrorContains(plan(map[string]interface{}{"catalog_type": "s3tablesrest", "warehouse": wh, "aws_connection": aws, "oauth2_server_url": "https://x/token"}), "oauth2_server_url is not supported")
	r.ErrorContains(plan(map[string]interface{}{"catalog_type": "s3tablesrest", "warehouse": wh, "aws_connection": aws, "scope": "all"}), "scope is not supported")
	r.ErrorContains(plan(map[string]interface{}{"catalog_type": "s3tablesrest", "warehouse": wh, "aws_connection": aws, "credential": cred}), "credential is not supported")
	// an empty block used to pass the plan and panic on apply
	r.ErrorContains(plan(map[string]interface{}{"catalog_type": "rest", "credential": []interface{}{map[string]interface{}{}}}), "credential must set text or secret")
	r.ErrorContains(plan(map[string]interface{}{"catalog_type": "rest", "credential": cred, "aws_connection": aws}), "aws_connection is not supported")
	r.ErrorContains(plan(map[string]interface{}{"catalog_type": "rest"}), "credential is required")
}

// Materialize cannot alter the credential, so a new text value has to replace
// the connection. Renaming the secret it references must not: Materialize
// keeps the secret by id, and a replacement would fail while sinks depend on
// the connection.
func TestResourceConnectionIcebergCatalogCredentialChangeReplaces(t *testing.T) {
	r := require.New(t)
	res := ConnectionIcebergCatalog()
	base := map[string]string{
		"name": "iceberg_conn", "schema_name": "public", "database_name": "materialize",
		"catalog_type": "rest", "url": "https://catalog.example.com/iceberg", "validate": "true",
	}
	withState := func(extra map[string]string) *terraform.InstanceState {
		a := map[string]string{}
		for k, v := range base {
			a[k] = v
		}
		for k, v := range extra {
			a[k] = v
		}
		return &terraform.InstanceState{ID: "u1", Attributes: a}
	}
	cfg := func(cred map[string]interface{}) *terraform.ResourceConfig {
		return terraform.NewResourceConfigRaw(map[string]interface{}{
			"name": "iceberg_conn", "catalog_type": "rest", "url": "https://catalog.example.com/iceberg",
			"credential": []interface{}{cred},
		})
	}

	text := withState(map[string]string{"credential.#": "1", "credential.0.text": "id:old", "credential.0.secret.#": "0"})
	diff, err := res.Diff(context.TODO(), text, cfg(map[string]interface{}{"text": "id:new"}), nil)
	r.NoError(err)
	r.True(diff != nil && diff.RequiresNew(), "rotating the text credential must replace")

	secret := withState(map[string]string{
		"credential.#": "1", "credential.0.text": "", "credential.0.secret.#": "1",
		"credential.0.secret.0.name": "old_secret", "credential.0.secret.0.schema_name": "public", "credential.0.secret.0.database_name": "materialize",
	})
	diff, err = res.Diff(context.TODO(), secret, cfg(map[string]interface{}{"secret": []interface{}{map[string]interface{}{"name": "renamed_secret"}}}), nil)
	r.NoError(err)
	r.False(diff != nil && diff.RequiresNew(), "renaming the referenced secret must not replace the connection")
}
