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
	schema             []string
	exposeProgress     string
}

func NewSourcePostgresBuilder(conn *sqlx.DB, obj ObjectSchemaStruct) *SourcePostgresBuilder {
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

func (b *SourcePostgresBuilder) Schema(t []string) *SourcePostgresBuilder {
	b.schema = t
	return b
}

func (b *SourcePostgresBuilder) ExposeProgress(e string) *SourcePostgresBuilder {
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

	if len(b.table) > 0 {
		q.WriteString(` FOR TABLES (`)
		for i, t := range b.table {
			if t.Alias == "" {
				t.Alias = t.Name
			}
			q.WriteString(fmt.Sprintf(`%s AS %s`, t.Name, t.Alias))
			if i < len(b.table)-1 {
				q.WriteString(`, `)
			}
		}
		q.WriteString(`)`)
	} else if len(b.schema) > 0 {
		s := strings.Join(b.schema, ", ")
		q.WriteString(fmt.Sprintf(` FOR SCHEMAS (%s)`, s))
	} else {
		q.WriteString(` FOR ALL TABLES`)
	}

	if b.exposeProgress != "" {
		q.WriteString(fmt.Sprintf(` EXPOSE PROGRESS AS %s`, b.exposeProgress))
	}

	if b.size != "" {
		q.WriteString(fmt.Sprintf(` WITH (SIZE = %s)`, QuoteString(b.size)))
	}

	q.WriteString(`;`)
	return b.ddl.exec(q.String())
}
