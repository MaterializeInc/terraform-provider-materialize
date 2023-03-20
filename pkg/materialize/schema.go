package materialize

import (
	"fmt"
)

type SchemaBuilder struct {
	schemaName   string
	databaseName string
}

func (b *SchemaBuilder) qualifiedName() string {
	return QualifiedName(b.databaseName, b.schemaName)
}

func NewSchemaBuilder(schemaName, databaseName string) *SchemaBuilder {
	return &SchemaBuilder{
		schemaName:   schemaName,
		databaseName: databaseName,
	}
}

func (b *SchemaBuilder) Create() string {
	return fmt.Sprintf(`CREATE SCHEMA %s;`, b.qualifiedName())
}

func (b *SchemaBuilder) Drop() string {
	return fmt.Sprintf(`DROP SCHEMA %s;`, b.qualifiedName())
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
			mz_schemas.name,
			mz_databases.name
		FROM mz_schemas JOIN mz_databases
			ON mz_schemas.database_id = mz_databases.id
		WHERE mz_schemas.id = %s;`, QuoteString(id))
}
