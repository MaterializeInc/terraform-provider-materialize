package materialize

import (
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"
)

type SourceSqlServerBuilder struct {
	Source
	clusterName    string
	size           string
	connection     IdentifierSchemaStruct
	textColumns    []string
	ignoreColumns  []string
	tables         []TableStruct
	allTables      bool
	exposeProgress IdentifierSchemaStruct
}

func NewSourceSqlServerBuilder(conn *sqlx.DB, obj MaterializeObject) *SourceSqlServerBuilder {
	b := Builder{conn, BaseSource}
	return &SourceSqlServerBuilder{
		Source: Source{b, obj.Name, obj.SchemaName, obj.DatabaseName},
	}
}

func (b *SourceSqlServerBuilder) ClusterName(c string) *SourceSqlServerBuilder {
	b.clusterName = c
	return b
}

func (b *SourceSqlServerBuilder) SqlServerConnection(conn IdentifierSchemaStruct) *SourceSqlServerBuilder {
	b.connection = conn
	return b
}

func (b *SourceSqlServerBuilder) Tables(tables []TableStruct) *SourceSqlServerBuilder {
	b.tables = tables
	return b
}

func (b *SourceSqlServerBuilder) AllTables() *SourceSqlServerBuilder {
	b.allTables = true
	return b
}

func (b *SourceSqlServerBuilder) ExposeProgress(exposeProgress IdentifierSchemaStruct) *SourceSqlServerBuilder {
	b.exposeProgress = exposeProgress
	return b
}

func (b *SourceSqlServerBuilder) TextColumns(columns []string) *SourceSqlServerBuilder {
	b.textColumns = columns
	return b
}

func (b *SourceSqlServerBuilder) IgnoreColumns(columns []string) *SourceSqlServerBuilder {
	b.ignoreColumns = columns
	return b
}

func (b *SourceSqlServerBuilder) Create() error {
	q := strings.Builder{}
	q.WriteString(fmt.Sprintf(`CREATE SOURCE %s`, b.QualifiedName()))

	if b.clusterName != "" {
		q.WriteString(fmt.Sprintf(` IN CLUSTER %s`, QuoteIdentifier(b.clusterName)))
	}

	q.WriteString(fmt.Sprintf(` FROM SQL SERVER CONNECTION %s`, b.connection.QualifiedName()))

	// Handle column options - these come before FOR clause
	if len(b.textColumns) > 0 || len(b.ignoreColumns) > 0 {
		q.WriteString(" (")
		var columnOptions []string
		
		if len(b.textColumns) > 0 {
			s := strings.Join(b.textColumns, ", ")
			columnOptions = append(columnOptions, fmt.Sprintf(`TEXT COLUMNS (%s)`, s))
		}
		
		if len(b.ignoreColumns) > 0 {
			s := strings.Join(b.ignoreColumns, ", ")
			columnOptions = append(columnOptions, fmt.Sprintf(`EXCLUDE COLUMNS (%s)`, s))
		}
		
		q.WriteString(strings.Join(columnOptions, ", "))
		q.WriteString(")")
	}

	// Handle table specifications
	if len(b.tables) > 0 {
		var tableSpecs []string
		for _, table := range b.tables {
			spec := QuoteIdentifier(table.UpstreamSchemaName) + "." + QuoteIdentifier(table.UpstreamName)
			if table.Name != "" {
				spec += " AS " + QuoteIdentifier(table.Name)
			}
			tableSpecs = append(tableSpecs, spec)
		}
		q.WriteString(fmt.Sprintf(` FOR TABLES (%s)`, strings.Join(tableSpecs, ", ")))
	} else if b.allTables {
		q.WriteString(` FOR ALL TABLES`)
	}

	// Handle expose progress
	if b.exposeProgress.Name != "" {
		q.WriteString(fmt.Sprintf(` EXPOSE PROGRESS AS %s`, b.exposeProgress.QualifiedName()))
	}

	q.WriteString(`;`)
	return b.ddl.exec(q.String())
}

func (b *SourceSqlServerBuilder) Rename(newName string) error {
	n := b.QualifiedName()
	return b.ddl.rename(n, QualifiedName(b.DatabaseName, b.SchemaName, newName))
}

func (b *SourceSqlServerBuilder) Drop() error {
	qn := b.QualifiedName()
	return b.ddl.drop(qn)
}