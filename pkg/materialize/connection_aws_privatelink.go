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
	return &ConnectionAwsPrivatelinkBuilder{
		Connection: Connection{conn, connectionName, schemaName, databaseName},
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

	_, err := b.conn.Exec(q.String())

	if err != nil {
		return err
	}

	return nil
}

type ConnectionAwsPrivatelinkParams struct {
	ConnectionName sql.NullString `db:"name"`
	SchemaName     sql.NullString `db:"schema"`
	DatabaseName   sql.NullString `db:"database"`
	Principal      sql.NullString `db:"principal"`
}

func (b *ConnectionAwsPrivatelinkBuilder) Params(catalogId string) (ConnectionAwsPrivatelinkParams, error) {
	q := fmt.Sprintf(`
		SELECT
			mz_connections.name,
			mz_schemas.name,
			mz_databases.name,
			mz_aws_privatelink_connections.principal
		FROM mz_connections
		JOIN mz_schemas
			ON mz_connections.schema_id = mz_schemas.id
		JOIN mz_databases
			ON mz_schemas.database_id = mz_databases.id
		JOIN mz_aws_privatelink_connections
			ON mz_connections.id = mz_aws_privatelink_connections.id
		WHERE mz_connections.id = %s;
	`, QuoteString(catalogId))

	var s ConnectionAwsPrivatelinkParams
	if err := b.conn.Get(&s, q); err != nil {
		return s, err
	}

	return s, nil
}
