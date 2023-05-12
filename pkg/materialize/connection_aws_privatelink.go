package materialize

import (
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
