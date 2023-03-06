package resources

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestResourceConnectoinReadId(t *testing.T) {
	r := require.New(t)
	b := newConnectionKafkaBuilder("connection", "schema", "database")
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
	b := newConnectionKafkaBuilder("connection", "schema", "database")
	r.Equal(`ALTER CONNECTION database.schema.connection RENAME TO database.schema.new_connection;`, b.Rename("new_connection"))
}

func TestResourceConnectionDrop(t *testing.T) {
	r := require.New(t)
	b := newConnectionKafkaBuilder("connection", "schema", "database")
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

func TestResourceConnectionCreatePostgres(t *testing.T) {
	r := require.New(t)
	b := newConnectionPostgresBuilder("postgres_conn", "schema", "database")
	b.PostgresHost("postgres_host")
	b.PostgresPort(5432)
	b.PostgresUser("user")
	b.PostgresPassword("password")
	b.PostgresDatabase("default")
	r.Equal(`CREATE CONNECTION database.schema.postgres_conn TO POSTGRES (HOST 'postgres_host', PORT 5432, USER 'user', PASSWORD SECRET password, DATABASE 'default');`, b.Create())
}

func TestResourceConnectionCreatePostgresSsh(t *testing.T) {
	r := require.New(t)
	b := newConnectionPostgresBuilder("postgres_conn", "schema", "database")
	b.PostgresHost("postgres_host")
	b.PostgresPort(5432)
	b.PostgresUser("user")
	b.PostgresPassword("password")
	b.PostgresDatabase("default")
	b.PostgresSSHTunnel("ssh_conn")
	r.Equal(`CREATE CONNECTION database.schema.postgres_conn TO POSTGRES (HOST 'postgres_host', PORT 5432, USER 'user', PASSWORD SECRET password, SSH TUNNEL 'ssh_conn', DATABASE 'default');`, b.Create())
}

func TestResourceConnectionCreatePostgresPrivateLink(t *testing.T) {
	r := require.New(t)
	b := newConnectionPostgresBuilder("postgres_conn", "schema", "database")
	b.PostgresHost("postgres_host")
	b.PostgresPort(5432)
	b.PostgresUser("user")
	b.PostgresPassword("password")
	b.PostgresDatabase("default")
	b.PostgresAWSPrivateLink("private_link")
	r.Equal(`CREATE CONNECTION database.schema.postgres_conn TO POSTGRES (HOST 'postgres_host', PORT 5432, USER 'user', PASSWORD SECRET password, AWS PRIVATELINK private_link, DATABASE 'default');`, b.Create())
}

func TestResourceConnectionCreatePostgresSsl(t *testing.T) {
	r := require.New(t)
	b := newConnectionPostgresBuilder("postgres_conn", "schema", "database")
	b.PostgresHost("postgres_host")
	b.PostgresPort(5432)
	b.PostgresUser("user")
	b.PostgresPassword("password")
	b.PostgresDatabase("default")
	b.PostgresSSLMode("verify-full")
	b.PostgresSSLCa("root")
	b.PostgresSSLCert("cert")
	b.PostgresSSLKey("key")
	r.Equal(`CREATE CONNECTION database.schema.postgres_conn TO POSTGRES (HOST 'postgres_host', PORT 5432, USER 'user', PASSWORD SECRET password, SSL MODE 'verify-full', SSL CERTIFICATE AUTHORITY SECRET root, SSL CERTIFICATE SECRET cert, SSL KEY SECRET key, DATABASE 'default');`, b.Create())
}
