package materialize

import (
	"fmt"
	"strings"
)

type SecretBuilder struct {
	secretName   string
	schemaName   string
	databaseName string
}

func (b *SecretBuilder) qualifiedName() string {
	return QualifiedName(b.databaseName, b.schemaName, b.secretName)
}

func NewSecretBuilder(secretName, schemaName, databaseName string) *SecretBuilder {
	return &SecretBuilder{
		secretName:   secretName,
		schemaName:   schemaName,
		databaseName: databaseName,
	}
}

func (b *SecretBuilder) Create(value string) string {
	escapedValue := QuoteString(value)
	return fmt.Sprintf(`CREATE SECRET %s AS %s;`, b.qualifiedName(), escapedValue)
}

func (b *SecretBuilder) Rename(newName string) string {
	n := QualifiedName(b.databaseName, b.schemaName, newName)
	return fmt.Sprintf(`ALTER SECRET %s RENAME TO %s;`, b.qualifiedName(), n)
}

func (b *SecretBuilder) UpdateValue(newValue string) string {
	escapedValue := QuoteString(newValue)
	return fmt.Sprintf(`ALTER SECRET %s AS %s;`, b.qualifiedName(), escapedValue)
}

func (b *SecretBuilder) Drop() string {
	return fmt.Sprintf(`DROP SECRET %s;`, b.qualifiedName())
}

func (b *SecretBuilder) ReadId() string {
	return fmt.Sprintf(`
		SELECT mz_secrets.id
		FROM mz_secrets
		JOIN mz_schemas
			ON mz_secrets.schema_id = mz_schemas.id
		JOIN mz_databases
			ON mz_schemas.database_id = mz_databases.id
		WHERE mz_secrets.name = '%s'
		AND mz_schemas.name = '%s'
		AND mz_databases.name = '%s';`, b.secretName, b.schemaName, b.databaseName)
}

func ReadSecretParams(id string) string {
	return fmt.Sprintf(`
		SELECT
			mz_secrets.name AS name,
			mz_schemas.name AS schema_name,
			mz_databases.name AS database_name
		FROM mz_secrets
		JOIN mz_schemas
			ON mz_secrets.schema_id = mz_schemas.id
		JOIN mz_databases
			ON mz_schemas.database_id = mz_databases.id
		WHERE mz_secrets.id = '%s';`, id)
}

func ReadSecretDatasource(databaseName, schemaName string) string {
	q := strings.Builder{}
	q.WriteString(`
		SELECT
			mz_secrets.id,
			mz_secrets.name,
			mz_schemas.name,
			mz_databases.name
		FROM mz_secrets
		JOIN mz_schemas
			ON mz_secrets.schema_id = mz_schemas.id
		JOIN mz_databases
			ON mz_schemas.database_id = mz_databases.id`)

	if databaseName != "" {
		q.WriteString(fmt.Sprintf(`
		WHERE mz_databases.name = '%s'`, databaseName))

		if schemaName != "" {
			q.WriteString(fmt.Sprintf(` AND mz_schemas.name = '%s'`, schemaName))
		}
	}

	q.WriteString(`;`)
	return q.String()
}
