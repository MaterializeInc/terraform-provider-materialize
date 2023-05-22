package materialize

import (
	"fmt"
	"strings"
)

type SchemaBuilder struct {
	schemaName   string
	databaseName string
}

func (b *SchemaBuilder) QualifiedName() string {
	return QualifiedName(b.databaseName, b.schemaName)
}

func NewSchemaBuilder(schemaName, databaseName string) *SchemaBuilder {
	return &SchemaBuilder{
		schemaName:   schemaName,
		databaseName: databaseName,
	}
}

func (b *SchemaBuilder) Create() string {
	return fmt.Sprintf(`CREATE SCHEMA %s;`, b.QualifiedName())
}

func (b *SchemaBuilder) Drop() string {
	return fmt.Sprintf(`DROP SCHEMA %s;`, b.QualifiedName())
}

func (b *SchemaBuilder) ReadId() string {
	return fmt.Sprintf(`
		SELECT mz_schemas.id
		FROM mz_schemas JOIN mz_databases
			ON mz_schemas.database_id = mz_databases.id
		WHERE mz_schemas.name = %s
		AND mz_databases.name = %s;
	`, QuoteString(b.schemaName), QuoteString(b.databaseName))
}

func ReadSchemaParams(id string) string {
	return fmt.Sprintf(`
		SELECT
			mz_schemas.name AS schema_name,
			mz_databases.name AS database_name
		FROM mz_schemas JOIN mz_databases
			ON mz_schemas.database_id = mz_databases.id
		WHERE mz_schemas.id = %s;`, QuoteString(id))
}

func ReadSchemaDatasource(databaseName string) string {
	q := strings.Builder{}
	q.WriteString(`
		SELECT
			mz_schemas.id,
			mz_schemas.name,
			mz_databases.name
		FROM mz_schemas JOIN mz_databases
			ON mz_schemas.database_id = mz_databases.id
	`)

	if databaseName != "" {
		q.WriteString(fmt.Sprintf(`WHERE mz_databases.name = %s`, QuoteString(databaseName)))
	}

	q.WriteString(`;`)
	return q.String()
}
