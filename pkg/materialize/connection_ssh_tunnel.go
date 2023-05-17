package materialize

import (
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"
)

type ConnectionSshTunnelBuilder struct {
	Connection
	sshHost string
	sshUser string
	sshPort int
}

func NewConnectionSshTunnelBuilder(conn *sqlx.DB, connectionName, schemaName, databaseName string) *ConnectionSshTunnelBuilder {
	return &ConnectionSshTunnelBuilder{
		Connection: Connection{conn, connectionName, schemaName, databaseName},
	}
}

func (b *ConnectionSshTunnelBuilder) SSHHost(sshHost string) *ConnectionSshTunnelBuilder {
	b.sshHost = sshHost
	return b
}

func (b *ConnectionSshTunnelBuilder) SSHUser(sshUser string) *ConnectionSshTunnelBuilder {
	b.sshUser = sshUser
	return b
}

func (b *ConnectionSshTunnelBuilder) SSHPort(sshPort int) *ConnectionSshTunnelBuilder {
	b.sshPort = sshPort
	return b
}

func (b *ConnectionSshTunnelBuilder) Create() error {
	q := strings.Builder{}
	q.WriteString(fmt.Sprintf(`CREATE CONNECTION %s TO SSH TUNNEL (`, b.QualifiedName()))

	q.WriteString(fmt.Sprintf(`HOST %s, `, QuoteString(b.sshHost)))
	q.WriteString(fmt.Sprintf(`USER %s, `, QuoteString(b.sshUser)))
	q.WriteString(fmt.Sprintf(`PORT %d`, b.sshPort))

	q.WriteString(`);`)

	_, err := b.conn.Exec(q.String())

	if err != nil {
		return err
	}

	return nil
}

type ConnectionSshTunnelParams struct {
	ConnectionName string `db:"name"`
	SchemaName     string `db:"schema"`
	DatabaseName   string `db:"database"`
	PublicKey1     string `db:"pk1"`
	PublicKey2     string `db:"pk2"`
}

func (b *ConnectionSshTunnelBuilder) Params(catalogId string) (ConnectionSshTunnelParams, error) {
	q := fmt.Sprintf(`
		SELECT
			mz_connections.name,
			mz_schemas.name,
			mz_databases.name,
			mz_ssh_tunnel_connections.public_key_1,
			mz_ssh_tunnel_connections.public_key_2
		FROM mz_connections
		JOIN mz_schemas
			ON mz_connections.schema_id = mz_schemas.id
		JOIN mz_databases
			ON mz_schemas.database_id = mz_databases.id
		LEFT JOIN mz_ssh_tunnel_connections
			ON mz_connections.id = mz_ssh_tunnel_connections.id
		WHERE mz_connections.id = %s;
	`, QuoteString(catalogId))

	var s ConnectionSshTunnelParams
	if err := b.conn.Get(&s, q); err != nil {
		return s, err
	}

	return s, nil
}
