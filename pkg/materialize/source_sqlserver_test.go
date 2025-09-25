package materialize

import (
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/MaterializeInc/terraform-provider-materialize/pkg/testhelpers"
	"github.com/jmoiron/sqlx"
)

var sourceSQLServer = MaterializeObject{Name: "source", SchemaName: "schema", DatabaseName: "database"}
var tableInputSQLServer = []TableStruct{
	{UpstreamName: "table_1", UpstreamSchemaName: "dbo"},
	{UpstreamName: "table_2", UpstreamSchemaName: "dbo", Name: "table_alias"},
}

func TestSourceSQLServerAllTablesCreate(t *testing.T) {
	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		mock.ExpectExec(
			`CREATE SOURCE "database"."schema"."source" FROM SQL SERVER CONNECTION "database"."schema"."sqlserver_connection" FOR ALL TABLES;`,
		).WillReturnResult(sqlmock.NewResult(1, 1))

		b := NewSourceSQLServerBuilder(db, sourceSQLServer)
		b.SQLServerConnection(IdentifierSchemaStruct{Name: "sqlserver_connection", SchemaName: "schema", DatabaseName: "database"})

		if err := b.Create(); err != nil {
			t.Fatal(err)
		}
	})
}

func TestSourceSQLServerSpecificTablesCreate(t *testing.T) {
	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		mock.ExpectExec(
			`CREATE SOURCE "database"."schema"."source" FROM SQL SERVER CONNECTION "database"."schema"."sqlserver_connection" FOR TABLES \("dbo"."table_1" AS "database"."schema"."s1_table_1", "custom"."table_2" AS "database"."schema"."table_alias"\) EXPOSE PROGRESS AS "database"."schema"."progress";`,
		).WillReturnResult(sqlmock.NewResult(1, 1))

		b := NewSourceSQLServerBuilder(db, sourceSQLServer)
		b.SQLServerConnection(IdentifierSchemaStruct{Name: "sqlserver_connection", SchemaName: "schema", DatabaseName: "database"})
		b.Table([]TableStruct{
			{
				UpstreamName:       "table_1",
				UpstreamSchemaName: "dbo",
				Name:               "s1_table_1",
			},
			{
				UpstreamName:       "table_2",
				UpstreamSchemaName: "custom",
				Name:               "table_alias",
			},
		})
		b.ExposeProgress(IdentifierSchemaStruct{Name: "progress", DatabaseName: "database", SchemaName: "schema"})

		if err := b.Create(); err != nil {
			t.Fatal(err)
		}
	})
}

func TestSourceSQLServerWithTextColumnsCreate(t *testing.T) {
	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		mock.ExpectExec(
			`CREATE SOURCE "database"."schema"."source" FROM SQL SERVER CONNECTION "database"."schema"."sqlserver_connection" \(TEXT COLUMNS \(xml_column, ntext_column\)\) FOR ALL TABLES;`,
		).WillReturnResult(sqlmock.NewResult(1, 1))

		b := NewSourceSQLServerBuilder(db, sourceSQLServer)
		b.SQLServerConnection(IdentifierSchemaStruct{Name: "sqlserver_connection", SchemaName: "schema", DatabaseName: "database"})
		b.TextColumns([]string{"xml_column", "ntext_column"})

		if err := b.Create(); err != nil {
			t.Fatal(err)
		}
	})
}

func TestSourceSQLServerWithExcludeColumnsCreate(t *testing.T) {
	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		mock.ExpectExec(
			`CREATE SOURCE "database"."schema"."source" FROM SQL SERVER CONNECTION "database"."schema"."sqlserver_connection" \(EXCLUDE COLUMNS \(geometry_column, geography_column\)\) FOR ALL TABLES;`,
		).WillReturnResult(sqlmock.NewResult(1, 1))

		b := NewSourceSQLServerBuilder(db, sourceSQLServer)
		b.SQLServerConnection(IdentifierSchemaStruct{Name: "sqlserver_connection", SchemaName: "schema", DatabaseName: "database"})
		b.ExcludeColumns([]string{"geometry_column", "geography_column"})

		if err := b.Create(); err != nil {
			t.Fatal(err)
		}
	})
}

func TestSourceSQLServerWithTextAndExcludeColumnsCreate(t *testing.T) {
	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		mock.ExpectExec(
			`CREATE SOURCE "database"."schema"."source" FROM SQL SERVER CONNECTION "database"."schema"."sqlserver_connection" \(TEXT COLUMNS \(xml_column, ntext_column\), EXCLUDE COLUMNS \(geometry_column, geography_column\)\) FOR ALL TABLES;`,
		).WillReturnResult(sqlmock.NewResult(1, 1))

		b := NewSourceSQLServerBuilder(db, sourceSQLServer)
		b.SQLServerConnection(IdentifierSchemaStruct{Name: "sqlserver_connection", SchemaName: "schema", DatabaseName: "database"})
		b.TextColumns([]string{"xml_column", "ntext_column"})
		b.ExcludeColumns([]string{"geometry_column", "geography_column"})

		if err := b.Create(); err != nil {
			t.Fatal(err)
		}
	})
}

