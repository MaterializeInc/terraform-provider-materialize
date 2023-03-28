package materialize

import (
	"fmt"
	"strings"
)

type ConnectionPostgresBuilder struct {
	connectionName         string
	schemaName             string
	databaseName           string
	postgresDatabase       string
	postgresHost           string
	postgresPort           int
	postgresUser           ValueSecretStruct
	postgresPassword       IdentifierSchemaStruct
	postgresSSHTunnel      IdentifierSchemaStruct
	postgresSSLCa          ValueSecretStruct
	postgresSSLCert        ValueSecretStruct
	postgresSSLKey         IdentifierSchemaStruct
	postgresSSLMode        string
	postgresAWSPrivateLink IdentifierSchemaStruct
}

func (b *ConnectionPostgresBuilder) qualifiedName() string {
	return QualifiedName(b.databaseName, b.schemaName, b.connectionName)
}

func NewConnectionPostgresBuilder(connectionName, schemaName, databaseName string) *ConnectionPostgresBuilder {
	return &ConnectionPostgresBuilder{
		connectionName: connectionName,
		schemaName:     schemaName,
		databaseName:   databaseName,
	}
}

func (b *ConnectionPostgresBuilder) PostgresDatabase(postgresDatabase string) *ConnectionPostgresBuilder {
	b.postgresDatabase = postgresDatabase
	return b
}

func (b *ConnectionPostgresBuilder) PostgresHost(postgresHost string) *ConnectionPostgresBuilder {
	b.postgresHost = postgresHost
	return b
}

func (b *ConnectionPostgresBuilder) PostgresPort(postgresPort int) *ConnectionPostgresBuilder {
	b.postgresPort = postgresPort
	return b
}

func (b *ConnectionPostgresBuilder) PostgresUser(postgresUser ValueSecretStruct) *ConnectionPostgresBuilder {
	b.postgresUser = postgresUser
	return b
}

func (b *ConnectionPostgresBuilder) PostgresPassword(postgresPassword IdentifierSchemaStruct) *ConnectionPostgresBuilder {
	b.postgresPassword = postgresPassword
	return b
}

func (b *ConnectionPostgresBuilder) PostgresSSHTunnel(postgresSSHTunnel IdentifierSchemaStruct) *ConnectionPostgresBuilder {
	b.postgresSSHTunnel = postgresSSHTunnel
	return b
}

func (b *ConnectionPostgresBuilder) PostgresSSLCa(postgresSSLCa ValueSecretStruct) *ConnectionPostgresBuilder {
	b.postgresSSLCa = postgresSSLCa
	return b
}

func (b *ConnectionPostgresBuilder) PostgresSSLCert(postgresSSLCert ValueSecretStruct) *ConnectionPostgresBuilder {
	b.postgresSSLCert = postgresSSLCert
	return b
}

func (b *ConnectionPostgresBuilder) PostgresSSLKey(postgresSSLKey IdentifierSchemaStruct) *ConnectionPostgresBuilder {
	b.postgresSSLKey = postgresSSLKey
	return b
}

func (b *ConnectionPostgresBuilder) PostgresSSLMode(postgresSSLMode string) *ConnectionPostgresBuilder {
	b.postgresSSLMode = postgresSSLMode
	return b
}

func (b *ConnectionPostgresBuilder) PostgresAWSPrivateLink(postgresAWSPrivateLink IdentifierSchemaStruct) *ConnectionPostgresBuilder {
	b.postgresAWSPrivateLink = postgresAWSPrivateLink
	return b
}

func (b *ConnectionPostgresBuilder) Create() string {
	q := strings.Builder{}
	q.WriteString(fmt.Sprintf(`CREATE CONNECTION %s TO POSTGRES (`, b.qualifiedName()))

	q.WriteString(fmt.Sprintf(`HOST %s`, QuoteString(b.postgresHost)))
	q.WriteString(fmt.Sprintf(`, PORT %d`, b.postgresPort))
	if b.postgresUser.Text != "" {
		q.WriteString(fmt.Sprintf(`, USER %s`, QuoteString(b.postgresUser.Text)))
	}
	if b.postgresUser.Secret.Name != "" {
		q.WriteString(fmt.Sprintf(`, USER SECRET %s`, QualifiedName(b.postgresUser.Secret.DatabaseName, b.postgresUser.Secret.SchemaName, b.postgresUser.Secret.Name)))
	}
	if b.postgresPassword.Name != "" {
		q.WriteString(fmt.Sprintf(`, PASSWORD SECRET %s`, QualifiedName(b.postgresPassword.DatabaseName, b.postgresPassword.SchemaName, b.postgresPassword.Name)))
	}
	if b.postgresSSLMode != "" {
		q.WriteString(fmt.Sprintf(`, SSL MODE %s`, QuoteString(b.postgresSSLMode)))
	}
	if b.postgresSSHTunnel.Name != "" {
		q.WriteString(fmt.Sprintf(`, SSH TUNNEL %s`, QualifiedName(b.postgresSSHTunnel.DatabaseName, b.postgresSSHTunnel.SchemaName, b.postgresSSHTunnel.Name)))
	}
	if b.postgresSSLCa.Text != "" {
		q.WriteString(fmt.Sprintf(`, SSL CERTIFICATE AUTHORITY %s`, QuoteString(b.postgresSSLCa.Text)))
	}
	if b.postgresSSLCa.Secret.Name != "" {
		q.WriteString(fmt.Sprintf(`, SSL CERTIFICATE AUTHORITY SECRET %s`, QualifiedName(b.postgresSSLCa.Secret.DatabaseName, b.postgresSSLCa.Secret.SchemaName, b.postgresSSLCa.Secret.Name)))
	}
	if b.postgresSSLCert.Text != "" {
		q.WriteString(fmt.Sprintf(`, SSL CERTIFICATE  %s`, QuoteString(b.postgresSSLCert.Text)))
	}
	if b.postgresSSLCert.Secret.Name != "" {
		q.WriteString(fmt.Sprintf(`, SSL CERTIFICATE SECRET %s`, QualifiedName(b.postgresSSLCert.Secret.DatabaseName, b.postgresSSLCert.Secret.SchemaName, b.postgresSSLCert.Secret.Name)))
	}
	if b.postgresSSLKey.Name != "" {
		q.WriteString(fmt.Sprintf(`, SSL KEY SECRET %s`, QualifiedName(b.postgresSSLKey.DatabaseName, b.postgresSSLKey.SchemaName, b.postgresSSLKey.Name)))
	}
	if b.postgresAWSPrivateLink.Name != "" {
		q.WriteString(fmt.Sprintf(`, AWS PRIVATELINK %s`, QualifiedName(b.postgresAWSPrivateLink.DatabaseName, b.postgresAWSPrivateLink.SchemaName, b.postgresAWSPrivateLink.Name)))
	}

	q.WriteString(fmt.Sprintf(`, DATABASE %s`, QuoteString(b.postgresDatabase)))

	q.WriteString(`);`)
	return q.String()
}

func (b *ConnectionPostgresBuilder) Rename(newConnectionName string) string {
	n := QualifiedName(b.databaseName, b.schemaName, newConnectionName)
	return fmt.Sprintf(`ALTER CONNECTION %s RENAME TO %s;`, b.qualifiedName(), n)
}

func (b *ConnectionPostgresBuilder) Drop() string {
	return fmt.Sprintf(`DROP CONNECTION %s;`, b.qualifiedName())
}

func (b *ConnectionPostgresBuilder) ReadId() string {
	return ReadConnectionId(b.connectionName, b.schemaName, b.databaseName)

}
