package materialize

import (
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/MaterializeInc/terraform-provider-materialize/pkg/testhelpers"
	"github.com/jmoiron/sqlx"
)

func TestSinkIcebergCreate(t *testing.T) {
	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		mock.ExpectExec(
			`CREATE SINK "database"."schema"."iceberg_sink" IN CLUSTER "my_cluster" FROM "database"."schema"."my_view" INTO ICEBERG CATALOG CONNECTION "database"."schema"."iceberg_catalog" \(NAMESPACE = 'my_namespace', TABLE = 'my_table'\) USING AWS CONNECTION "database"."schema"."aws_conn" KEY \(id\) MODE UPSERT WITH \(COMMIT INTERVAL = '10s'\);`,
		).WillReturnResult(sqlmock.NewResult(1, 1))

		o := MaterializeObject{Name: "iceberg_sink", SchemaName: "schema", DatabaseName: "database"}
		b := NewSinkIcebergBuilder(db, o)
		b.ClusterName("my_cluster")
		b.From(IdentifierSchemaStruct{Name: "my_view", SchemaName: "schema", DatabaseName: "database"})
		b.IcebergCatalogConnection(IdentifierSchemaStruct{Name: "iceberg_catalog", SchemaName: "schema", DatabaseName: "database"})
		b.Namespace("my_namespace")
		b.Table("my_table")
		b.AwsConnection(IdentifierSchemaStruct{Name: "aws_conn", SchemaName: "schema", DatabaseName: "database"})
		b.Key([]string{"id"})
		b.CommitInterval("10s")

		if err := b.Create(); err != nil {
			t.Fatal(err)
		}
	})
}

func TestSinkIcebergCreateWithMultipleKeys(t *testing.T) {
	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		mock.ExpectExec(
			`CREATE SINK "database"."schema"."iceberg_sink" IN CLUSTER "my_cluster" FROM "database"."schema"."my_view" INTO ICEBERG CATALOG CONNECTION "database"."schema"."iceberg_catalog" \(NAMESPACE = 'my_namespace', TABLE = 'my_table'\) USING AWS CONNECTION "database"."schema"."aws_conn" KEY \(id, tenant_id\) MODE UPSERT WITH \(COMMIT INTERVAL = '1m'\);`,
		).WillReturnResult(sqlmock.NewResult(1, 1))

		o := MaterializeObject{Name: "iceberg_sink", SchemaName: "schema", DatabaseName: "database"}
		b := NewSinkIcebergBuilder(db, o)
		b.ClusterName("my_cluster")
		b.From(IdentifierSchemaStruct{Name: "my_view", SchemaName: "schema", DatabaseName: "database"})
		b.IcebergCatalogConnection(IdentifierSchemaStruct{Name: "iceberg_catalog", SchemaName: "schema", DatabaseName: "database"})
		b.Namespace("my_namespace")
		b.Table("my_table")
		b.AwsConnection(IdentifierSchemaStruct{Name: "aws_conn", SchemaName: "schema", DatabaseName: "database"})
		b.Key([]string{"id", "tenant_id"})
		b.CommitInterval("1m")

		if err := b.Create(); err != nil {
			t.Fatal(err)
		}
	})
}

func TestSinkIcebergCreateWithKeyNotEnforced(t *testing.T) {
	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		mock.ExpectExec(
			`CREATE SINK "database"."schema"."iceberg_sink" IN CLUSTER "my_cluster" FROM "database"."schema"."my_view" INTO ICEBERG CATALOG CONNECTION "database"."schema"."iceberg_catalog" \(NAMESPACE = 'my_namespace', TABLE = 'my_table'\) USING AWS CONNECTION "database"."schema"."aws_conn" KEY \(id\) NOT ENFORCED MODE UPSERT WITH \(COMMIT INTERVAL = '30s'\);`,
		).WillReturnResult(sqlmock.NewResult(1, 1))

		o := MaterializeObject{Name: "iceberg_sink", SchemaName: "schema", DatabaseName: "database"}
		b := NewSinkIcebergBuilder(db, o)
		b.ClusterName("my_cluster")
		b.From(IdentifierSchemaStruct{Name: "my_view", SchemaName: "schema", DatabaseName: "database"})
		b.IcebergCatalogConnection(IdentifierSchemaStruct{Name: "iceberg_catalog", SchemaName: "schema", DatabaseName: "database"})
		b.Namespace("my_namespace")
		b.Table("my_table")
		b.AwsConnection(IdentifierSchemaStruct{Name: "aws_conn", SchemaName: "schema", DatabaseName: "database"})
		b.Key([]string{"id"})
		b.KeyNotEnforced(true)
		b.CommitInterval("30s")

		if err := b.Create(); err != nil {
			t.Fatal(err)
		}
	})
}

func TestSinkIcebergCreateMinimal(t *testing.T) {
	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		mock.ExpectExec(
			`CREATE SINK "database"."schema"."iceberg_sink" FROM "database"."schema"."my_view" INTO ICEBERG CATALOG CONNECTION "database"."schema"."iceberg_catalog" \(NAMESPACE = 'ns', TABLE = 'tbl'\) USING AWS CONNECTION "database"."schema"."aws_conn" KEY \(id\) MODE UPSERT WITH \(COMMIT INTERVAL = '10s'\);`,
		).WillReturnResult(sqlmock.NewResult(1, 1))

		o := MaterializeObject{Name: "iceberg_sink", SchemaName: "schema", DatabaseName: "database"}
		b := NewSinkIcebergBuilder(db, o)
		b.From(IdentifierSchemaStruct{Name: "my_view", SchemaName: "schema", DatabaseName: "database"})
		b.IcebergCatalogConnection(IdentifierSchemaStruct{Name: "iceberg_catalog", SchemaName: "schema", DatabaseName: "database"})
		b.Namespace("ns")
		b.Table("tbl")
		b.AwsConnection(IdentifierSchemaStruct{Name: "aws_conn", SchemaName: "schema", DatabaseName: "database"})
		b.Key([]string{"id"})
		b.CommitInterval("10s")

		if err := b.Create(); err != nil {
			t.Fatal(err)
		}
	})
}
