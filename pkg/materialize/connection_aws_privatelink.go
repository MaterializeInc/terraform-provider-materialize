package materialize

import (
	"fmt"
	"strings"
)

type ConnectionAwsPrivatelinkBuilder struct {
	connectionName               string
	schemaName                   string
	databaseName                 string
	privateLinkServiceName       string
	privateLinkAvailabilityZones []string
}

func (b *ConnectionAwsPrivatelinkBuilder) qualifiedName() string {
	return QualifiedName(b.databaseName, b.schemaName, b.connectionName)
}

func NewConnectionAwsPrivatelinkBuilder(connectionName, schemaName, databaseName string) *ConnectionAwsPrivatelinkBuilder {
	return &ConnectionAwsPrivatelinkBuilder{
		connectionName: connectionName,
		schemaName:     schemaName,
		databaseName:   databaseName,
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

func (b *ConnectionAwsPrivatelinkBuilder) Create() string {
	q := strings.Builder{}
	q.WriteString(fmt.Sprintf(`CREATE CONNECTION %s TO AWS PRIVATELINK (`, b.qualifiedName()))

	q.WriteString(fmt.Sprintf(`SERVICE NAME %s,`, QuoteString(b.privateLinkServiceName)))
	q.WriteString(`AVAILABILITY ZONES (`)
	for i, az := range b.privateLinkAvailabilityZones {
		if i > 0 {
			q.WriteString(`, `)
		}
		q.WriteString(QuoteString(az))
	}

	q.WriteString(`));`)
	return q.String()
}

func (b *ConnectionAwsPrivatelinkBuilder) Rename(newConnectionName string) string {
	n := QualifiedName(b.databaseName, b.schemaName, newConnectionName)
	return fmt.Sprintf(`ALTER CONNECTION %s RENAME TO %s;`, b.qualifiedName(), n)
}

func (b *ConnectionAwsPrivatelinkBuilder) Drop() string {
	return fmt.Sprintf(`DROP CONNECTION %s;`, b.qualifiedName())
}

func (b *ConnectionAwsPrivatelinkBuilder) ReadId() string {
	return ReadConnectionId(b.connectionName, b.schemaName, b.databaseName)
}
