package materialize

import (
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"
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
	conn           *sqlx.DB
	ConnectionName string
	SchemaName     string
	DatabaseName   string
}

func NewConnection(conn *sqlx.DB, name, schema, database string) *Connection {
	return &Connection{
		conn:           conn,
		ConnectionName: name,
		SchemaName:     schema,
		DatabaseName:   database,
	}
}

func (c *Connection) QualifiedName() string {
	return QualifiedName(c.DatabaseName, c.SchemaName, c.ConnectionName)
}

func (b *Connection) Rename(newName string) error {
	n := QualifiedName(b.DatabaseName, b.SchemaName, newName)
	q := fmt.Sprintf(`ALTER CONNECTION %s RENAME TO %s;`, b.QualifiedName(), n)

	_, err := b.conn.Exec(q)
	if err != nil {
		return err
	}

	return nil
}

func (b *Connection) Drop() error {
	q := fmt.Sprintf(`DROP CONNECTION %s;`, b.QualifiedName())

	_, err := b.conn.Exec(q)
	if err != nil {
		return err
	}

	return nil
}

func (b *Connection) ReadId() (string, error) {
	q := fmt.Sprintf(`
		SELECT mz_connections.id
		FROM mz_connections
		JOIN mz_schemas
			ON mz_connections.schema_id = mz_schemas.id
		JOIN mz_databases
			ON mz_schemas.database_id = mz_databases.id
		WHERE mz_connections.name = %s
		AND mz_schemas.name = %s
		AND mz_databases.name = %s;
	`, QuoteString(b.ConnectionName), QuoteString(b.SchemaName), QuoteString(b.DatabaseName))

	var i string
	if err := b.conn.QueryRowx(q).Scan(&i); err != nil {
		return "", err
	}

	return i, nil
}

type ConnectionParams struct {
	ConnectionName string `db:"name"`
	SchemaName     string `db:"schema"`
	DatabaseName   string `db:"database"`
}

func (b *Connection) Params(catalogId string) (ConnectionParams, error) {
	q := fmt.Sprintf(`
		SELECT
			mz_connections.name,
			mz_schemas.name,
			mz_databases.name
		FROM mz_connections
		JOIN mz_schemas
			ON mz_connections.schema_id = mz_schemas.id
		JOIN mz_databases
			ON mz_schemas.database_id = mz_databases.id
		WHERE mz_connections.id = %s;
	`, QuoteString(catalogId))

	var s ConnectionParams
	if err := b.conn.Get(&s, q); err != nil {
		return s, err
	}

	return s, nil
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
