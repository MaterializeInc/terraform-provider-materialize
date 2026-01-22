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

		if d.Get("catalog_type").(string) != "s3tablesrest" {
			t.Fatalf("unexpected catalog_type: %s", d.Get("catalog_type").(string))
		}

		if d.Get("url").(string) != "https://s3tables.us-east-1.amazonaws.com/iceberg" {
			t.Fatalf("unexpected url: %s", d.Get("url").(string))
		}

		if d.Get("warehouse").(string) != "arn:aws:s3tables:us-east-1:123456789012:bucket/my-bucket" {
			t.Fatalf("unexpected warehouse: %s", d.Get("warehouse").(string))
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
		mock.ExpectExec(`ALTER CONNECTION "database"."schema"."" RENAME TO "iceberg_conn";`).WillReturnResult(sqlmock.NewResult(1, 1))

		mock.ExpectExec(`ALTER CONNECTION "database"."schema"."old_conn" SET \(CATALOG TYPE = 's3tablesrest'\) WITH \(validate false\);`).WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectExec(`ALTER CONNECTION "database"."schema"."old_conn" SET \(URL = 'https://s3tables.us-east-1.amazonaws.com/iceberg'\) WITH \(validate false\);`).WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectExec(`ALTER CONNECTION "database"."schema"."old_conn" SET \(WAREHOUSE = 'arn:aws:s3tables:us-east-1:123456789012:bucket/my-bucket'\) WITH \(validate false\);`).WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectExec(`ALTER CONNECTION "database"."schema"."old_conn" SET \(AWS CONNECTION = "materialize"."public"."aws_conn"\) WITH \(validate false\);`).WillReturnResult(sqlmock.NewResult(1, 1))

		// Query Params
		pp := `WHERE mz_connections.id = 'u1'`
		testhelpers.MockConnectionIcebergCatalogScan(mock, pp)

		if err := connectionIcebergCatalogUpdate(context.TODO(), d, db); err != nil {
			t.Fatal(err)
		}
	})
}
