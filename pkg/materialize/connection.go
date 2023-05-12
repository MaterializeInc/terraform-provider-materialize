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
	ConnectionName string
	SchemaName     string
	DatabaseName   string
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

	var name, schema, database string
	if err := b.conn.QueryRowx(q).Scan(name, schema, database); err != nil {
		return ConnectionParams{}, err
	}

	return ConnectionParams{
		ConnectionName: name,
		SchemaName:     schema,
		DatabaseName:   database,
	}, nil
}

type ConnectionAwsPrivatelinkParams struct {
	ConnectionName string
	SchemaName     string
	DatabaseName   string
	Principal      string
}

func (b *Connection) AwsPrivatelinkParams(catalogId string) (ConnectionAwsPrivatelinkParams, error) {
	q := fmt.Sprintf(`
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
		WHERE mz_connections.id = %s;
	`, QuoteString(catalogId))

	var name, schema, database, principal string
	if err := b.conn.QueryRowx(q).Scan(name, schema, database, principal); err != nil {
		return ConnectionAwsPrivatelinkParams{}, err
	}

	return ConnectionAwsPrivatelinkParams{
		ConnectionName: name,
		SchemaName:     schema,
		DatabaseName:   database,
		Principal:      principal,
	}, nil
}

type ConnectionSshTunnelParams struct {
	ConnectionName string
	SchemaName     string
	DatabaseName   string
	PublicKey1     string
	PublicKey2     string
}

func (b *Connection) SshTunnelParams(catalogId string) (ConnectionSshTunnelParams, error) {
	q := fmt.Sprintf(`
		SELECT
			mz_connections.name,
			mz_schemas.name,
			mz_databases.name,
			mz_ssh_tunnel_connections.public_key_1,
			mz_ssh_tunnel_connections.public_key_2
		FROM mz_connections
		JOIN mz_schemas
			ON mz_connections.schema_id = mz_schemas.id
		JOIN mz_databases
			ON mz_schemas.database_id = mz_databases.id
		LEFT JOIN mz_ssh_tunnel_connections
			ON mz_connections.id = mz_ssh_tunnel_connections.id
		WHERE mz_connections.id = %s;
	`, QuoteString(catalogId))

	var name, schema, database, publick1, publick2 string
	if err := b.conn.QueryRowx(q).Scan(name, schema, database, publick1, publick2); err != nil {
		return ConnectionSshTunnelParams{}, err
	}

	return ConnectionSshTunnelParams{
		ConnectionName: name,
		SchemaName:     schema,
		DatabaseName:   database,
		PublicKey1:     publick1,
		PublicKey2:     publick2,
	}, nil
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
