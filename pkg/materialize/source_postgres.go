package materialize

import (
	"fmt"
	"strings"
)

type TablePostgres struct {
	Name  string
	Alias string
}

type SourcePostgresBuilder struct {
	sourceName         string
	schemaName         string
	databaseName       string
	clusterName        string
	size               string
	postgresConnection IdentifierSchemaStruct
	publication        string
	textColumns        []string
	tables             []TablePostgres
}

func (b *SourcePostgresBuilder) qualifiedName() string {
	return QualifiedName(b.databaseName, b.schemaName, b.sourceName)
}

func NewSourcePostgresBuilder(sourceName, schemaName, databaseName string) *SourcePostgresBuilder {
	return &SourcePostgresBuilder{
		sourceName:   sourceName,
		schemaName:   schemaName,
		databaseName: databaseName,
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

func (b *SourcePostgresBuilder) Tables(t []TablePostgres) *SourcePostgresBuilder {
	b.tables = t
	return b
}

func (b *SourcePostgresBuilder) Create() string {
	q := strings.Builder{}
	q.WriteString(fmt.Sprintf(`CREATE SOURCE %s`, b.qualifiedName()))

	if b.clusterName != "" {
		q.WriteString(fmt.Sprintf(` IN CLUSTER %s`, QuoteIdentifier(b.clusterName)))
	}

	q.WriteString(fmt.Sprintf(` FROM POSTGRES CONNECTION %s`, QualifiedName(b.postgresConnection.DatabaseName, b.postgresConnection.SchemaName, b.postgresConnection.Name)))

	// Publication
	p := fmt.Sprintf(`PUBLICATION %s`, QuoteString(b.publication))

	if len(b.textColumns) > 0 {
		s := strings.Join(b.textColumns, ", ")
		p = p + fmt.Sprintf(`, TEXT COLUMNS (%s)`, s)
	}

	q.WriteString(fmt.Sprintf(` (%s)`, p))

	if len(b.tables) > 0 {
		q.WriteString(` FOR TABLES (`)
		for i, t := range b.tables {
			if t.Alias == "" {
				t.Alias = t.Name
			}
			q.WriteString(fmt.Sprintf(`%s AS %s`, t.Name, t.Alias))
			if i < len(b.tables)-1 {
				q.WriteString(`, `)
			}
		}
		q.WriteString(`)`)
	} else {
		q.WriteString(` FOR ALL TABLES`)
	}

	if b.size != "" {
		q.WriteString(fmt.Sprintf(` WITH (SIZE = %s)`, QuoteString(b.size)))
	}

	q.WriteString(`;`)
	return q.String()
}

func (b *SourcePostgresBuilder) Rename(newName string) string {
	n := QualifiedName(b.databaseName, b.schemaName, newName)
	return fmt.Sprintf(`ALTER SOURCE %s RENAME TO %s;`, b.qualifiedName(), n)
}

func (b *SourcePostgresBuilder) UpdateSize(newSize string) string {
	return fmt.Sprintf(`ALTER SOURCE %s SET (SIZE = %s);`, b.qualifiedName(), QuoteString(newSize))
}

func (b *SourcePostgresBuilder) Drop() string {
	return fmt.Sprintf(`DROP SOURCE %s;`, b.qualifiedName())
}

func (b *SourcePostgresBuilder) ReadId() string {
	return ReadSourceId(b.sourceName, b.schemaName, b.databaseName)
}
