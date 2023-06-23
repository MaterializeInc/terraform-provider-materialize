package materialize

import (
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/MaterializeInc/terraform-provider-materialize/pkg/testhelpers"
	"github.com/jmoiron/sqlx"
)

var connConfluentSchema = ObjectSchemaStruct{Name: "csr_conn", SchemaName: "schema", DatabaseName: "database"}

func TestConnectionConfluentSchemaRegistryCreate(t *testing.T) {
	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		mock.ExpectExec(
			`CREATE CONNECTION "database"."schema"."csr_conn" TO CONFLUENT SCHEMA REGISTRY \(URL 'http://localhost:8081', USERNAME = 'user', PASSWORD = SECRET "database"."schema"."password"\);`,
		).WillReturnResult(sqlmock.NewResult(1, 1))

		b := NewConnectionConfluentSchemaRegistryBuilder(db, connConfluentSchema)
		b.ConfluentSchemaRegistryUrl("http://localhost:8081")
		b.ConfluentSchemaRegistryUsername(ValueSecretStruct{Text: "user"})
		b.ConfluentSchemaRegistryPassword(IdentifierSchemaStruct{SchemaName: "schema", Name: "password", DatabaseName: "database"})

		b.Create()
	})
}

func TestConnectionConfluentSchemaRegistryUsernameSecretCreate(t *testing.T) {
	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		mock.ExpectExec(
			`CREATE CONNECTION "database"."schema"."csr_conn" TO CONFLUENT SCHEMA REGISTRY \(URL 'http://localhost:8081', USERNAME = SECRET "database"."schema"."user", PASSWORD = SECRET "database"."schema"."password"\);`,
		).WillReturnResult(sqlmock.NewResult(1, 1))

		b := NewConnectionConfluentSchemaRegistryBuilder(db, connConfluentSchema)
		b.ConfluentSchemaRegistryUrl("http://localhost:8081")
		b.ConfluentSchemaRegistryUsername(ValueSecretStruct{Secret: IdentifierSchemaStruct{SchemaName: "schema", Name: "user", DatabaseName: "database"}})
		b.ConfluentSchemaRegistryPassword(IdentifierSchemaStruct{SchemaName: "schema", Name: "password", DatabaseName: "database"})

		if err := b.Create(); err != nil {
			t.Fatal(err)
		}
	})
}
