package materialize

import (
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/MaterializeInc/terraform-provider-materialize/pkg/testhelpers"
	"github.com/jmoiron/sqlx"
)

func TestConnectionIcebergCatalogCreate(t *testing.T) {
	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		mock.ExpectExec(
			`CREATE CONNECTION "database"."schema"."iceberg_conn" TO ICEBERG CATALOG \(CATALOG TYPE = 's3tablesrest', URL = 'https://s3tables.us-east-1.amazonaws.com/iceberg', WAREHOUSE = 'arn:aws:s3tables:us-east-1:123456789012:bucket/my-bucket', AWS CONNECTION = "database"."schema"."aws_conn"\) WITH \(VALIDATE = false\);`,
		).WillReturnResult(sqlmock.NewResult(1, 1))

		o := MaterializeObject{Name: "iceberg_conn", SchemaName: "schema", DatabaseName: "database"}
		b := NewConnectionIcebergCatalogBuilder(db, o)
		b.CatalogType("s3tablesrest")
		b.Url("https://s3tables.us-east-1.amazonaws.com/iceberg")
		b.Warehouse("arn:aws:s3tables:us-east-1:123456789012:bucket/my-bucket")
		b.AwsConnection(IdentifierSchemaStruct{Name: "aws_conn", DatabaseName: "database", SchemaName: "schema"})
		b.Validate(false)

		if err := b.Create(); err != nil {
			t.Fatal(err)
		}
	})
}

func TestConnectionIcebergCatalogCreateWithValidation(t *testing.T) {
	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		mock.ExpectExec(
			`CREATE CONNECTION "database"."schema"."iceberg_conn" TO ICEBERG CATALOG \(CATALOG TYPE = 's3tablesrest', URL = 'https://s3tables.us-east-1.amazonaws.com/iceberg', WAREHOUSE = 'arn:aws:s3tables:us-east-1:123456789012:bucket/my-bucket', AWS CONNECTION = "database"."schema"."aws_conn"\);`,
		).WillReturnResult(sqlmock.NewResult(1, 1))

		o := MaterializeObject{Name: "iceberg_conn", SchemaName: "schema", DatabaseName: "database"}
		b := NewConnectionIcebergCatalogBuilder(db, o)
		b.CatalogType("s3tablesrest")
		b.Url("https://s3tables.us-east-1.amazonaws.com/iceberg")
		b.Warehouse("arn:aws:s3tables:us-east-1:123456789012:bucket/my-bucket")
		b.AwsConnection(IdentifierSchemaStruct{Name: "aws_conn", DatabaseName: "database", SchemaName: "schema"})
		b.Validate(true)

		if err := b.Create(); err != nil {
			t.Fatal(err)
		}
	})
}

func TestScanConnectionIcebergCatalog(t *testing.T) {
	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		// Mock the scan query response
		pp := `WHERE mz_connections.id = 'u1'`
		testhelpers.MockConnectionIcebergCatalogScan(mock, pp)

		// Test ScanConnectionIcebergCatalog
		params, err := ScanConnectionIcebergCatalog(db, "u1")
		if err != nil {
			t.Fatal(err)
		}

		if params.CatalogType.String != "s3tablesrest" {
			t.Fatalf("Expected catalog_type to be s3tablesrest, got %s", params.CatalogType.String)
		}

		if params.Url.String != "https://s3tables.us-east-1.amazonaws.com/iceberg" {
			t.Fatalf("Expected url to be https://s3tables.us-east-1.amazonaws.com/iceberg, got %s", params.Url.String)
		}

		if params.Warehouse.String != "arn:aws:s3tables:us-east-1:123456789012:bucket/my-bucket" {
			t.Fatalf("Expected warehouse to be arn:aws:s3tables:us-east-1:123456789012:bucket/my-bucket, got %s", params.Warehouse.String)
		}

		if params.AwsConnectionId.String != "u2" {
			t.Fatalf("Expected aws_connection_id to be u2, got %s", params.AwsConnectionId.String)
		}
	})
}