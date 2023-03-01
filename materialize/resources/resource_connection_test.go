package resources

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestResourceConnectoinReadId(t *testing.T) {
	r := require.New(t)
	b := newConnectionBuilder("connection", "schema", "database")
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

func TestResourceConnectionRename(t *testing.T) {
	r := require.New(t)
	b := newConnectionBuilder("connection", "schema", "database")
	r.Equal(`ALTER CONNECTION database.schema.connection RENAME TO database.schema.new_connection;`, b.Rename("new_connection"))
}

func TestResourceConnectionDrop(t *testing.T) {
	r := require.New(t)
	b := newConnectionBuilder("connection", "schema", "database")
	r.Equal(`DROP CONNECTION database.schema.connection;`, b.Drop())
}

func TestResourceConnectionReadParams(t *testing.T) {
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
func TestResourceConnectionCreateSsh(t *testing.T) {
	r := require.New(t)

	b := newConnectionBuilder("ssh_conn", "schema", "database")
	b.ConnectionType("SSH TUNNEL")
	b.SSHHost("localhost")
	b.SSHPort(123)
	b.SSHUser("user")
	r.Equal(`CREATE CONNECTION database.schema.ssh_conn TO SSH TUNNEL (HOST 'localhost', USER 'user', PORT 123);`, b.Create())

}

func TestResourceConnectionCreateAwsPrivateLink(t *testing.T) {
	r := require.New(t)

	b := newConnectionBuilder("privatelink_conn", "schema", "database")
	b.ConnectionType("AWS PRIVATELINK")
	b.PrivateLinkServiceName("com.amazonaws.us-east-1.materialize.example")
	b.PrivateLinkAvailabilityZones([]string{"use1-az1", "use1-az2"})
	r.Equal(`CREATE CONNECTION database.schema.privatelink_conn TO AWS PRIVATELINK (SERVICE NAME 'com.amazonaws.us-east-1.materialize.example',AVAILABILITY ZONES ('use1-az1', 'use1-az2'));`, b.Create())
}

func TestResourceConnectionCreateConfluentSchemaRegistry(t *testing.T) {
	r := require.New(t)
	b := newConnectionBuilder("csr_conn", "schema", "database")
	b.ConnectionType("CONFLUENT SCHEMA REGISTRY")
	b.ConfluentSchemaRegistryUrl("http://localhost:8081")
	b.ConfluentSchemaRegistryUsername("user")
	b.ConfluentSchemaRegistryPassword("password")
	r.Equal(`CREATE CONNECTION database.schema.csr_conn TO CONFLUENT SCHEMA REGISTRY (URL 'http://localhost:8081', USERNAME = 'user', PASSWORD = SECRET password);`, b.Create())

}
