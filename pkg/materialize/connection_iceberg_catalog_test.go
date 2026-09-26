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

func TestConnectionIcebergCatalogCreateRest(t *testing.T) {
	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		mock.ExpectExec(
			`CREATE CONNECTION "database"."schema"."iceberg_conn" TO ICEBERG CATALOG \(CATALOG TYPE = 'rest', URL = 'https://dbc.cloud.databricks.com/api/2.1/unity-catalog/iceberg-rest', WAREHOUSE = 'main', CREDENTIAL = SECRET "database"."schema"."databricks_oauth", OAUTH2 SERVER URL = 'https://dbc.cloud.databricks.com/oidc/v1/token', SCOPE = 'all-apis', ACCESS DELEGATION = 'vended-credentials'\) WITH \(VALIDATE = false\);`,
		).WillReturnResult(sqlmock.NewResult(1, 1))

		o := MaterializeObject{Name: "iceberg_conn", SchemaName: "schema", DatabaseName: "database"}
		b := NewConnectionIcebergCatalogBuilder(db, o)
		b.CatalogType("rest")
		b.Url("https://dbc.cloud.databricks.com/api/2.1/unity-catalog/iceberg-rest")
		b.Warehouse("main")
		b.Credential(ValueSecretStruct{Secret: IdentifierSchemaStruct{Name: "databricks_oauth", SchemaName: "schema", DatabaseName: "database"}})
		b.Oauth2ServerUrl("https://dbc.cloud.databricks.com/oidc/v1/token")
		b.Scope("all-apis")
		b.AccessDelegation("vended-credentials")
		b.Validate(false)

		if err := b.Create(); err != nil {
			t.Fatal(err)
		}
	})
}

func TestConnectionIcebergCatalogCreateRestTextCredential(t *testing.T) {
	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		mock.ExpectExec(
			`CREATE CONNECTION "database"."schema"."iceberg_conn" TO ICEBERG CATALOG \(CATALOG TYPE = 'rest', URL = 'https://catalog.example.com/iceberg', CREDENTIAL = 'client:secret'\);`,
		).WillReturnResult(sqlmock.NewResult(1, 1))

		o := MaterializeObject{Name: "iceberg_conn", SchemaName: "schema", DatabaseName: "database"}
		b := NewConnectionIcebergCatalogBuilder(db, o)
		b.CatalogType("rest")
		b.Url("https://catalog.example.com/iceberg")
		b.Credential(ValueSecretStruct{Text: "client:secret"})
		b.Validate(true)

		if err := b.Create(); err != nil {
			t.Fatal(err)
		}
	})
}
