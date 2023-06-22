package materialize

import (
	"database/sql"

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
	ddl            Builder
	ConnectionName string
	SchemaName     string
	DatabaseName   string
}

func NewConnection(conn *sqlx.DB, obj ObjectSchemaStruct) *Connection {
	return &Connection{
		ddl:            Builder{conn, BaseConnection},
		ConnectionName: obj.Name,
		SchemaName:     obj.SchemaName,
		DatabaseName:   obj.DatabaseName,
	}
}

func (c *Connection) QualifiedName() string {
	return QualifiedName(c.DatabaseName, c.SchemaName, c.ConnectionName)
}

func (b *Connection) Rename(newConnectionName string) error {
	n := QualifiedName(newConnectionName)
	return b.ddl.rename(b.QualifiedName(), n)
}

func (b *Connection) Drop() error {
	qn := b.QualifiedName()
	return b.ddl.drop(qn)
}

type ConnectionParams struct {
	ConnectionId   sql.NullString `db:"id"`
	ConnectionName sql.NullString `db:"connection_name"`
	SchemaName     sql.NullString `db:"schema_name"`
	DatabaseName   sql.NullString `db:"database_name"`
	ConnectionType sql.NullString `db:"connection_type"`
	OwnerName      sql.NullString `db:"owner_name"`
	Privileges     sql.NullString `db:"privileges"`
}

var connectionQuery = NewBaseQuery(`
	SELECT
		mz_connections.id,
		mz_connections.name AS connection_name,
		mz_schemas.name AS schema_name,
		mz_databases.name AS database_name,
		mz_connections.type AS connection_type,
		mz_roles.name AS owner_name,
		mz_connections.privileges
	FROM mz_connections
	JOIN mz_schemas
		ON mz_connections.schema_id = mz_schemas.id
	JOIN mz_databases
		ON mz_schemas.database_id = mz_databases.id
	JOIN mz_roles
		ON mz_connections.owner_id = mz_roles.id`)

func ConnectionId(conn *sqlx.DB, obj ObjectSchemaStruct) (string, error) {
	p := map[string]string{
		"mz_connections.name": obj.Name,
		"mz_databases.name":   obj.DatabaseName,
		"mz_schemas.name":     obj.SchemaName,
	}
	q := connectionQuery.QueryPredicate(p)

	var c ConnectionParams
	if err := conn.Get(&c, q); err != nil {
		return "", err
	}

	return c.ConnectionId.String, nil
}

func ScanConnection(conn *sqlx.DB, id string) (ConnectionParams, error) {
	q := connectionQuery.QueryPredicate(map[string]string{"mz_connections.id": id})

	var c ConnectionParams
	if err := conn.Get(&c, q); err != nil {
		return c, err
	}

	return c, nil
}

func ListConnections(conn *sqlx.DB, schemaName, databaseName string) ([]ConnectionParams, error) {
	p := map[string]string{
		"mz_schemas.name":   schemaName,
		"mz_databases.name": databaseName,
	}
	q := connectionQuery.QueryPredicate(p)

	var c []ConnectionParams
	if err := conn.Select(&c, q); err != nil {
		return c, err
	}

	return c, nil
}
