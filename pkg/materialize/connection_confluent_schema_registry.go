package materialize

import (
	"fmt"
	"strings"
)

type ConnectionConfluentSchemaRegistryBuilder struct {
	Connection
	confluentSchemaRegistryUrl            string
	confluentSchemaRegistrySSLCa          ValueSecretStruct
	confluentSchemaRegistrySSLCert        ValueSecretStruct
	confluentSchemaRegistrySSLKey         IdentifierSchemaStruct
	confluentSchemaRegistryUsername       ValueSecretStruct
	confluentSchemaRegistryPassword       IdentifierSchemaStruct
	confluentSchemaRegistrySSHTunnel      IdentifierSchemaStruct
	confluentSchemaRegistryAWSPrivateLink IdentifierSchemaStruct
}

func NewConnectionConfluentSchemaRegistryBuilder(connectionName, schemaName, databaseName string) *ConnectionConfluentSchemaRegistryBuilder {
	return &ConnectionConfluentSchemaRegistryBuilder{
		Connection: Connection{connectionName, schemaName, databaseName},
	}
}

func (b *ConnectionConfluentSchemaRegistryBuilder) ConfluentSchemaRegistryUrl(confluentSchemaRegistryUrl string) *ConnectionConfluentSchemaRegistryBuilder {
	b.confluentSchemaRegistryUrl = confluentSchemaRegistryUrl
	return b
}

func (b *ConnectionConfluentSchemaRegistryBuilder) ConfluentSchemaRegistryUsername(confluentSchemaRegistryUsername ValueSecretStruct) *ConnectionConfluentSchemaRegistryBuilder {
	b.confluentSchemaRegistryUsername = confluentSchemaRegistryUsername
	return b
}

func (b *ConnectionConfluentSchemaRegistryBuilder) ConfluentSchemaRegistryPassword(confluentSchemaRegistryPassword IdentifierSchemaStruct) *ConnectionConfluentSchemaRegistryBuilder {
	b.confluentSchemaRegistryPassword = confluentSchemaRegistryPassword
	return b
}

func (b *ConnectionConfluentSchemaRegistryBuilder) ConfluentSchemaRegistrySSLCa(confluentSchemaRegistrySSLCa ValueSecretStruct) *ConnectionConfluentSchemaRegistryBuilder {
	b.confluentSchemaRegistrySSLCa = confluentSchemaRegistrySSLCa
	return b
}

func (b *ConnectionConfluentSchemaRegistryBuilder) ConfluentSchemaRegistrySSLCert(confluentSchemaRegistrySSLCert ValueSecretStruct) *ConnectionConfluentSchemaRegistryBuilder {
	b.confluentSchemaRegistrySSLCert = confluentSchemaRegistrySSLCert
	return b
}

func (b *ConnectionConfluentSchemaRegistryBuilder) ConfluentSchemaRegistrySSLKey(confluentSchemaRegistrySSLKey IdentifierSchemaStruct) *ConnectionConfluentSchemaRegistryBuilder {
	b.confluentSchemaRegistrySSLKey = confluentSchemaRegistrySSLKey
	return b
}

func (b *ConnectionConfluentSchemaRegistryBuilder) ConfluentSchemaRegistrySSHTunnel(confluentSchemaRegistrySSHTunnel IdentifierSchemaStruct) *ConnectionConfluentSchemaRegistryBuilder {
	b.confluentSchemaRegistrySSHTunnel = confluentSchemaRegistrySSHTunnel
	return b
}

func (b *ConnectionConfluentSchemaRegistryBuilder) ConfluentSchemaRegistryAWSPrivateLink(confluentSchemaRegistryAWSPrivateLink IdentifierSchemaStruct) *ConnectionConfluentSchemaRegistryBuilder {
	b.confluentSchemaRegistryAWSPrivateLink = confluentSchemaRegistryAWSPrivateLink
	return b
}

