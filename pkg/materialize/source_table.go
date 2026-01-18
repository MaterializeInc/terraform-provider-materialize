package materialize

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"
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
	UpstreamName       sql.NullString `db:"upstream_table_name"`
	UpstreamSchemaName sql.NullString `db:"upstream_schema_name"`
	Comment            sql.NullString `db:"comment"`
	OwnerName          sql.NullString `db:"owner_name"`
	Privileges         StringArray    `db:"privileges"`
}

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
        COALESCE(mz_kafka_source_tables.topic,
                 mz_mysql_source_tables.table_name,
                 mz_postgres_source_tables.table_name,
                 mz_sql_server_source_tables.table_name) AS upstream_table_name,
        COALESCE(mz_mysql_source_tables.schema_name,
                 mz_postgres_source_tables.schema_name,
                 mz_sql_server_source_tables.schema_name) AS upstream_schema_name,
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
    LEFT JOIN mz_internal.mz_kafka_source_tables
        ON mz_tables.id = mz_kafka_source_tables.id
    LEFT JOIN mz_internal.mz_mysql_source_tables
        ON mz_tables.id = mz_mysql_source_tables.id
    LEFT JOIN mz_internal.mz_postgres_source_tables
        ON mz_tables.id = mz_postgres_source_tables.id
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

// BaseCreate provides a template for the Create method
func (b *SourceTableBuilder) BaseCreate(sourceType string, additionalOptions func() string) error {
	q := strings.Builder{}
	q.WriteString(fmt.Sprintf(`CREATE TABLE %s`, b.QualifiedName()))
	q.WriteString(fmt.Sprintf(` FROM SOURCE %s`, b.source.QualifiedName()))

	// Reference is not required for Kafka sources and single-output load generator sources
	if b.upstreamName != "" {
		q.WriteString(` (REFERENCE `)

		if b.upstreamSchemaName != "" {
			q.WriteString(fmt.Sprintf(`%s.`, QuoteIdentifier(b.upstreamSchemaName)))
		}
		q.WriteString(QuoteIdentifier(b.upstreamName))

		q.WriteString(")")
	}

	if additionalOptions != nil {
		options := additionalOptions()
		if options != "" {
			q.WriteString(options)
		}
	}

	q.WriteString(`;`)
	return b.ddl.exec(q.String())
}

func ListSourceTables(conn *sqlx.DB, schemaName, databaseName string) ([]SourceTableParams, error) {
	p := map[string]string{
		"mz_schemas.name":   schemaName,
		"mz_databases.name": databaseName,
	}
	q := sourceTableQuery.QueryPredicate(p)

	var c []SourceTableParams
	if err := conn.Select(&c, q); err != nil {
		return c, err
	}

	return c, nil
}
