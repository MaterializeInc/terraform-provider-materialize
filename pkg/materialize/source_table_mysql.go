package materialize

import (
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"
)

// MySQL specific params and query
type SourceTableMySQLParams struct {
	SourceTableParams
	ExcludeColumns StringArray `db:"exclude_columns"`
	TextColumns    StringArray `db:"text_columns"`
}

var sourceTableMySQLQuery = `
	SELECT
		mz_tables.id,
		mz_tables.name,
		mz_schemas.name AS schema_name,
		mz_databases.name AS database_name,
		mz_sources.name AS source_name,
		source_schemas.name AS source_schema_name,
		source_databases.name AS source_database_name,
		mz_mysql_source_tables.table_name AS upstream_table_name,
		mz_mysql_source_tables.schema_name AS upstream_schema_name,
		mz_sources.type AS source_type,
		comments.comment AS comment,
		mz_roles.name AS owner_name,
		mz_tables.privileges
	FROM mz_tables
	JOIN mz_schemas
		ON mz_tables.schema_id = mz_schemas.id
	JOIN mz_databases
		ON mz_schemas.database_id = mz_databases.id
	JOIN mz_sources
		ON mz_tables.source_id = mz_sources.id
	JOIN mz_schemas AS source_schemas
		ON mz_sources.schema_id = source_schemas.id
	JOIN mz_databases AS source_databases
		ON source_schemas.database_id = source_databases.id
	LEFT JOIN mz_internal.mz_mysql_source_tables
		ON mz_tables.id = mz_mysql_source_tables.id
	JOIN mz_roles
		ON mz_tables.owner_id = mz_roles.id
	LEFT JOIN (
		SELECT id, comment
		FROM mz_internal.mz_comments
		WHERE object_type = 'table'
		AND object_sub_id IS NULL
	) comments
		ON mz_tables.id = comments.id
`

func SourceTableMySQLId(conn *sqlx.DB, obj MaterializeObject) (string, error) {
	p := map[string]string{
		"mz_tables.name":    obj.Name,
		"mz_schemas.name":   obj.SchemaName,
		"mz_databases.name": obj.DatabaseName,
	}
	q := NewBaseQuery(sourceTableMySQLQuery).QueryPredicate(p)

	var t SourceTableParams
	if err := conn.Get(&t, q); err != nil {
		return "", err
	}

	return t.TableId.String, nil
}

func ScanSourceTableMySQL(conn *sqlx.DB, id string) (SourceTableMySQLParams, error) {
	q := NewBaseQuery(sourceTableMySQLQuery).QueryPredicate(map[string]string{"mz_tables.id": id})

	var params SourceTableMySQLParams
	if err := conn.Get(&params, q); err != nil {
		return params, err
	}

	return params, nil
}

// SourceTableMySQLBuilder for MySQL sources
type SourceTableMySQLBuilder struct {
	*SourceTableBuilder
	textColumns    []string
	excludeColumns []string
}

func NewSourceTableMySQLBuilder(conn *sqlx.DB, obj MaterializeObject) *SourceTableMySQLBuilder {
	return &SourceTableMySQLBuilder{
		SourceTableBuilder: NewSourceTableBuilder(conn, obj),
	}
}

func (b *SourceTableMySQLBuilder) TextColumns(c []string) *SourceTableMySQLBuilder {
	b.textColumns = c
	return b
}

func (b *SourceTableMySQLBuilder) ExcludeColumns(c []string) *SourceTableMySQLBuilder {
	b.excludeColumns = c
	return b
}

func (b *SourceTableMySQLBuilder) Create() error {
	return b.BaseCreate("mysql", func() string {
		q := strings.Builder{}
		var options []string
		if len(b.textColumns) > 0 {
			var quotedCols []string
			for _, col := range b.textColumns {
				quotedCols = append(quotedCols, QuoteIdentifier(col))
			}
			s := strings.Join(quotedCols, ", ")
			options = append(options, fmt.Sprintf(`TEXT COLUMNS (%s)`, s))
		}

		if len(b.excludeColumns) > 0 {
			var quotedCols []string
			for _, col := range b.excludeColumns {
				quotedCols = append(quotedCols, QuoteIdentifier(col))
			}
			s := strings.Join(quotedCols, ", ")
			options = append(options, fmt.Sprintf(`EXCLUDE COLUMNS (%s)`, s))
		}

		if len(options) > 0 {
			q.WriteString(" WITH (")
			q.WriteString(strings.Join(options, ", "))
			q.WriteString(")")
		}

		return q.String()
	})
}
