package materialize

import (
	"fmt"
	"strings"
)

type ValueSecretStruct struct {
	Text   string
	Secret IdentifierSchemaStruct
}

func GetValueSecretStruct(databaseName string, schemaName string, v interface{}) ValueSecretStruct {
	var value ValueSecretStruct
	u := v.([]interface{})[0].(map[string]interface{})
	if v, ok := u["text"]; ok {
		value.Text = v.(string)
	}
	if v, ok := u["secret"]; ok && len(v.([]interface{})) > 0 {
		value.Secret = GetIdentifierSchemaStruct(databaseName, schemaName, v)
	}
	return value
}

type Connection struct {
	ConnectionName string
	SchemaName     string
	DatabaseName   string
}

func (c *Connection) QualifiedName() string {
	return QualifiedName(c.DatabaseName, c.SchemaName, c.ConnectionName)
}

func ReadConnectionId(name, schema, database string) string {
	return fmt.Sprintf(`
		SELECT mz_connections.id
		FROM mz_connections
		JOIN mz_schemas
			ON mz_connections.schema_id = mz_schemas.id
		JOIN mz_databases
			ON mz_schemas.database_id = mz_databases.id
		WHERE mz_connections.name = %s
		AND mz_schemas.name = %s
		AND mz_databases.name = %s;
	`, QuoteString(name), QuoteString(schema), QuoteString(database))
}

func ReadConnectionParams(id string) string {
	return fmt.Sprintf(`
		SELECT
			mz_connections.name,
			mz_schemas.name,
			mz_databases.name
		FROM mz_connections
		JOIN mz_schemas
			ON mz_connections.schema_id = mz_schemas.id
		JOIN mz_databases
			ON mz_schemas.database_id = mz_databases.id
		WHERE mz_connections.id = %s;`, QuoteString(id))
}

func ReadAwsPrivatelinkConnectionParams(id string) string {
	return fmt.Sprintf(`
		SELECT
			mz_connections.name,
			mz_schemas.name,
			mz_databases.name,
			mz_aws_privatelink_connections.principal
		FROM mz_connections
		JOIN mz_schemas
			ON mz_connections.schema_id = mz_schemas.id
		JOIN mz_databases
			ON mz_schemas.database_id = mz_databases.id
		JOIN mz_aws_privatelink_connections
			ON mz_connections.id = mz_aws_privatelink_connections.id
		WHERE mz_connections.id = %s;`, QuoteString(id))
}

func ReadConnectionDatasource(databaseName, schemaName string) string {
	q := strings.Builder{}
	q.WriteString(`
		SELECT
			mz_connections.id,
			mz_connections.name,
			mz_schemas.name AS schema_name,
			mz_databases.name AS database_name,
			mz_connections.type
		FROM mz_connections
		JOIN mz_schemas
			ON mz_connections.schema_id = mz_schemas.id
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
