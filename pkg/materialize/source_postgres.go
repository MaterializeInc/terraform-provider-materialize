package materialize

import (
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"
)

type SourcePostgresBuilder struct {
	Source
	clusterName        string
	size               string
	postgresConnection IdentifierSchemaStruct
	publication        string
	textColumns        []string
	table              []TableStruct
	exposeProgress     IdentifierSchemaStruct
}

func NewSourcePostgresBuilder(conn *sqlx.DB, obj MaterializeObject) *SourcePostgresBuilder {
	b := Builder{conn, BaseSource}
	return &SourcePostgresBuilder{
		Source: Source{b, obj.Name, obj.SchemaName, obj.DatabaseName},
	}
}

func (b *SourcePostgresBuilder) ClusterName(c string) *SourcePostgresBuilder {
	b.clusterName = c
	return b
}

func (b *SourcePostgresBuilder) Size(s string) *SourcePostgresBuilder {
	b.size = s
	return b
}

func (b *SourcePostgresBuilder) PostgresConnection(p IdentifierSchemaStruct) *SourcePostgresBuilder {
	b.postgresConnection = p
	return b
}

func (b *SourcePostgresBuilder) Publication(p string) *SourcePostgresBuilder {
	b.publication = p
	return b
}

func (b *SourcePostgresBuilder) TextColumns(t []string) *SourcePostgresBuilder {
	b.textColumns = t
	return b
}

func (b *SourcePostgresBuilder) Table(t []TableStruct) *SourcePostgresBuilder {
	b.table = t
	return b
}

func (b *SourcePostgresBuilder) ExposeProgress(e IdentifierSchemaStruct) *SourcePostgresBuilder {
	b.exposeProgress = e
	return b
}

func (b *SourcePostgresBuilder) Create() error {
	q := strings.Builder{}
	q.WriteString(fmt.Sprintf(`CREATE SOURCE %s`, b.QualifiedName()))

	if b.clusterName != "" {
		q.WriteString(fmt.Sprintf(` IN CLUSTER %s`, QuoteIdentifier(b.clusterName)))
	}

	q.WriteString(fmt.Sprintf(` FROM POSTGRES CONNECTION %s`, b.postgresConnection.QualifiedName()))

	// Publication
	p := fmt.Sprintf(`PUBLICATION %s`, QuoteString(b.publication))

	if len(b.textColumns) > 0 {
		s := strings.Join(b.textColumns, ", ")
		p = p + fmt.Sprintf(`, TEXT COLUMNS (%s)`, s)
	}

	q.WriteString(fmt.Sprintf(` (%s)`, p))

	if b.table != nil && len(b.table) > 0 {
		q.WriteString(` FOR TABLES (`)
		for i, t := range b.table {
			if t.UpstreamSchemaName == "" {
				t.UpstreamSchemaName = b.SchemaName
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
	}

	if b.exposeProgress.Name != "" {
		q.WriteString(fmt.Sprintf(` EXPOSE PROGRESS AS %s`, b.exposeProgress.QualifiedName()))
	}

	q.WriteString(`;`)
	return b.ddl.exec(q.String())
}
