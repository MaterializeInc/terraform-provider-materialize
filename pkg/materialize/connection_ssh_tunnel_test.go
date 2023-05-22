package materialize

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestConnectionSshTunnelReadIdQuery(t *testing.T) {
	r := require.New(t)
	b := NewConnectionSshTunnelBuilder("connection", "schema", "database")
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

func TestConnectionSshTunnelRenameQuery(t *testing.T) {
	r := require.New(t)
	b := NewConnectionSshTunnelBuilder("connection", "schema", "database")
	r.Equal(`ALTER CONNECTION "database"."schema"."connection" RENAME TO "database"."schema"."new_connection";`, b.Rename("new_connection"))
}

func TestConnectionSshTunnelDropQuery(t *testing.T) {
	r := require.New(t)
	b := NewConnectionSshTunnelBuilder("connection", "schema", "database")
	r.Equal(`DROP CONNECTION "database"."schema"."connection";`, b.Drop())
}

func TestConnectionSshTunnelReadParamsQuery(t *testing.T) {
	r := require.New(t)
	b := ReadConnectionSshTunnelParams("u1")
	r.Equal(`
		SELECT
			mz_connections.name AS connection_name,
			mz_schemas.name AS schema_name,
			mz_databases.name AS database_name,
			mz_ssh_tunnel_connections.public_key_1,
			mz_ssh_tunnel_connections.public_key_2
		FROM mz_connections
		JOIN mz_schemas
			ON mz_connections.schema_id = mz_schemas.id
		JOIN mz_databases
			ON mz_schemas.database_id = mz_databases.id
		LEFT JOIN mz_ssh_tunnel_connections
			ON mz_connections.id = mz_ssh_tunnel_connections.id
		WHERE mz_connections.id = 'u1';`, b)
}

func TestConnectionSshTunnelCreateQuery(t *testing.T) {
	r := require.New(t)

	b := NewConnectionSshTunnelBuilder("ssh_conn", "schema", "database")
	b.SSHHost("localhost")
	b.SSHPort(123)
	b.SSHUser("user")
	r.Equal(`CREATE CONNECTION "database"."schema"."ssh_conn" TO SSH TUNNEL (HOST 'localhost', USER 'user', PORT 123);`, b.Create())
}
