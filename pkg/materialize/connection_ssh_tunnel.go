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
