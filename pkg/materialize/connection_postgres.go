package materialize

import (
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"
)

type ConnectionPostgresBuilder struct {
	Connection
	connectionType         string
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

func NewConnectionPostgresBuilder(conn *sqlx.DB, connectionName, schemaName, databaseName string) *ConnectionPostgresBuilder {
	b := Builder{conn, BaseConnection}
	return &ConnectionPostgresBuilder{
		Connection: Connection{b, connectionName, schemaName, databaseName},
	}
}

func (b *ConnectionPostgresBuilder) ConnectionType(connectionType string) *ConnectionPostgresBuilder {
	b.connectionType = connectionType
	return b
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

func (b *ConnectionPostgresBuilder) Create() error {
	q := strings.Builder{}
	q.WriteString(fmt.Sprintf(`CREATE CONNECTION %s TO POSTGRES (`, b.QualifiedName()))

	q.WriteString(fmt.Sprintf(`HOST %s`, QuoteString(b.postgresHost)))
	q.WriteString(fmt.Sprintf(`, PORT %d`, b.postgresPort))
	if b.postgresUser.Text != "" {
		q.WriteString(fmt.Sprintf(`, USER %s`, QuoteString(b.postgresUser.Text)))
	}
	if b.postgresUser.Secret.Name != "" {
		q.WriteString(fmt.Sprintf(`, USER SECRET %s`, b.postgresUser.Secret.QualifiedName()))
	}
	if b.postgresPassword.Name != "" {
		q.WriteString(fmt.Sprintf(`, PASSWORD SECRET %s`, b.postgresPassword.QualifiedName()))
	}
	if b.postgresSSLMode != "" {
		q.WriteString(fmt.Sprintf(`, SSL MODE %s`, QuoteString(b.postgresSSLMode)))
	}
	if b.postgresSSHTunnel.Name != "" {
		q.WriteString(fmt.Sprintf(`, SSH TUNNEL %s`, b.postgresSSHTunnel.QualifiedName()))
	}
	if b.postgresSSLCa.Text != "" {
		q.WriteString(fmt.Sprintf(`, SSL CERTIFICATE AUTHORITY %s`, QuoteString(b.postgresSSLCa.Text)))
	}
	if b.postgresSSLCa.Secret.Name != "" {
		q.WriteString(fmt.Sprintf(`, SSL CERTIFICATE AUTHORITY SECRET %s`, b.postgresSSLCa.Secret.QualifiedName()))
	}
	if b.postgresSSLCert.Text != "" {
		q.WriteString(fmt.Sprintf(`, SSL CERTIFICATE  %s`, QuoteString(b.postgresSSLCert.Text)))
	}
	if b.postgresSSLCert.Secret.Name != "" {
		q.WriteString(fmt.Sprintf(`, SSL CERTIFICATE SECRET %s`, b.postgresSSLCert.Secret.QualifiedName()))
	}
	if b.postgresSSLKey.Name != "" {
		q.WriteString(fmt.Sprintf(`, SSL KEY SECRET %s`, b.postgresSSLKey.QualifiedName()))
	}
	if b.postgresAWSPrivateLink.Name != "" {
		q.WriteString(fmt.Sprintf(`, AWS PRIVATELINK %s`, b.postgresAWSPrivateLink.QualifiedName()))
	}

	q.WriteString(fmt.Sprintf(`, DATABASE %s`, QuoteString(b.postgresDatabase)))

	q.WriteString(`);`)
	return b.ddl.exec(q.String())
}
