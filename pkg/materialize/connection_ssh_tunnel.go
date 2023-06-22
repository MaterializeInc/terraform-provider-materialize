package materialize

import (
	"database/sql"
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

func NewConnectionSshTunnelBuilder(conn *sqlx.DB, obj ObjectSchemaStruct) *ConnectionSshTunnelBuilder {
	b := Builder{conn, BaseConnection}
	return &ConnectionSshTunnelBuilder{
		Connection: Connection{b, obj.Name, obj.SchemaName, obj.DatabaseName},
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

	q.WriteString(fmt.Sprintf(`HOST %s, USER %s, PORT %d`, QuoteString(b.sshHost), QuoteString(b.sshUser), b.sshPort))

	q.WriteString(`);`)
	return b.ddl.exec(q.String())
}

type ConnectionSshTunnelParams struct {
	ConnectionId   sql.NullString `db:"id"`
	ConnectionName sql.NullString `db:"connection_name"`
	SchemaName     sql.NullString `db:"schema_name"`
	DatabaseName   sql.NullString `db:"database_name"`
	PublicKey1     sql.NullString `db:"public_key_1"`
	PublicKey2     sql.NullString `db:"public_key_2"`
	OwnerName      sql.NullString `db:"owner_name"`
}

var connectionSshTunnelQuery = NewBaseQuery(`
	SELECT
		mz_connections.id,
		mz_connections.name AS connection_name,
		mz_schemas.name AS schema_name,
		mz_databases.name AS database_name,
		mz_ssh_tunnel_connections.public_key_1,
		mz_ssh_tunnel_connections.public_key_2,
		mz_roles.name AS owner_name
	FROM mz_connections
	JOIN mz_schemas
		ON mz_connections.schema_id = mz_schemas.id
	JOIN mz_databases
		ON mz_schemas.database_id = mz_databases.id
	LEFT JOIN mz_ssh_tunnel_connections
		ON mz_connections.id = mz_ssh_tunnel_connections.id
	JOIN mz_roles
		ON mz_connections.owner_id = mz_roles.id`)

func ScanConnectionSshTunnel(conn *sqlx.DB, id string) (ConnectionSshTunnelParams, error) {
	q := connectionSshTunnelQuery.QueryPredicate(map[string]string{"mz_connections.id": id})

	var c ConnectionSshTunnelParams
	if err := conn.Get(&c, q); err != nil {
		return c, err
	}

	return c, nil
}
