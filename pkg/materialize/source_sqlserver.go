package materialize

import (
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"
)

type SourceSQLServerBuilder struct {
	Source
	clusterName             string
	size                    string
	sqlserverConnection     IdentifierSchemaStruct
	textColumns             []string
	excludeColumns          []string
	table                   []TableStruct
	exposeProgress          IdentifierSchemaStruct
	sslMode                 string
	sslCertificateAuthority ValueSecretStruct
	awsPrivateLink          IdentifierSchemaStruct
}

func NewSourceSQLServerBuilder(conn *sqlx.DB, obj MaterializeObject) *SourceSQLServerBuilder {
	b := Builder{conn, BaseSource}
	return &SourceSQLServerBuilder{
		Source: Source{b, obj.Name, obj.SchemaName, obj.DatabaseName},
	}
}

func (b *SourceSQLServerBuilder) ClusterName(c string) *SourceSQLServerBuilder {
	b.clusterName = c
	return b
}

func (b *SourceSQLServerBuilder) Size(s string) *SourceSQLServerBuilder {
	b.size = s
	return b
}

func (b *SourceSQLServerBuilder) SQLServerConnection(conn IdentifierSchemaStruct) *SourceSQLServerBuilder {
	b.sqlserverConnection = conn
	return b
}

func (b *SourceSQLServerBuilder) TextColumns(t []string) *SourceSQLServerBuilder {
	b.textColumns = t
	return b
}

func (b *SourceSQLServerBuilder) ExcludeColumns(e []string) *SourceSQLServerBuilder {
	b.excludeColumns = e
	return b
}

func (b *SourceSQLServerBuilder) Table(t []TableStruct) *SourceSQLServerBuilder {
	b.table = t
	return b
}

func (b *SourceSQLServerBuilder) ExposeProgress(e IdentifierSchemaStruct) *SourceSQLServerBuilder {
	b.exposeProgress = e
	return b
}

func (b *SourceSQLServerBuilder) SSLMode(sslMode string) *SourceSQLServerBuilder {
	b.sslMode = sslMode
	return b
}

func (b *SourceSQLServerBuilder) SSLCertificateAuthority(sslCertificateAuthority ValueSecretStruct) *SourceSQLServerBuilder {
	b.sslCertificateAuthority = sslCertificateAuthority
	return b
}

func (b *SourceSQLServerBuilder) AWSPrivateLink(awsPrivateLink IdentifierSchemaStruct) *SourceSQLServerBuilder {
	b.awsPrivateLink = awsPrivateLink
	return b
}

func (b *SourceSQLServerBuilder) Create() error {
	q := strings.Builder{}
	q.WriteString(fmt.Sprintf(`CREATE SOURCE %s`, b.QualifiedName()))

	if b.clusterName != "" {
		q.WriteString(fmt.Sprintf(` IN CLUSTER %s`, QuoteIdentifier(b.clusterName)))
	}

	q.WriteString(fmt.Sprintf(` FROM SQL SERVER CONNECTION %s`, b.sqlserverConnection.QualifiedName()))

	// Build options
	var options []string

	if len(b.textColumns) > 0 {
		s := strings.Join(b.textColumns, ", ")
		options = append(options, fmt.Sprintf(`TEXT COLUMNS (%s)`, s))
	}

	if len(b.excludeColumns) > 0 {
		s := strings.Join(b.excludeColumns, ", ")
		options = append(options, fmt.Sprintf(`EXCLUDE COLUMNS (%s)`, s))
	}

	if b.sslMode != "" {
		options = append(options, fmt.Sprintf(`SSL MODE %s`, QuoteString(b.sslMode)))
	}

	if b.sslCertificateAuthority.Text != "" {
		options = append(options, fmt.Sprintf(`SSL CERTIFICATE AUTHORITY %s`, QuoteString(b.sslCertificateAuthority.Text)))
	}
	if b.sslCertificateAuthority.Secret.Name != "" {
		options = append(options, fmt.Sprintf(`SSL CERTIFICATE AUTHORITY SECRET %s`, b.sslCertificateAuthority.Secret.QualifiedName()))
	}

	if b.awsPrivateLink.Name != "" {
		options = append(options, fmt.Sprintf(`AWS PRIVATELINK %s`, b.awsPrivateLink.QualifiedName()))
	}

	if len(options) > 0 {
		q.WriteString(fmt.Sprintf(` (%s)`, strings.Join(options, ", ")))
	}

	// Handle tables
	if len(b.table) > 0 {
		q.WriteString(` FOR TABLES (`)
		for i, t := range b.table {
			if t.UpstreamSchemaName == "" {
				t.UpstreamSchemaName = "dbo" // Default SQL Server schema
			}
			if t.Name == "" {
				t.Name = t.UpstreamName
			}
			if t.SchemaName == "" {
				t.SchemaName = b.SchemaName
			}
			if t.DatabaseName == "" {
				t.DatabaseName = b.DatabaseName
			}
			q.WriteString(fmt.Sprintf(`%s.%s AS %s.%s.%s`, QuoteIdentifier(t.UpstreamSchemaName), QuoteIdentifier(t.UpstreamName), QuoteIdentifier(t.DatabaseName), QuoteIdentifier(t.SchemaName), QuoteIdentifier(t.Name)))
			if i < len(b.table)-1 {
				q.WriteString(`, `)
			}
		}
		q.WriteString(`)`)
	} else {
		q.WriteString(` FOR ALL TABLES`)
	}

	if b.exposeProgress.Name != "" {
		q.WriteString(fmt.Sprintf(` EXPOSE PROGRESS AS %s`, b.exposeProgress.QualifiedName()))
	}

	q.WriteString(`;`)
	return b.ddl.exec(q.String())
}
