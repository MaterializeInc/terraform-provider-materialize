package materialize

import (
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"
)

type ConnectionMySQLBuilder struct {
	Connection
	connectionType string
	mysqlHost      string
	mysqlPort      int
	mysqlUser      ValueSecretStruct
	mysqlPassword  IdentifierSchemaStruct
	mysqlSSHTunnel IdentifierSchemaStruct
	mysqlSSLMode   string
	mysqlSSLCa     ValueSecretStruct
	mysqlSSLCert   ValueSecretStruct
	mysqlSSLKey    IdentifierSchemaStruct
	validate       bool
}

func NewConnectionMySQLBuilder(conn *sqlx.DB, obj MaterializeObject) *ConnectionMySQLBuilder {
	b := Builder{conn, BaseConnection}
	return &ConnectionMySQLBuilder{
		Connection: Connection{b, obj.Name, obj.SchemaName, obj.DatabaseName},
	}
}

func (b *ConnectionMySQLBuilder) ConnectionType(connectionType string) *ConnectionMySQLBuilder {
	b.connectionType = connectionType
	return b
}

func (b *ConnectionMySQLBuilder) MySQLHost(mysqlHost string) *ConnectionMySQLBuilder {
	b.mysqlHost = mysqlHost
	return b
}

func (b *ConnectionMySQLBuilder) MySQLPort(mysqlPort int) *ConnectionMySQLBuilder {
	b.mysqlPort = mysqlPort
	return b
}

func (b *ConnectionMySQLBuilder) MySQLUser(mysqlUser ValueSecretStruct) *ConnectionMySQLBuilder {
	b.mysqlUser = mysqlUser
	return b
}

func (b *ConnectionMySQLBuilder) MySQLPassword(mysqlPassword IdentifierSchemaStruct) *ConnectionMySQLBuilder {
	b.mysqlPassword = mysqlPassword
	return b
}

func (b *ConnectionMySQLBuilder) MySQLSSHTunnel(mysqlSSHTunnel IdentifierSchemaStruct) *ConnectionMySQLBuilder {
	b.mysqlSSHTunnel = mysqlSSHTunnel
	return b
}

func (b *ConnectionMySQLBuilder) MySQLSSLMode(mysqlSSLMode string) *ConnectionMySQLBuilder {
	b.mysqlSSLMode = mysqlSSLMode
	return b
}

func (b *ConnectionMySQLBuilder) MySQLSSLCa(mysqlSSLCa ValueSecretStruct) *ConnectionMySQLBuilder {
	b.mysqlSSLCa = mysqlSSLCa
	return b
}

func (b *ConnectionMySQLBuilder) MySQLSSLCert(mysqlSSLCert ValueSecretStruct) *ConnectionMySQLBuilder {
	b.mysqlSSLCert = mysqlSSLCert
	return b
}

func (b *ConnectionMySQLBuilder) MySQLSSLKey(mysqlSSLKey IdentifierSchemaStruct) *ConnectionMySQLBuilder {
	b.mysqlSSLKey = mysqlSSLKey
	return b
}

func (b *ConnectionMySQLBuilder) Validate(validate bool) *ConnectionMySQLBuilder {
	b.validate = validate
	return b
}

func (b *ConnectionMySQLBuilder) Create() error {
	q := strings.Builder{}
	q.WriteString(fmt.Sprintf(`CREATE CONNECTION %s TO MYSQL (`, b.QualifiedName()))

	q.WriteString(fmt.Sprintf(`HOST %s`, QuoteString(b.mysqlHost)))
	q.WriteString(fmt.Sprintf(`, PORT %d`, b.mysqlPort))
	if b.mysqlUser.Text != "" {
		q.WriteString(fmt.Sprintf(`, USER %s`, QuoteString(b.mysqlUser.Text)))
	}
	if b.mysqlUser.Secret.Name != "" {
		q.WriteString(fmt.Sprintf(`, USER SECRET %s`, b.mysqlUser.Secret.QualifiedName()))
	}
	if b.mysqlPassword.Name != "" {
		q.WriteString(fmt.Sprintf(`, PASSWORD SECRET %s`, b.mysqlPassword.QualifiedName()))
	}
	if b.mysqlSSLMode != "" {
		q.WriteString(fmt.Sprintf(`, SSL MODE %s`, QuoteString(b.mysqlSSLMode)))
	}
	if b.mysqlSSHTunnel.Name != "" {
		q.WriteString(fmt.Sprintf(`, SSH TUNNEL %s`, b.mysqlSSHTunnel.QualifiedName()))
	}
	if b.mysqlSSLCa.Secret.Name != "" {
		q.WriteString(fmt.Sprintf(`, SSL CERTIFICATE AUTHORITY SECRET %s`, b.mysqlSSLCa.Secret.QualifiedName()))
	}
	if b.mysqlSSLCert.Secret.Name != "" {
		q.WriteString(fmt.Sprintf(`, SSL CERTIFICATE SECRET %s`, b.mysqlSSLCert.Secret.QualifiedName()))
	}
	if b.mysqlSSLKey.Name != "" {
		q.WriteString(fmt.Sprintf(`, SSL KEY SECRET %s`, b.mysqlSSLKey.QualifiedName()))
	}

	q.WriteString(`)`)

	if !b.validate {
		q.WriteString(` WITH (VALIDATE = false)`)
	}

	q.WriteString(`;`)
	return b.ddl.exec(q.String())
}
