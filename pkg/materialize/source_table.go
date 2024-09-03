package materialize

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

type SourceTableParams struct {
	TableId            sql.NullString `db:"id"`
	TableName          sql.NullString `db:"name"`
	SchemaName         sql.NullString `db:"schema_name"`
	DatabaseName       sql.NullString `db:"database_name"`
	SourceName         sql.NullString `db:"source_name"`
	SourceSchemaName   sql.NullString `db:"source_schema_name"`
	SourceDatabaseName sql.NullString `db:"source_database_name"`
	UpstreamName       sql.NullString `db:"upstream_name"`
	UpstreamSchemaName sql.NullString `db:"upstream_schema_name"`
	TextColumns        pq.StringArray `db:"text_columns"`
	Comment            sql.NullString `db:"comment"`
	OwnerName          sql.NullString `db:"owner_name"`
	Privileges         pq.StringArray `db:"privileges"`
}

// TODO: Extend this query to include the upstream table name and schema name and the source
var sourceTableQuery = NewBaseQuery(`
	SELECT
		mz_tables.id,
		mz_tables.name,
		mz_schemas.name AS schema_name,
		mz_databases.name AS database_name,
		comments.comment AS comment,
		mz_roles.name AS owner_name,
		mz_tables.privileges
	FROM mz_tables
	JOIN mz_schemas
		ON mz_tables.schema_id = mz_schemas.id
	JOIN mz_databases
		ON mz_schemas.database_id = mz_databases.id
	JOIN mz_roles
		ON mz_tables.owner_id = mz_roles.id
	LEFT JOIN (
		SELECT id, comment
		FROM mz_internal.mz_comments
		WHERE object_type = 'table'
		AND object_sub_id IS NULL
	) comments
		ON mz_tables.id = comments.id
`)

func SourceTableId(conn *sqlx.DB, obj MaterializeObject) (string, error) {
	p := map[string]string{
		"mz_tables.name":    obj.Name,
		"mz_schemas.name":   obj.SchemaName,
		"mz_databases.name": obj.DatabaseName,
	}
	q := sourceTableQuery.QueryPredicate(p)

	var t SourceTableParams
	if err := conn.Get(&t, q); err != nil {
		return "", err
	}

	return t.TableId.String, nil
}

func ScanSourceTable(conn *sqlx.DB, id string) (SourceTableParams, error) {
	q := sourceTableQuery.QueryPredicate(map[string]string{"mz_tables.id": id})

	var t SourceTableParams
	if err := conn.Get(&t, q); err != nil {
		return t, err
	}

	return t, nil
}

type SourceTableBuilder struct {
	ddl                Builder
	tableName          string
	schemaName         string
	databaseName       string
	source             IdentifierSchemaStruct
	upstreamName       string
	upstreamSchemaName string
	textColumns        []string
}

func NewSourceTableBuilder(conn *sqlx.DB, obj MaterializeObject) *SourceTableBuilder {
	return &SourceTableBuilder{
		ddl:          Builder{conn, Table},
		tableName:    obj.Name,
		schemaName:   obj.SchemaName,
		databaseName: obj.DatabaseName,
	}
}

func (b *SourceTableBuilder) QualifiedName() string {
	return QualifiedName(b.databaseName, b.schemaName, b.tableName)
}

func (b *SourceTableBuilder) Source(s IdentifierSchemaStruct) *SourceTableBuilder {
	b.source = s
	return b
}

func (b *SourceTableBuilder) UpstreamName(n string) *SourceTableBuilder {
	b.upstreamName = n
	return b
}

func (b *SourceTableBuilder) UpstreamSchemaName(n string) *SourceTableBuilder {
	b.upstreamSchemaName = n
	return b
}

func (b *SourceTableBuilder) TextColumns(c []string) *SourceTableBuilder {
	b.textColumns = c
	return b
}

func (b *SourceTableBuilder) Create() error {
	q := strings.Builder{}
	q.WriteString(fmt.Sprintf(`CREATE TABLE %s`, b.QualifiedName()))
	q.WriteString(fmt.Sprintf(` FROM SOURCE %s`, b.source.QualifiedName()))
	q.WriteString(` (REFERENCE `)

	if b.upstreamSchemaName != "" {
		q.WriteString(fmt.Sprintf(`%s.`, QuoteIdentifier(b.upstreamSchemaName)))
	}
	q.WriteString(QuoteIdentifier(b.upstreamName))

	q.WriteString(")")

	if len(b.textColumns) > 0 {
		q.WriteString(fmt.Sprintf(` WITH (TEXT COLUMNS (%s))`, strings.Join(b.textColumns, ", ")))
	}

	q.WriteString(`;`)
	return b.ddl.exec(q.String())
}

func (b *SourceTableBuilder) Rename(newName string) error {
	oldName := b.QualifiedName()
	b.tableName = newName
	newName = b.QualifiedName()
	return b.ddl.rename(oldName, newName)
}

func (b *SourceTableBuilder) Drop() error {
	qn := b.QualifiedName()
	return b.ddl.drop(qn)
}
