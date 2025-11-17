package materialize

import (
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

// SQL Server specific params and query
type SourceTableSQLServerParams struct {
	SourceTableParams
	ExcludeColumns pq.StringArray `db:"exclude_columns"`
	TextColumns    pq.StringArray `db:"text_columns"`
}

var sourceTableSQLServerQuery = `
	SELECT
		mz_tables.id,
		mz_tables.name,
		mz_schemas.name AS schema_name,
		mz_databases.name AS database_name,
		mz_sources.name AS source_name,
		source_schemas.name AS source_schema_name,
		source_databases.name AS source_database_name,
		mz_sql_server_source_tables.table_name AS upstream_table_name,
		mz_sql_server_source_tables.schema_name AS upstream_schema_name,
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
	LEFT JOIN mz_internal.mz_sql_server_source_tables
		ON mz_tables.id = mz_sql_server_source_tables.id
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

func SourceTableSQLServerId(conn *sqlx.DB, obj MaterializeObject) (string, error) {
	p := map[string]string{
		"mz_tables.name":    obj.Name,
		"mz_schemas.name":   obj.SchemaName,
		"mz_databases.name": obj.DatabaseName,
	}
	q := NewBaseQuery(sourceTableSQLServerQuery).QueryPredicate(p)

	var t SourceTableParams
	if err := conn.Get(&t, q); err != nil {
		return "", err
	}

	return t.TableId.String, nil
}

func ScanSourceTableSQLServer(conn *sqlx.DB, id string) (SourceTableSQLServerParams, error) {
	q := NewBaseQuery(sourceTableSQLServerQuery).QueryPredicate(map[string]string{"mz_tables.id": id})

	var params SourceTableSQLServerParams
	if err := conn.Get(&params, q); err != nil {
		return params, err
	}

	return params, nil
}

// SourceTableSQLServerBuilder for SQL Server sources
type SourceTableSQLServerBuilder struct {
	*SourceTableBuilder
	textColumns    []ColumnReferenceStruct
	excludeColumns []ColumnReferenceStruct
}

func NewSourceTableSQLServerBuilder(conn *sqlx.DB, obj MaterializeObject) *SourceTableSQLServerBuilder {
	return &SourceTableSQLServerBuilder{
		SourceTableBuilder: NewSourceTableBuilder(conn, obj),
	}
}

func (b *SourceTableSQLServerBuilder) TextColumns(c []ColumnReferenceStruct) *SourceTableSQLServerBuilder {
	b.textColumns = c
	return b
}

func (b *SourceTableSQLServerBuilder) ExcludeColumns(c []ColumnReferenceStruct) *SourceTableSQLServerBuilder {
	b.excludeColumns = c
	return b
}

func (b *SourceTableSQLServerBuilder) Create() error {
	return b.BaseCreate("sqlserver", func() string {
		q := strings.Builder{}
		var options []string
		if len(b.textColumns) > 0 {
			var quotedCols []string
			for _, col := range b.textColumns {
				quotedCols = append(quotedCols, QuoteIdentifier(col.ColumnName))
			}
			s := strings.Join(quotedCols, ", ")
			options = append(options, fmt.Sprintf(`TEXT COLUMNS (%s)`, s))
		}

		if len(b.excludeColumns) > 0 {
			var quotedCols []string
			for _, col := range b.excludeColumns {
				quotedCols = append(quotedCols, QuoteIdentifier(col.ColumnName))
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
