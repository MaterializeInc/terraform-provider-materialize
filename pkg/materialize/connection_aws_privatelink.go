package materialize

import (
	"fmt"
	"strings"
)

type ConnectionAwsPrivatelinkBuilder struct {
	Connection
	privateLinkServiceName       string
	privateLinkAvailabilityZones []string
}

func GetAvailabilityZones(v []interface{}) []string {
	var azs []string
	for _, az := range v {
		azs = append(azs, az.(string))
	}
	return azs
}

func NewConnectionAwsPrivatelinkBuilder(connectionName, schemaName, databaseName string) *ConnectionAwsPrivatelinkBuilder {
	return &ConnectionAwsPrivatelinkBuilder{
		Connection: Connection{connectionName, schemaName, databaseName},
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
	return q.String()
}

func (b *ConnectionAwsPrivatelinkBuilder) Rename(newConnectionName string) string {
	n := QualifiedName(b.DatabaseName, b.SchemaName, newConnectionName)
	return fmt.Sprintf(`ALTER CONNECTION %s RENAME TO %s;`, b.QualifiedName(), n)
}

func (b *ConnectionAwsPrivatelinkBuilder) Drop() string {
	return fmt.Sprintf(`DROP CONNECTION %s;`, b.QualifiedName())
}

func (b *ConnectionAwsPrivatelinkBuilder) ReadId() string {
	return ReadConnectionId(b.ConnectionName, b.SchemaName, b.DatabaseName)
}
