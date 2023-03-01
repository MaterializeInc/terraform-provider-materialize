package resources

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestResourceConnectionConfluentSchemaRegistryReadId(t *testing.T) {
	r := require.New(t)
	b := newConnectionConfluentSchemaRegistryBuilder("connection", "schema", "database")
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

func TestResourceConnectionConfluentSchemaRegistryRename(t *testing.T) {
	r := require.New(t)
	b := newConnectionConfluentSchemaRegistryBuilder("connection", "schema", "database")
	r.Equal(`ALTER CONNECTION database.schema.connection RENAME TO database.schema.new_connection;`, b.Rename("new_connection"))
}

func TestResourceConnectionConfluentSchemaRegistryDrop(t *testing.T) {
	r := require.New(t)
	b := newConnectionConfluentSchemaRegistryBuilder("connection", "schema", "database")
	r.Equal(`DROP CONNECTION database.schema.connection;`, b.Drop())
}

func TestResourceConnectionConfluentSchemaRegistryReadParams(t *testing.T) {
	r := require.New(t)
	b := readConnectionParams("u1")
	r.Equal(`
		SELECT
			mz_connections.name,
			mz_schemas.name,
			mz_databases.name,
			mz_connections.type
		FROM mz_connections
		JOIN mz_schemas
			ON mz_connections.schema_id = mz_schemas.id
		JOIN mz_databases
			ON mz_schemas.database_id = mz_databases.id
		WHERE mz_connections.id = 'u1';`, b)
}

// here are common ^

func TestResourceConnectionCreateConfluentSchemaRegistry(t *testing.T) {
	r := require.New(t)
	b := newConnectionConfluentSchemaRegistryBuilder("csr_conn", "schema", "database")
	b.ConfluentSchemaRegistryUrl("http://localhost:8081")
	b.ConfluentSchemaRegistryUsername("user")
	b.ConfluentSchemaRegistryPassword("password")
	r.Equal(`CREATE CONNECTION database.schema.csr_conn TO CONFLUENT SCHEMA REGISTRY (URL 'http://localhost:8081', USERNAME = 'user', PASSWORD = SECRET password);`, b.Create())

}