func TestSourceSQLServerWithClusterCreate(t *testing.T) {
	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		mock.ExpectExec(
			`CREATE SOURCE "database"."schema"."source" IN CLUSTER "test_cluster" FROM SQL SERVER CONNECTION "database"."schema"."sqlserver_connection" FOR ALL TABLES;`,
		).WillReturnResult(sqlmock.NewResult(1, 1))

		b := NewSourceSQLServerBuilder(db, sourceSQLServer)
		b.ClusterName("test_cluster")
		b.SQLServerConnection(IdentifierSchemaStruct{Name: "sqlserver_connection", SchemaName: "schema", DatabaseName: "database"})

		if err := b.Create(); err != nil {
			t.Fatal(err)
		}
	})
}

func TestSourceSQLServerDefaultSchemaHandling(t *testing.T) {
	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		mock.ExpectExec(
			`CREATE SOURCE "database"."schema"."source" FROM SQL SERVER CONNECTION "database"."schema"."sqlserver_connection" FOR TABLES \("dbo"."users" AS "database"."schema"."users", "dbo"."orders" AS "database"."schema"."orders"\);`,
		).WillReturnResult(sqlmock.NewResult(1, 1))

		b := NewSourceSQLServerBuilder(db, sourceSQLServer)
		b.SQLServerConnection(IdentifierSchemaStruct{Name: "sqlserver_connection", SchemaName: "schema", DatabaseName: "database"})
		b.Table([]TableStruct{
			{
				UpstreamName: "users",
				// Should default to "dbo" schema
			},
			{
				UpstreamName: "orders",
				// Should default to "dbo" schema
			},
		})

		if err := b.Create(); err != nil {
			t.Fatal(err)
		}
	})
}

func TestSourceSQLServerWithSSLCreate(t *testing.T) {
	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		mock.ExpectExec(
			`CREATE SOURCE "database"."schema"."source" FROM SQL SERVER CONNECTION "database"."schema"."sqlserver_connection" \(SSL MODE 'require', SSL CERTIFICATE AUTHORITY '-----BEGIN CERTIFICATE-----'\) FOR ALL TABLES;`,
		).WillReturnResult(sqlmock.NewResult(1, 1))

		b := NewSourceSQLServerBuilder(db, sourceSQLServer)
		b.SQLServerConnection(IdentifierSchemaStruct{Name: "sqlserver_connection", SchemaName: "schema", DatabaseName: "database"})
		b.SSLMode("require")
		b.SSLCertificateAuthority(ValueSecretStruct{Text: "-----BEGIN CERTIFICATE-----"})

		if err := b.Create(); err != nil {
			t.Fatal(err)
		}
	})
}

func TestSourceSQLServerWithSSLSecretCreate(t *testing.T) {
	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		mock.ExpectExec(
			`CREATE SOURCE "database"."schema"."source" FROM SQL SERVER CONNECTION "database"."schema"."sqlserver_connection" \(SSL MODE 'verify-ca', SSL CERTIFICATE AUTHORITY SECRET "database"."schema"."ssl_ca_secret"\) FOR ALL TABLES;`,
		).WillReturnResult(sqlmock.NewResult(1, 1))

		b := NewSourceSQLServerBuilder(db, sourceSQLServer)
		b.SQLServerConnection(IdentifierSchemaStruct{Name: "sqlserver_connection", SchemaName: "schema", DatabaseName: "database"})
		b.SSLMode("verify-ca")
		b.SSLCertificateAuthority(ValueSecretStruct{Secret: IdentifierSchemaStruct{Name: "ssl_ca_secret", SchemaName: "schema", DatabaseName: "database"}})

		if err := b.Create(); err != nil {
			t.Fatal(err)
		}
	})
}

func TestSourceSQLServerWithAWSPrivateLinkCreate(t *testing.T) {
	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		mock.ExpectExec(
			`CREATE SOURCE "database"."schema"."source" FROM SQL SERVER CONNECTION "database"."schema"."sqlserver_connection" \(AWS PRIVATELINK "database"."schema"."aws_privatelink_conn"\) FOR ALL TABLES;`,
		).WillReturnResult(sqlmock.NewResult(1, 1))

		b := NewSourceSQLServerBuilder(db, sourceSQLServer)
		b.SQLServerConnection(IdentifierSchemaStruct{Name: "sqlserver_connection", SchemaName: "schema", DatabaseName: "database"})
		b.AWSPrivateLink(IdentifierSchemaStruct{Name: "aws_privatelink_conn", SchemaName: "schema", DatabaseName: "database"})

		if err := b.Create(); err != nil {
			t.Fatal(err)
		}
	})
}

func TestSourceSQLServerAddSubsource(t *testing.T) {
	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		mock.ExpectExec(
			`ALTER SOURCE "database"."schema"."source" ADD SUBSOURCE "dbo"."table_1", "dbo"."table_2" AS "database"."schema"."table_alias";`,
		).WillReturnResult(sqlmock.NewResult(1, 1))

		b := NewSource(db, sourceSQLServer)
		if err := b.AddSubsource(tableInputSQLServer, []string{}); err != nil {
			t.Fatal(err)
		}
	})
}

func TestSourceSQLServerDropSubsource(t *testing.T) {
	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		mock.ExpectExec(
			`DROP SOURCE "database"."schema"."table_1", "database"."schema"."table_alias"`,
		).WillReturnResult(sqlmock.NewResult(1, 1))

		b := NewSourceSQLServerBuilder(db, sourceSQLServer)
		if err := b.DropSubsource(tableInputSQLServer); err != nil {
			t.Fatal(err)
		}
	})
}
