package materialize

import (
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
