package materialize

import (
	"fmt"
	"strings"
)

type TableColumn struct {
	ColName string
	ColType string
	NotNull bool
}

func GetTableColumnStruct(v []interface{}) []TableColumn {
	var columns []TableColumn
	for _, column := range v {
		c := column.(map[string]interface{})
		columns = append(columns, TableColumn{
			ColName: c["name"].(string),
			ColType: c["type"].(string),
			NotNull: c["nullable"].(bool),
		})
	}
	return columns
}

type TableBuilder struct {
	tableName    string
	schemaName   string
	databaseName string
	column       []TableColumn
}

func (b *TableBuilder) QualifiedName() string {
	return QualifiedName(b.databaseName, b.schemaName, b.tableName)
}

func NewTableBuilder(tableName, schemaName, databaseName string) *TableBuilder {
	return &TableBuilder{
		tableName:    tableName,
		schemaName:   schemaName,
		databaseName: databaseName,
	}
}

func (b *TableBuilder) Column(c []TableColumn) *TableBuilder {
	b.column = c
	return b
}

func (b *TableBuilder) Create() string {
	q := strings.Builder{}
	q.WriteString(fmt.Sprintf(`CREATE TABLE %s`, b.QualifiedName()))

	var column []string
	for _, c := range b.column {
		s := strings.Builder{}

		s.WriteString(fmt.Sprintf(`%s %s`, c.ColName, c.ColType))
		if c.NotNull {
			s.WriteString(` NOT NULL`)
		}
		o := s.String()
		column = append(column, o)

	}
	p := strings.Join(column[:], ", ")
	q.WriteString(fmt.Sprintf(` (%s);`, p))
	return q.String()
}

func (b *TableBuilder) Rename(newName string) string {
	n := QualifiedName(b.databaseName, b.schemaName, newName)
	return fmt.Sprintf(`ALTER TABLE %s RENAME TO %s;`, b.QualifiedName(), n)
}

func (b *TableBuilder) Drop() string {
	return fmt.Sprintf(`DROP TABLE %s;`, b.QualifiedName())
}

func (b *TableBuilder) ReadId() string {
	return fmt.Sprintf(`
		SELECT mz_tables.id
		FROM mz_tables
		JOIN mz_schemas
			ON mz_tables.schema_id = mz_schemas.id
		JOIN mz_databases
			ON mz_schemas.database_id = mz_databases.id
		WHERE mz_tables.name = %s
		AND mz_schemas.name = %s
		AND mz_databases.name = %s;
	`, QuoteString(b.tableName), QuoteString(b.schemaName), QuoteString(b.databaseName))
}

func ReadTableParams(id string) string {
	return fmt.Sprintf(`
		SELECT
			mz_tables.name AS table_name,
			mz_schemas.name AS schema_name,
			mz_databases.name AS database_name
		FROM mz_tables
		JOIN mz_schemas
			ON mz_tables.schema_id = mz_schemas.id
		JOIN mz_databases
			ON mz_schemas.database_id = mz_databases.id
		WHERE mz_tables.id = %s;`, QuoteString(id))
}

func ReadTableDatasource(databaseName, schemaName string) string {
	q := strings.Builder{}
	q.WriteString(`
		SELECT
			mz_tables.id,
			mz_tables.name,
			mz_schemas.name,
			mz_databases.name
		FROM mz_tables
		JOIN mz_schemas
			ON mz_tables.schema_id = mz_schemas.id
		JOIN mz_databases
			ON mz_schemas.database_id = mz_databases.id`)

	if databaseName != "" {
		q.WriteString(fmt.Sprintf(`
		WHERE mz_databases.name = %s`, QuoteString(databaseName)))

		if schemaName != "" {
			q.WriteString(fmt.Sprintf(` AND mz_schemas.name = %s`, QuoteString(schemaName)))
		}
	}

	q.WriteString(`;`)
	return q.String()
}
