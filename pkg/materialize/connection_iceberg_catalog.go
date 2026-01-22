package materialize

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"
)

type ConnectionIcebergCatalogBuilder struct {
	Connection
	catalogType   string
	url           string
	warehouse     string
	awsConnection IdentifierSchemaStruct
	validate      bool
}

func NewConnectionIcebergCatalogBuilder(conn *sqlx.DB, obj MaterializeObject) *ConnectionIcebergCatalogBuilder {
	b := Builder{conn, BaseConnection}
	return &ConnectionIcebergCatalogBuilder{
		Connection: Connection{b, obj.Name, obj.SchemaName, obj.DatabaseName},
	}
}

func (b *ConnectionIcebergCatalogBuilder) CatalogType(s string) *ConnectionIcebergCatalogBuilder {
	b.catalogType = s
	return b
}

func (b *ConnectionIcebergCatalogBuilder) Url(s string) *ConnectionIcebergCatalogBuilder {
	b.url = s
	return b
}

func (b *ConnectionIcebergCatalogBuilder) Warehouse(s string) *ConnectionIcebergCatalogBuilder {
	b.warehouse = s
	return b
}

func (b *ConnectionIcebergCatalogBuilder) AwsConnection(s IdentifierSchemaStruct) *ConnectionIcebergCatalogBuilder {
	b.awsConnection = s
	return b
}

func (b *ConnectionIcebergCatalogBuilder) Validate(validate bool) *ConnectionIcebergCatalogBuilder {
	b.validate = validate
	return b
}

func (b *ConnectionIcebergCatalogBuilder) Create() error {
	q := strings.Builder{}
	q.WriteString(fmt.Sprintf(`CREATE CONNECTION %s TO ICEBERG CATALOG`, b.QualifiedName()))

	w := []string{}
	if b.catalogType != "" {
		o := fmt.Sprintf(`CATALOG TYPE = %s`, QuoteString(b.catalogType))
		w = append(w, o)
	}
	if b.url != "" {
		o := fmt.Sprintf(`URL = %s`, QuoteString(b.url))
		w = append(w, o)
	}
	if b.warehouse != "" {
		o := fmt.Sprintf(`WAREHOUSE = %s`, QuoteString(b.warehouse))
		w = append(w, o)
	}
	if b.awsConnection.Name != "" {
		o := fmt.Sprintf(`AWS CONNECTION = %s`, b.awsConnection.QualifiedName())
		w = append(w, o)
	}

	f := strings.Join(w, ", ")
	q.WriteString(fmt.Sprintf(` (%s)`, f))

	if !b.validate {
		q.WriteString(` WITH (VALIDATE = false)`)
	}

	q.WriteString(`;`)

	return b.ddl.exec(q.String())
}

type ConnectionIcebergCatalogParams struct {
	ConnectionId    sql.NullString `db:"id"`
	ConnectionName  sql.NullString `db:"connection_name"`
	SchemaName      sql.NullString `db:"schema_name"`
	DatabaseName    sql.NullString `db:"database_name"`
	CatalogType     sql.NullString `db:"catalog_type"`
	Url             sql.NullString `db:"url"`
	Warehouse       sql.NullString `db:"warehouse"`
	AwsConnectionId sql.NullString `db:"aws_connection_id"`
	Comment         sql.NullString `db:"comment"`
	OwnerName       sql.NullString `db:"owner_name"`
}

var connectionIcebergCatalogQuery = NewBaseQuery(`
	SELECT
		mz_connections.id,
		mz_connections.name AS connection_name,
		mz_schemas.name AS schema_name,
		mz_databases.name AS database_name,
		mz_iceberg_catalog_connections.catalog_type,
		mz_iceberg_catalog_connections.url,
		mz_iceberg_catalog_connections.warehouse,
		mz_iceberg_catalog_connections.aws_connection_id,
		comments.comment AS comment,
		mz_roles.name AS owner_name
	FROM mz_connections
	JOIN mz_schemas
		ON mz_connections.schema_id = mz_schemas.id
	JOIN mz_databases
		ON mz_schemas.database_id = mz_databases.id
	LEFT JOIN mz_internal.mz_iceberg_catalog_connections
		ON mz_connections.id = mz_iceberg_catalog_connections.id
	JOIN mz_roles
		ON mz_connections.owner_id = mz_roles.id
	LEFT JOIN (
		SELECT id, comment
		FROM mz_internal.mz_comments
		WHERE object_type = 'connection'
	) comments
		ON mz_connections.id = comments.id`)

func ScanConnectionIcebergCatalog(conn *sqlx.DB, id string) (ConnectionIcebergCatalogParams, error) {
	q := connectionIcebergCatalogQuery.QueryPredicate(map[string]string{"mz_connections.id": id})

	var c ConnectionIcebergCatalogParams
	if err := conn.Get(&c, q); err != nil {
		return c, err
	}

	return c, nil
}
