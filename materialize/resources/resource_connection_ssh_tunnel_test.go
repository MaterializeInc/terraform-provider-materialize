package resources

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestResourceConnectionSshTunnelReadId(t *testing.T) {
	r := require.New(t)
	b := newConnectionSshTunnelBuilder("connection", "schema", "database")
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

func TestResourceConnectionSshTunnelRename(t *testing.T) {
	r := require.New(t)
	b := newConnectionSshTunnelBuilder("connection", "schema", "database")
	r.Equal(`ALTER CONNECTION database.schema.connection RENAME TO database.schema.new_connection;`, b.Rename("new_connection"))
}

func TestResourceConnectionSshTunnelDrop(t *testing.T) {
	r := require.New(t)
	b := newConnectionSshTunnelBuilder("connection", "schema", "database")
	r.Equal(`DROP CONNECTION database.schema.connection;`, b.Drop())
}

func TestResourceConnectionSshTunnelReadParams(t *testing.T) {
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

func TestResourceConnectionCreateSsh(t *testing.T) {
	r := require.New(t)

	b := newConnectionSshTunnelBuilder("ssh_conn", "schema", "database")
	b.SSHHost("localhost")
	b.SSHPort(123)
	b.SSHUser("user")
	r.Equal(`CREATE CONNECTION database.schema.ssh_conn TO SSH TUNNEL (HOST 'localhost', USER 'user', PORT 123);`, b.Create())

}
