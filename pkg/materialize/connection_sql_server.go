package materialize

import (
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"
)

type ConnectionSqlServerBuilder struct {
	Connection
	host           string
	port           int
	user           ValueSecretStruct
	password       IdentifierSchemaStruct
	database       string
	sslMode        string
	sslCa          ValueSecretStruct
	sslKey         IdentifierSchemaStruct
	sshTunnel      IdentifierSchemaStruct
	awsPrivatelink IdentifierSchemaStruct
	validate       bool
}

func NewConnectionSqlServerBuilder(conn *sqlx.DB, obj MaterializeObject) *ConnectionSqlServerBuilder {
	b := Builder{conn, BaseConnection}
	return &ConnectionSqlServerBuilder{
		Connection: Connection{b, obj.Name, obj.SchemaName, obj.DatabaseName},
	}
}

func (b *ConnectionSqlServerBuilder) SqlServerHost(host string) *ConnectionSqlServerBuilder {
	b.host = host
	return b
}

func (b *ConnectionSqlServerBuilder) SqlServerPort(port int) *ConnectionSqlServerBuilder {
	b.port = port
	return b
}

func (b *ConnectionSqlServerBuilder) SqlServerUser(user ValueSecretStruct) *ConnectionSqlServerBuilder {
	b.user = user
	return b
}

func (b *ConnectionSqlServerBuilder) SqlServerPassword(password IdentifierSchemaStruct) *ConnectionSqlServerBuilder {
	b.password = password
	return b
}

func (b *ConnectionSqlServerBuilder) SqlServerDatabase(database string) *ConnectionSqlServerBuilder {
	b.database = database
	return b
}

func (b *ConnectionSqlServerBuilder) SqlServerSSLMode(sslMode string) *ConnectionSqlServerBuilder {
	b.sslMode = sslMode
	return b
}

func (b *ConnectionSqlServerBuilder) SqlServerSSLCa(sslCa ValueSecretStruct) *ConnectionSqlServerBuilder {
	b.sslCa = sslCa
	return b
}

func (b *ConnectionSqlServerBuilder) SqlServerSSLKey(sslKey IdentifierSchemaStruct) *ConnectionSqlServerBuilder {
	b.sslKey = sslKey
	return b
}

func (b *ConnectionSqlServerBuilder) SqlServerSSHTunnel(sshTunnel IdentifierSchemaStruct) *ConnectionSqlServerBuilder {
	b.sshTunnel = sshTunnel
	return b
}

func (b *ConnectionSqlServerBuilder) SqlServerAWSPrivateLink(awsPrivatelink IdentifierSchemaStruct) *ConnectionSqlServerBuilder {
	b.awsPrivatelink = awsPrivatelink
	return b
}

func (b *ConnectionSqlServerBuilder) Validate(validate bool) *ConnectionSqlServerBuilder {
	b.validate = validate
	return b
}

func (b *ConnectionSqlServerBuilder) Create() error {
	q := strings.Builder{}
	q.WriteString(fmt.Sprintf(`CREATE CONNECTION %s TO SQL SERVER (`, b.QualifiedName()))

	var options []string

	if b.host != "" {
		options = append(options, fmt.Sprintf(`HOST '%s'`, b.host))
	}

	if b.port > 0 {
		options = append(options, fmt.Sprintf(`PORT %d`, b.port))
	}

	if b.user.Text != "" {
		options = append(options, fmt.Sprintf(`USER '%s'`, b.user.Text))
	}
	if b.user.Secret.Name != "" {
		options = append(options, fmt.Sprintf(`USER SECRET %s`, b.user.Secret.QualifiedName()))
	}

	if b.password.Name != "" {
		options = append(options, fmt.Sprintf(`PASSWORD SECRET %s`, b.password.QualifiedName()))
	}

	if b.database != "" {
		options = append(options, fmt.Sprintf(`DATABASE '%s'`, b.database))
	}

	if b.sslMode != "" {
		options = append(options, fmt.Sprintf(`SSL MODE '%s'`, b.sslMode))
	}

	if b.sslCa.Text != "" {
		options = append(options, fmt.Sprintf(`SSL CERTIFICATE AUTHORITY '%s'`, b.sslCa.Text))
	}
	if b.sslCa.Secret.Name != "" {
		options = append(options, fmt.Sprintf(`SSL CERTIFICATE AUTHORITY SECRET %s`, b.sslCa.Secret.QualifiedName()))
	}

	if b.sslKey.Name != "" {
		options = append(options, fmt.Sprintf(`SSL KEY SECRET %s`, b.sslKey.QualifiedName()))
	}

	if b.sshTunnel.Name != "" {
		options = append(options, fmt.Sprintf(`SSH TUNNEL %s`, b.sshTunnel.QualifiedName()))
	}

	if b.awsPrivatelink.Name != "" {
		options = append(options, fmt.Sprintf(`AWS PRIVATELINK %s`, b.awsPrivatelink.QualifiedName()))
	}

	q.WriteString(strings.Join(options, ", "))
	q.WriteString(")")

	if b.validate {
		q.WriteString(" WITH (VALIDATE)")
	} else {
		q.WriteString(" WITH (VALIDATE = false)")
	}

	q.WriteString(`;`)
	return b.ddl.exec(q.String())
}

func (b *ConnectionSqlServerBuilder) Rename(newName string) error {
	n := b.QualifiedName()
	return b.ddl.rename(n, QualifiedName(b.DatabaseName, b.SchemaName, newName))
}

func (b *ConnectionSqlServerBuilder) Drop() error {
	qn := b.QualifiedName()
	return b.ddl.drop(qn)
}
