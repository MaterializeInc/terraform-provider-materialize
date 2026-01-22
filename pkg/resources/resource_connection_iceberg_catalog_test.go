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

		// Query Params
		pp := `WHERE mz_connections.id = 'u1'`
		testhelpers.MockConnectionIcebergCatalogScan(mock, pp)

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
		// Query Params
		pp := `WHERE mz_connections.id = 'u1'`
		testhelpers.MockConnectionIcebergCatalogScan(mock, pp)

		if err := connectionIcebergCatalogRead(context.TODO(), d, db); err != nil {
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

		// Query Params
		pp := `WHERE mz_connections.id = 'u1'`
		testhelpers.MockConnectionIcebergCatalogScan(mock, pp)

		if err := connectionIcebergCatalogUpdate(context.TODO(), d, db); err != nil {
			t.Fatal(err)
		}
	})
}
