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

		// Note: catalog_type, url, warehouse, and aws_connection_id are not
		// available from the database yet (mz_internal.mz_iceberg_catalog_connections
		// does not exist). We only verify base connection fields.
		if params.ConnectionName.String != "connection" {
			t.Fatalf("Expected connection_name to be connection, got %s", params.ConnectionName.String)
		}

		if params.SchemaName.String != "schema" {
			t.Fatalf("Expected schema_name to be schema, got %s", params.SchemaName.String)
		}

		if params.DatabaseName.String != "database" {
			t.Fatalf("Expected database_name to be database, got %s", params.DatabaseName.String)
		}
	})
}
