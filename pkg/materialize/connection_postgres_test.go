package materialize

import (
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/MaterializeInc/terraform-provider-materialize/pkg/testhelpers"
	"github.com/jmoiron/sqlx"
)

var connPostgres = ObjectSchemaStruct{Name: "postgres_conn", SchemaName: "schema", DatabaseName: "database"}

func TestConnectionPostgresCreate(t *testing.T) {
	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		mock.ExpectExec(
			`CREATE CONNECTION "database"."schema"."postgres_conn" TO POSTGRES \(HOST 'postgres_host', PORT 5432, USER 'user', PASSWORD SECRET "database"."schema"."password", DATABASE 'default'\);`,
		).WillReturnResult(sqlmock.NewResult(1, 1))

		b := NewConnectionPostgresBuilder(db, connPostgres)
		b.PostgresHost("postgres_host")
		b.PostgresPort(5432)
		b.PostgresUser(ValueSecretStruct{Text: "user"})
		b.PostgresPassword(IdentifierSchemaStruct{Name: "password", SchemaName: "schema", DatabaseName: "database"})
		b.PostgresDatabase("default")

		if err := b.Create(); err != nil {
			t.Fatal(err)
		}
	})
}

func TestConnectionPostgresSshCreate(t *testing.T) {
	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		mock.ExpectExec(
			`CREATE CONNECTION "database"."schema"."postgres_conn" TO POSTGRES \(HOST 'postgres_host', PORT 5432, USER 'user', PASSWORD SECRET "database"."schema"."password", SSH TUNNEL "database"."schema"."ssh_conn", DATABASE 'default'\);`,
		).WillReturnResult(sqlmock.NewResult(1, 1))

		b := NewConnectionPostgresBuilder(db, connPostgres)
		b.PostgresHost("postgres_host")
		b.PostgresPort(5432)
		b.PostgresUser(ValueSecretStruct{Text: "user"})
		b.PostgresPassword(IdentifierSchemaStruct{Name: "password", SchemaName: "schema", DatabaseName: "database"})
		b.PostgresDatabase("default")
		b.PostgresSSHTunnel(IdentifierSchemaStruct{Name: "ssh_conn", SchemaName: "schema", DatabaseName: "database"})

		if err := b.Create(); err != nil {
			t.Fatal(err)
		}
	})
}

func TestConnectionPostgresPrivateLinkCreate(t *testing.T) {
	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		mock.ExpectExec(
			`CREATE CONNECTION "database"."schema"."postgres_conn" TO POSTGRES \(HOST 'postgres_host', PORT 5432, USER 'user', PASSWORD SECRET "database"."schema"."password", AWS PRIVATELINK "database"."schema"."private_link", DATABASE 'default'\);`,
		).WillReturnResult(sqlmock.NewResult(1, 1))

		b := NewConnectionPostgresBuilder(db, connPostgres)
		b.PostgresHost("postgres_host")
		b.PostgresPort(5432)
		b.PostgresUser(ValueSecretStruct{Text: "user"})
		b.PostgresPassword(IdentifierSchemaStruct{Name: "password", SchemaName: "schema", DatabaseName: "database"})
		b.PostgresDatabase("default")
		b.PostgresAWSPrivateLink(IdentifierSchemaStruct{Name: "private_link", SchemaName: "schema", DatabaseName: "database"})

		if err := b.Create(); err != nil {
			t.Fatal(err)
		}
	})
}

func TestConnectionPostgresSslCreate(t *testing.T) {
	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		mock.ExpectExec(
			`CREATE CONNECTION "database"."schema"."postgres_conn" TO POSTGRES \(HOST 'postgres_host', PORT 5432, USER SECRET "database"."schema"."user", PASSWORD SECRET "database"."schema"."password", SSL MODE 'verify-full', SSL CERTIFICATE AUTHORITY SECRET "database"."schema"."root", SSL CERTIFICATE SECRET "database"."schema"."cert", SSL KEY SECRET "database"."schema"."key", DATABASE 'default'\);`,
		).WillReturnResult(sqlmock.NewResult(1, 1))

		b := NewConnectionPostgresBuilder(db, connPostgres)
		b.PostgresHost("postgres_host")
		b.PostgresPort(5432)
		b.PostgresUser(ValueSecretStruct{Secret: IdentifierSchemaStruct{Name: "user", SchemaName: "schema", DatabaseName: "database"}})
		b.PostgresPassword(IdentifierSchemaStruct{Name: "password", SchemaName: "schema", DatabaseName: "database"})
		b.PostgresDatabase("default")
		b.PostgresSSLMode("verify-full")
		b.PostgresSSLCa(ValueSecretStruct{Secret: IdentifierSchemaStruct{Name: "root", SchemaName: "schema", DatabaseName: "database"}})
		b.PostgresSSLCert(ValueSecretStruct{Secret: IdentifierSchemaStruct{Name: "cert", SchemaName: "schema", DatabaseName: "database"}})
		b.PostgresSSLKey(IdentifierSchemaStruct{Name: "key", SchemaName: "schema", DatabaseName: "database"})

		if err := b.Create(); err != nil {
			t.Fatal(err)
		}
	})
}
