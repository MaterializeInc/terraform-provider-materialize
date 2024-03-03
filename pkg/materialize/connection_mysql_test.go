package materialize

import (
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/MaterializeInc/terraform-provider-materialize/pkg/testhelpers"
	"github.com/jmoiron/sqlx"
)

var connMySQL = MaterializeObject{Name: "mysql_conn", SchemaName: "schema", DatabaseName: "database"}

func TestConnectionMySQLCreate(t *testing.T) {
	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		mock.ExpectExec(
			`CREATE CONNECTION "database"."schema"."mysql_conn" TO MYSQL \(HOST 'mysql_host', PORT 3306, USER 'user', PASSWORD SECRET "database"."schema"."password"\);`,
		).WillReturnResult(sqlmock.NewResult(1, 1))

		b := NewConnectionMySQLBuilder(db, connMySQL)
		b.MySQLHost("mysql_host")
		b.MySQLPort(3306)
		b.MySQLUser(ValueSecretStruct{Text: "user"})
		b.MySQLPassword(IdentifierSchemaStruct{Name: "password", SchemaName: "schema", DatabaseName: "database"})
		b.Validate(true)

		if err := b.Create(); err != nil {
			t.Fatal(err)
		}
	})
}

func TestConnectionMySQLSshCreate(t *testing.T) {
	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		mock.ExpectExec(
			`CREATE CONNECTION "database"."schema"."mysql_conn" TO MYSQL \(HOST 'mysql_host', PORT 3306, USER 'user', PASSWORD SECRET "database"."schema"."password", SSH TUNNEL "database"."schema"."ssh_conn"\);`,
		).WillReturnResult(sqlmock.NewResult(1, 1))

		b := NewConnectionMySQLBuilder(db, connMySQL)
		b.MySQLHost("mysql_host")
		b.MySQLPort(3306)
		b.MySQLUser(ValueSecretStruct{Text: "user"})
		b.MySQLPassword(IdentifierSchemaStruct{Name: "password", SchemaName: "schema", DatabaseName: "database"})
		b.MySQLSSHTunnel(IdentifierSchemaStruct{Name: "ssh_conn", SchemaName: "schema", DatabaseName: "database"})
		b.Validate(true)

		if err := b.Create(); err != nil {
			t.Fatal(err)
		}
	})
}

func TestConnectionMySQLSslCreate(t *testing.T) {
	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		mock.ExpectExec(
			`CREATE CONNECTION "database"."schema"."mysql_conn" TO MYSQL \(HOST 'mysql_host', PORT 3306, USER SECRET "database"."schema"."user", PASSWORD SECRET "database"."schema"."password", SSL MODE 'verify-ca', SSL CERTIFICATE AUTHORITY SECRET "database"."schema"."root", SSL CERTIFICATE SECRET "database"."schema"."cert", SSL KEY SECRET "database"."schema"."key"\);`,
		).WillReturnResult(sqlmock.NewResult(1, 1))

		b := NewConnectionMySQLBuilder(db, connMySQL)
		b.MySQLHost("mysql_host")
		b.MySQLPort(3306)
		b.MySQLUser(ValueSecretStruct{Secret: IdentifierSchemaStruct{Name: "user", SchemaName: "schema", DatabaseName: "database"}})
		b.MySQLPassword(IdentifierSchemaStruct{Name: "password", SchemaName: "schema", DatabaseName: "database"})
		b.MySQLSSLMode("verify-ca")
		b.MySQLSSLCa(ValueSecretStruct{Secret: IdentifierSchemaStruct{Name: "root", SchemaName: "schema", DatabaseName: "database"}})
		b.MySQLSSLCert(ValueSecretStruct{Secret: IdentifierSchemaStruct{Name: "cert", SchemaName: "schema", DatabaseName: "database"}})
		b.MySQLSSLKey(IdentifierSchemaStruct{Name: "key", SchemaName: "schema", DatabaseName: "database"})
		b.Validate(true)

		if err := b.Create(); err != nil {
			t.Fatal(err)
		}
	})
}
