package materialize

import (
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"
)

type ConnectionSQLServerBuilder struct {
	Connection
	connectionType          string
	sqlserverDatabase       string
	sqlserverHost           string
	sqlserverPort           int
	sqlserverUser           ValueSecretStruct
	sqlserverPassword       IdentifierSchemaStruct
	sqlserverSSHTunnel      IdentifierSchemaStruct
	sqlserverAWSPrivateLink IdentifierSchemaStruct
	validate                bool
}

func NewConnectionSQLServerBuilder(conn *sqlx.DB, obj MaterializeObject) *ConnectionSQLServerBuilder {
	b := Builder{conn, BaseConnection}
	return &ConnectionSQLServerBuilder{
		Connection: Connection{b, obj.Name, obj.SchemaName, obj.DatabaseName},
	}
}

func (b *ConnectionSQLServerBuilder) ConnectionType(connectionType string) *ConnectionSQLServerBuilder {
	b.connectionType = connectionType
	return b
}

func (b *ConnectionSQLServerBuilder) SQLServerDatabase(sqlserverDatabase string) *ConnectionSQLServerBuilder {
	b.sqlserverDatabase = sqlserverDatabase
	return b
}

func (b *ConnectionSQLServerBuilder) SQLServerHost(sqlserverHost string) *ConnectionSQLServerBuilder {
	b.sqlserverHost = sqlserverHost
	return b
}

func (b *ConnectionSQLServerBuilder) SQLServerPort(sqlserverPort int) *ConnectionSQLServerBuilder {
	b.sqlserverPort = sqlserverPort
	return b
}

func (b *ConnectionSQLServerBuilder) SQLServerUser(sqlserverUser ValueSecretStruct) *ConnectionSQLServerBuilder {
	b.sqlserverUser = sqlserverUser
	return b
}

func (b *ConnectionSQLServerBuilder) SQLServerPassword(sqlserverPassword IdentifierSchemaStruct) *ConnectionSQLServerBuilder {
	b.sqlserverPassword = sqlserverPassword
	return b
}

func (b *ConnectionSQLServerBuilder) SQLServerSSHTunnel(sqlserverSSHTunnel IdentifierSchemaStruct) *ConnectionSQLServerBuilder {
	b.sqlserverSSHTunnel = sqlserverSSHTunnel
	return b
}

func (b *ConnectionSQLServerBuilder) SQLServerAWSPrivateLink(sqlserverAWSPrivateLink IdentifierSchemaStruct) *ConnectionSQLServerBuilder {
	b.sqlserverAWSPrivateLink = sqlserverAWSPrivateLink
	return b
}

func (b *ConnectionSQLServerBuilder) Validate(validate bool) *ConnectionSQLServerBuilder {
	b.validate = validate
	return b
}

func (b *ConnectionSQLServerBuilder) Create() error {
	q := strings.Builder{}
	q.WriteString(fmt.Sprintf(`CREATE CONNECTION %s TO SQL SERVER (`, b.QualifiedName()))

	q.WriteString(fmt.Sprintf(`HOST %s`, QuoteString(b.sqlserverHost)))
	q.WriteString(fmt.Sprintf(`, PORT %d`, b.sqlserverPort))

	if b.sqlserverUser.Text != "" {
		q.WriteString(fmt.Sprintf(`, USER %s`, QuoteString(b.sqlserverUser.Text)))
	}
	if b.sqlserverUser.Secret.Name != "" {
		q.WriteString(fmt.Sprintf(`, USER SECRET %s`, b.sqlserverUser.Secret.QualifiedName()))
	}
	if b.sqlserverPassword.Name != "" {
		q.WriteString(fmt.Sprintf(`, PASSWORD SECRET %s`, b.sqlserverPassword.QualifiedName()))
	}
	if b.sqlserverSSHTunnel.Name != "" {
		q.WriteString(fmt.Sprintf(`, SSH TUNNEL %s`, b.sqlserverSSHTunnel.QualifiedName()))
	}
	if b.sqlserverAWSPrivateLink.Name != "" {
		q.WriteString(fmt.Sprintf(`, AWS PRIVATELINK %s`, b.sqlserverAWSPrivateLink.QualifiedName()))
	}

	q.WriteString(fmt.Sprintf(`, DATABASE %s`, QuoteString(b.sqlserverDatabase)))

	q.WriteString(`)`)

	if !b.validate {
		q.WriteString(` WITH (VALIDATE = false)`)
	}

	q.WriteString(`;`)
	return b.ddl.exec(q.String())
}
