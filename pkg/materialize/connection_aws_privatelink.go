package materialize

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"
)

type ConnectionAwsPrivatelinkBuilder struct {
	Connection
	privateLinkServiceName       string
	privateLinkAvailabilityZones []string
}

func NewConnectionAwsPrivatelinkBuilder(conn *sqlx.DB, connectionName, schemaName, databaseName string) *ConnectionAwsPrivatelinkBuilder {
	b := Builder{conn, BaseConnection}
	return &ConnectionAwsPrivatelinkBuilder{
		Connection: Connection{b, connectionName, schemaName, databaseName},
	}
}

func (b *ConnectionAwsPrivatelinkBuilder) PrivateLinkServiceName(privateLinkServiceName string) *ConnectionAwsPrivatelinkBuilder {
	b.privateLinkServiceName = privateLinkServiceName
	return b
}

func (b *ConnectionAwsPrivatelinkBuilder) PrivateLinkAvailabilityZones(privateLinkAvailabilityZones []string) *ConnectionAwsPrivatelinkBuilder {
	b.privateLinkAvailabilityZones = privateLinkAvailabilityZones
	return b
}

func (b *ConnectionAwsPrivatelinkBuilder) Create() error {
	q := strings.Builder{}
	q.WriteString(fmt.Sprintf(`CREATE CONNECTION %s TO AWS PRIVATELINK (`, b.QualifiedName()))

	q.WriteString(fmt.Sprintf(`SERVICE NAME %s,`, QuoteString(b.privateLinkServiceName)))
	q.WriteString(`AVAILABILITY ZONES (`)
	for i, az := range b.privateLinkAvailabilityZones {
		if i > 0 {
			q.WriteString(`, `)
		}
		q.WriteString(QuoteString(az))
	}

	q.WriteString(`));`)
	return b.ddl.exec(q.String())
}

type ConnectionAwsPrivatelinkParams struct {
	ConnectionId   sql.NullString `db:"id"`
	ConnectionName sql.NullString `db:"connection_name"`
	SchemaName     sql.NullString `db:"schema_name"`
	DatabaseName   sql.NullString `db:"database_name"`
	Principal      sql.NullString `db:"principal"`
}

var connectionAwsPrivatelinkQuery = `
	SELECT
		mz_connections.id,
		mz_connections.name AS connection_name,
		mz_schemas.name AS schema_name,
		mz_databases.name AS database_name,
		mz_aws_privatelink_connections.principal
	FROM mz_connections
	JOIN mz_schemas
		ON mz_connections.schema_id = mz_schemas.id
	JOIN mz_databases
		ON mz_schemas.database_id = mz_databases.id
	LEFT JOIN mz_aws_privatelink_connections
		ON mz_connections.id = mz_aws_privatelink_connections.id`

func ScanConnectionAwsPrivatelink(conn *sqlx.DB, id string) (ConnectionAwsPrivatelinkParams, error) {
	p := map[string]string{"mz_connections.id": id}
	q := queryPredicate(connectionAwsPrivatelinkQuery, p)

	var c ConnectionAwsPrivatelinkParams
	if err := conn.Get(&c, q); err != nil {
		return c, err
	}

	return c, nil
}
