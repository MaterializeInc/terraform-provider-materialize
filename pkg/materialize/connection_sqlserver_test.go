package materialize

import (
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/MaterializeInc/terraform-provider-materialize/pkg/testhelpers"
	"github.com/jmoiron/sqlx"
)

var connSQLServer = MaterializeObject{Name: "sqlserver_conn", SchemaName: "schema", DatabaseName: "database"}

func TestConnectionSQLServerCreate(t *testing.T) {
	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		mock.ExpectExec(
			`CREATE CONNECTION "database"."schema"."sqlserver_conn" TO SQL SERVER \(HOST 'sqlserver_host', PORT 1433, USER 'user', PASSWORD SECRET "database"."schema"."password", DATABASE 'testdb'\);`,
		).WillReturnResult(sqlmock.NewResult(1, 1))

		b := NewConnectionSQLServerBuilder(db, connSQLServer)
		b.SQLServerHost("sqlserver_host")
		b.SQLServerPort(1433)
		b.SQLServerUser(ValueSecretStruct{Text: "user"})
		b.SQLServerPassword(IdentifierSchemaStruct{Name: "password", SchemaName: "schema", DatabaseName: "database"})
		b.SQLServerDatabase("testdb")
		b.Validate(true)

		if err := b.Create(); err != nil {
			t.Fatal(err)
		}
	})
}

func TestConnectionSQLServerSshCreate(t *testing.T) {
	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		mock.ExpectExec(
			`CREATE CONNECTION "database"."schema"."sqlserver_conn" TO SQL SERVER \(HOST 'sqlserver_host', PORT 1433, USER 'user', PASSWORD SECRET "database"."schema"."password", SSH TUNNEL "database"."schema"."ssh_conn", DATABASE 'testdb'\);`,
		).WillReturnResult(sqlmock.NewResult(1, 1))

		b := NewConnectionSQLServerBuilder(db, connSQLServer)
		b.SQLServerHost("sqlserver_host")
		b.SQLServerPort(1433)
		b.SQLServerUser(ValueSecretStruct{Text: "user"})
		b.SQLServerPassword(IdentifierSchemaStruct{Name: "password", SchemaName: "schema", DatabaseName: "database"})
		b.SQLServerSSHTunnel(IdentifierSchemaStruct{Name: "ssh_conn", SchemaName: "schema", DatabaseName: "database"})
		b.SQLServerDatabase("testdb")
		b.Validate(true)

		if err := b.Create(); err != nil {
			t.Fatal(err)
		}
	})
}

func TestConnectionSQLServerAWSPrivateLinkCreate(t *testing.T) {
	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		mock.ExpectExec(
			`CREATE CONNECTION "database"."schema"."sqlserver_conn" TO SQL SERVER \(HOST 'sqlserver_host', PORT 1433, USER SECRET "database"."schema"."user", PASSWORD SECRET "database"."schema"."password", AWS PRIVATELINK "database"."schema"."aws_conn", DATABASE 'testdb'\);`,
		).WillReturnResult(sqlmock.NewResult(1, 1))

		b := NewConnectionSQLServerBuilder(db, connSQLServer)
		b.SQLServerHost("sqlserver_host")
		b.SQLServerPort(1433)
		b.SQLServerUser(ValueSecretStruct{Secret: IdentifierSchemaStruct{Name: "user", SchemaName: "schema", DatabaseName: "database"}})
		b.SQLServerPassword(IdentifierSchemaStruct{Name: "password", SchemaName: "schema", DatabaseName: "database"})
		b.SQLServerAWSPrivateLink(IdentifierSchemaStruct{Name: "aws_conn", SchemaName: "schema", DatabaseName: "database"})
		b.SQLServerDatabase("testdb")
		b.Validate(true)

		if err := b.Create(); err != nil {
			t.Fatal(err)
		}
	})
}

func TestConnectionSQLServerWithoutValidation(t *testing.T) {
	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		mock.ExpectExec(
			`CREATE CONNECTION "database"."schema"."sqlserver_conn" TO SQL SERVER \(HOST 'sqlserver_host', PORT 1433, USER 'user', PASSWORD SECRET "database"."schema"."password", DATABASE 'testdb'\) WITH \(VALIDATE = false\);`,
		).WillReturnResult(sqlmock.NewResult(1, 1))

		b := NewConnectionSQLServerBuilder(db, connSQLServer)
		b.SQLServerHost("sqlserver_host")
		b.SQLServerPort(1433)
		b.SQLServerUser(ValueSecretStruct{Text: "user"})
		b.SQLServerPassword(IdentifierSchemaStruct{Name: "password", SchemaName: "schema", DatabaseName: "database"})
		b.SQLServerDatabase("testdb")
		b.Validate(false)

		if err := b.Create(); err != nil {
			t.Fatal(err)
		}
	})
}

func TestConnectionSQLServerDefaultPort(t *testing.T) {
	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		mock.ExpectExec(
			`CREATE CONNECTION "database"."schema"."sqlserver_conn" TO SQL SERVER \(HOST 'sqlserver_host', PORT 0, USER 'user', PASSWORD SECRET "database"."schema"."password", DATABASE 'testdb'\);`,
		).WillReturnResult(sqlmock.NewResult(1, 1))

		b := NewConnectionSQLServerBuilder(db, connSQLServer)
		b.SQLServerHost("sqlserver_host")
		// Default port should be 0 when not set (will be corrected in future iterations)
		b.SQLServerUser(ValueSecretStruct{Text: "user"})
		b.SQLServerPassword(IdentifierSchemaStruct{Name: "password", SchemaName: "schema", DatabaseName: "database"})
		b.SQLServerDatabase("testdb")
		b.Validate(true)

		if err := b.Create(); err != nil {
			t.Fatal(err)
		}
	})
}
