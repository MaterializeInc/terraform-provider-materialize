package materialize

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestResourceConnectionConfluentSchemaRegistryReadId(t *testing.T) {
	r := require.New(t)
	b := NewConnectionConfluentSchemaRegistryBuilder("connection", "schema", "database")
	r.Equal(`
		SELECT mz_connections.id
		FROM mz_connections
		JOIN mz_schemas
			ON mz_connections.schema_id = mz_schemas.id
		JOIN mz_databases
			ON mz_schemas.database_id = mz_databases.id
		WHERE mz_connections.name = 'connection'
		AND mz_schemas.name = 'schema'
		AND mz_databases.name = 'database';
	`, b.ReadId())
}

func TestConnectionConfluentSchemaRegistryRenameQuery(t *testing.T) {
	r := require.New(t)
	b := NewConnectionConfluentSchemaRegistryBuilder("connection", "schema", "database")
	r.Equal(`ALTER CONNECTION "database"."schema"."connection" RENAME TO "database"."schema"."new_connection";`, b.Rename("new_connection"))
}

func TestConnectionConfluentSchemaRegistryDropQuery(t *testing.T) {
	r := require.New(t)
	b := NewConnectionConfluentSchemaRegistryBuilder("connection", "schema", "database")
	r.Equal(`DROP CONNECTION "database"."schema"."connection";`, b.Drop())
}

func TestConnectionConfluentSchemaRegistryReadParamsQuery(t *testing.T) {
	r := require.New(t)
	b := ReadConnectionParams("u1")
	r.Equal(`
		SELECT
			mz_connections.name,
			mz_schemas.name,
			mz_databases.name
		FROM mz_connections
		JOIN mz_schemas
			ON mz_connections.schema_id = mz_schemas.id
		JOIN mz_databases
			ON mz_schemas.database_id = mz_databases.id
		WHERE mz_connections.id = 'u1';`, b)
}

func TestConnectionCreateConfluentSchemaRegistryQuery(t *testing.T) {
	r := require.New(t)
	b := NewConnectionConfluentSchemaRegistryBuilder("csr_conn", "schema", "database")
	b.ConfluentSchemaRegistryUrl("http://localhost:8081")
	b.ConfluentSchemaRegistryUsername(ValueSecretStruct{Text: "user"})
	b.ConfluentSchemaRegistryPassword(IdentifierSchemaStruct{SchemaName: "schema", Name: "password", DatabaseName: "database"})
	r.Equal(`CREATE CONNECTION "database"."schema"."csr_conn" TO CONFLUENT SCHEMA REGISTRY (URL 'http://localhost:8081', USERNAME = 'user', PASSWORD = SECRET "database"."schema"."password");`, b.Create())
}

func TestConnectionCreateConfluentSchemaRegistryQueryUsernameSecret(t *testing.T) {
	r := require.New(t)
	b := NewConnectionConfluentSchemaRegistryBuilder("csr_conn", "schema", "database")
	b.ConfluentSchemaRegistryUrl("http://localhost:8081")
	b.ConfluentSchemaRegistryUsername(ValueSecretStruct{Secret: IdentifierSchemaStruct{SchemaName: "schema", Name: "user", DatabaseName: "database"}})
	b.ConfluentSchemaRegistryPassword(IdentifierSchemaStruct{SchemaName: "schema", Name: "password", DatabaseName: "database"})
	r.Equal(`CREATE CONNECTION "database"."schema"."csr_conn" TO CONFLUENT SCHEMA REGISTRY (URL 'http://localhost:8081', USERNAME = SECRET "database"."schema"."user", PASSWORD = SECRET "database"."schema"."password");`, b.Create())
}
