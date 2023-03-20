package materialize

import (
	"fmt"
	"strings"
)

type ConnectionSshTunnelBuilder struct {
	connectionName string
	schemaName     string
	databaseName   string
	sshHost        string
	sshUser        string
	sshPort        int
}

func (b *ConnectionSshTunnelBuilder) qualifiedName() string {
	return QualifiedName(b.databaseName, b.schemaName, b.connectionName)
}

func NewConnectionSshTunnelBuilder(connectionName, schemaName, databaseName string) *ConnectionSshTunnelBuilder {
	return &ConnectionSshTunnelBuilder{
		connectionName: connectionName,
		schemaName:     schemaName,
		databaseName:   databaseName,
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

func (b *ConnectionSshTunnelBuilder) Create() string {
	q := strings.Builder{}
	q.WriteString(fmt.Sprintf(`CREATE CONNECTION %s TO SSH TUNNEL (`, b.qualifiedName()))

	q.WriteString(fmt.Sprintf(`HOST %s, `, QuoteString(b.sshHost)))
	q.WriteString(fmt.Sprintf(`USER %s, `, QuoteString(b.sshUser)))
	q.WriteString(fmt.Sprintf(`PORT %d`, b.sshPort))

	q.WriteString(`);`)
	return q.String()
}

func (b *ConnectionSshTunnelBuilder) Rename(newConnectionName string) string {
	n := QualifiedName(b.databaseName, b.schemaName, newConnectionName)
	return fmt.Sprintf(`ALTER CONNECTION %s RENAME TO %s;`, b.qualifiedName(), n)
}

func (b *ConnectionSshTunnelBuilder) Drop() string {
	return fmt.Sprintf(`DROP CONNECTION %s;`, b.qualifiedName())
}

func (b *ConnectionSshTunnelBuilder) ReadId() string {
	return fmt.Sprintf(`
		SELECT mz_connections.id
		FROM mz_connections
		JOIN mz_schemas
			ON mz_connections.schema_id = mz_schemas.id
		JOIN mz_databases
			ON mz_schemas.database_id = mz_databases.id
		WHERE mz_connections.name = %s
		AND mz_schemas.name = %s
		AND mz_databases.name = %s;
	`, QuoteString(b.connectionName), QuoteString(b.schemaName), QuoteString(b.databaseName))
}
