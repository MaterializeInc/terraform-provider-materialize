package materialize

import (
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"
)

type SourceMySQLBuilder struct {
	Source
	clusterName     string
	size            string
	mysqlConnection IdentifierSchemaStruct
	ignoreColumns   []string
	textColumns     []string
	tables          []TableStruct
	exposeProgress  IdentifierSchemaStruct
}

func NewSourceMySQLBuilder(conn *sqlx.DB, obj MaterializeObject) *SourceMySQLBuilder {
	b := Builder{conn, BaseSource}
	return &SourceMySQLBuilder{
		Source: Source{b, obj.Name, obj.SchemaName, obj.DatabaseName},
	}
}

func (b *SourceMySQLBuilder) ClusterName(c string) *SourceMySQLBuilder {
	b.clusterName = c
	return b
}

func (b *SourceMySQLBuilder) Size(s string) *SourceMySQLBuilder {
	b.size = s
	return b
}

func (b *SourceMySQLBuilder) MySQLConnection(mysqlConn IdentifierSchemaStruct) *SourceMySQLBuilder {
	b.mysqlConnection = mysqlConn
	return b
}

func (b *SourceMySQLBuilder) IgnoreColumns(i []string) *SourceMySQLBuilder {
	b.ignoreColumns = i
	return b
}

func (b *SourceMySQLBuilder) TextColumns(t []string) *SourceMySQLBuilder {
	b.textColumns = t
	return b
}

func (b *SourceMySQLBuilder) Tables(tables []TableStruct) *SourceMySQLBuilder {
	b.tables = tables
	return b
}

func (b *SourceMySQLBuilder) ExposeProgress(e IdentifierSchemaStruct) *SourceMySQLBuilder {
	b.exposeProgress = e
	return b
}

func (b *SourceMySQLBuilder) Create() error {
	q := strings.Builder{}

	q.WriteString(fmt.Sprintf(`CREATE SOURCE %s`, b.QualifiedName()))

	if b.clusterName != "" {
		q.WriteString(fmt.Sprintf(` IN CLUSTER %s`, QuoteIdentifier(b.clusterName)))
	}

	q.WriteString(fmt.Sprintf(` FROM MYSQL CONNECTION %s`, b.mysqlConnection.QualifiedName()))

	var options []string

	if len(b.ignoreColumns) > 0 {
		s := strings.Join(b.ignoreColumns, ", ")
		options = append(options, fmt.Sprintf(`IGNORE COLUMNS (%s)`, s))
	}

	if len(b.textColumns) > 0 {
		s := strings.Join(b.textColumns, ", ")
		options = append(options, fmt.Sprintf(`TEXT COLUMNS (%s)`, s))
	}

	if len(options) > 0 {
		q.WriteString(" (")
		q.WriteString(strings.Join(options, ", "))
		q.WriteString(")")
	}

	if len(b.tables) > 0 {
		q.WriteString(` FOR TABLES (`)
		for i, t := range b.tables {
			if t.SchemaName == "" {
				t.SchemaName = b.SchemaName
			}
			if t.Alias == "" {
				t.Alias = t.Name
			}
			if t.AliasSchemaName == "" {
				t.AliasSchemaName = b.SchemaName
			}
			q.WriteString(fmt.Sprintf(`%s.%s AS %s.%s.%s`, QuoteIdentifier(t.SchemaName), QuoteIdentifier(t.Name), QuoteIdentifier(b.DatabaseName), QuoteIdentifier(t.AliasSchemaName), QuoteIdentifier(t.Alias)))
			if i < len(b.tables)-1 {
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