func (b *ConnectionConfluentSchemaRegistryBuilder) Create() string {
	q := strings.Builder{}
	q.WriteString(fmt.Sprintf(`CREATE CONNECTION %s TO CONFLUENT SCHEMA REGISTRY (`, b.QualifiedName()))

	q.WriteString(fmt.Sprintf(`URL %s`, QuoteString(b.confluentSchemaRegistryUrl)))
	if b.confluentSchemaRegistryUsername.Text != "" {
		q.WriteString(fmt.Sprintf(`, USERNAME = %s`, QuoteString(b.confluentSchemaRegistryUsername.Text)))
	}
	if b.confluentSchemaRegistryUsername.Secret.Name != "" {
		q.WriteString(fmt.Sprintf(`, USERNAME = SECRET %s`, b.confluentSchemaRegistryUsername.Secret.QualifiedName()))
	}
	if b.confluentSchemaRegistryPassword.Name != "" {
		q.WriteString(fmt.Sprintf(`, PASSWORD = SECRET %s`, b.confluentSchemaRegistryPassword.QualifiedName()))
	}
	if b.confluentSchemaRegistrySSLCa.Text != "" {
		q.WriteString(fmt.Sprintf(`, SSL CERTIFICATE AUTHORITY = %s`, QuoteString(b.confluentSchemaRegistrySSLCa.Text)))
	}
	if b.confluentSchemaRegistrySSLCa.Secret.Name != "" {
		q.WriteString(fmt.Sprintf(`, SSL CERTIFICATE AUTHORITY = SECRET %s`, b.confluentSchemaRegistrySSLCa.Secret.QualifiedName()))
	}
	if b.confluentSchemaRegistrySSLCert.Text != "" {
		q.WriteString(fmt.Sprintf(`, SSL CERTIFICATE = %s`, QuoteString(b.confluentSchemaRegistrySSLCert.Text)))
	}
	if b.confluentSchemaRegistrySSLCert.Secret.Name != "" {
		q.WriteString(fmt.Sprintf(`, SSL CERTIFICATE = SECRET %s`, b.confluentSchemaRegistrySSLCert.Secret.QualifiedName()))
	}
	if b.confluentSchemaRegistrySSLKey.Name != "" {
		q.WriteString(fmt.Sprintf(`, SSL KEY = SECRET %s`, b.confluentSchemaRegistrySSLKey.QualifiedName()))
	}
	if b.confluentSchemaRegistryAWSPrivateLink.Name != "" {
		q.WriteString(fmt.Sprintf(`, AWS PRIVATELINK %s`, QualifiedName(b.confluentSchemaRegistryAWSPrivateLink.DatabaseName, b.confluentSchemaRegistryAWSPrivateLink.SchemaName, b.confluentSchemaRegistryAWSPrivateLink.Name)))
	}
	if b.confluentSchemaRegistrySSHTunnel.Name != "" {
		q.WriteString(fmt.Sprintf(`, SSH TUNNEL %s`, QualifiedName(b.confluentSchemaRegistrySSHTunnel.DatabaseName, b.confluentSchemaRegistrySSHTunnel.SchemaName, b.confluentSchemaRegistrySSHTunnel.Name)))
	}

	q.WriteString(`);`)
	return q.String()
}

func (b *ConnectionConfluentSchemaRegistryBuilder) Rename(newConnectionName string) string {
	n := QualifiedName(b.DatabaseName, b.SchemaName, newConnectionName)
	return fmt.Sprintf(`ALTER CONNECTION %s RENAME TO %s;`, b.QualifiedName(), n)
}

func (b *ConnectionConfluentSchemaRegistryBuilder) Drop() string {
	return fmt.Sprintf(`DROP CONNECTION %s;`, b.QualifiedName())
}

func (b *ConnectionConfluentSchemaRegistryBuilder) ReadId() string {
	return ReadConnectionId(b.ConnectionName, b.SchemaName, b.DatabaseName)
}
