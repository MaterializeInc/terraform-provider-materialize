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
	SourceType         sql.NullString `db:"source_type"`
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
		mz_sources.name AS source_name,
		source_schemas.name AS source_schema_name,
		source_databases.name AS source_database_name,
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
	ignoreColumns      []string
	sourceType         string
	conn               *sqlx.DB
}

func NewSourceTableBuilder(conn *sqlx.DB, obj MaterializeObject) *SourceTableBuilder {
	return &SourceTableBuilder{
		ddl:          Builder{conn, Table},
		tableName:    obj.Name,
		schemaName:   obj.SchemaName,
		databaseName: obj.DatabaseName,
		conn:         conn,
	}
}

func (b *SourceTableBuilder) GetSourceType() (string, error) {
	if b.sourceType != "" {
		return b.sourceType, nil
	}

	q := sourceQuery.QueryPredicate(map[string]string{
		"mz_sources.name":   b.source.Name,
		"mz_schemas.name":   b.source.SchemaName,
		"mz_databases.name": b.source.DatabaseName,
	})

	var s SourceParams
	if err := b.conn.Get(&s, q); err != nil {
		return "", err
	}

	b.sourceType = s.SourceType.String

	return b.sourceType, nil
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

func (b *SourceTableBuilder) IgnoreColumns(c []string) *SourceTableBuilder {
	b.ignoreColumns = c
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

	var options []string

	if len(b.textColumns) > 0 {
		s := strings.Join(b.textColumns, ", ")
		options = append(options, fmt.Sprintf(`TEXT COLUMNS (%s)`, s))
	}

	sourceType, err := b.GetSourceType()
	if err != nil {
		return err
	}

	if strings.EqualFold(sourceType, "mysql") && len(b.ignoreColumns) > 0 {
		s := strings.Join(b.ignoreColumns, ", ")
		options = append(options, fmt.Sprintf(`IGNORE COLUMNS (%s)`, s))
	}

	if len(options) > 0 {
		q.WriteString(" WITH (")
		q.WriteString(strings.Join(options, ", "))
		q.WriteString(")")
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
