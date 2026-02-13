package materialize

import (
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"
)

type SinkIcebergBuilder struct {
	Sink
	clusterName              string
	from                     IdentifierSchemaStruct
	icebergCatalogConnection IdentifierSchemaStruct
	namespace                string
	table                    string
	awsConnection            IdentifierSchemaStruct
	key                      []string
	keyNotEnforced           bool
	commitInterval           string
}

func NewSinkIcebergBuilder(conn *sqlx.DB, obj MaterializeObject) *SinkIcebergBuilder {
	b := Builder{conn, BaseSink}
	return &SinkIcebergBuilder{
		Sink: Sink{b, obj.Name, obj.SchemaName, obj.DatabaseName},
	}
}

func (b *SinkIcebergBuilder) ClusterName(c string) *SinkIcebergBuilder {
	b.clusterName = c
	return b
}

func (b *SinkIcebergBuilder) From(i IdentifierSchemaStruct) *SinkIcebergBuilder {
	b.from = i
	return b
}

func (b *SinkIcebergBuilder) IcebergCatalogConnection(i IdentifierSchemaStruct) *SinkIcebergBuilder {
	b.icebergCatalogConnection = i
	return b
}

func (b *SinkIcebergBuilder) Namespace(n string) *SinkIcebergBuilder {
	b.namespace = n
	return b
}

func (b *SinkIcebergBuilder) Table(t string) *SinkIcebergBuilder {
	b.table = t
	return b
}

func (b *SinkIcebergBuilder) AwsConnection(a IdentifierSchemaStruct) *SinkIcebergBuilder {
	b.awsConnection = a
	return b
}

func (b *SinkIcebergBuilder) Key(k []string) *SinkIcebergBuilder {
	b.key = k
	return b
}

func (b *SinkIcebergBuilder) KeyNotEnforced(k bool) *SinkIcebergBuilder {
	b.keyNotEnforced = k
	return b
}

func (b *SinkIcebergBuilder) CommitInterval(c string) *SinkIcebergBuilder {
	b.commitInterval = c
	return b
}

func (b *SinkIcebergBuilder) Create() error {
	q := strings.Builder{}
	q.WriteString(fmt.Sprintf(`CREATE SINK %s`, b.QualifiedName()))

	if b.clusterName != "" {
		q.WriteString(fmt.Sprintf(` IN CLUSTER %s`, QuoteIdentifier(b.clusterName)))
	}

	q.WriteString(fmt.Sprintf(` FROM %s`, b.from.QualifiedName()))

	// INTO ICEBERG CATALOG CONNECTION
	q.WriteString(fmt.Sprintf(` INTO ICEBERG CATALOG CONNECTION %s`, b.icebergCatalogConnection.QualifiedName()))

	// Catalog options (NAMESPACE, TABLE)
	catalogOptions := []string{}
	if b.namespace != "" {
		catalogOptions = append(catalogOptions, fmt.Sprintf(`NAMESPACE = %s`, QuoteString(b.namespace)))
	}
	if b.table != "" {
		catalogOptions = append(catalogOptions, fmt.Sprintf(`TABLE = %s`, QuoteString(b.table)))
	}
	if len(catalogOptions) > 0 {
		q.WriteString(fmt.Sprintf(` (%s)`, strings.Join(catalogOptions, ", ")))
	}

	// USING AWS CONNECTION
	if b.awsConnection.Name != "" {
		q.WriteString(fmt.Sprintf(` USING AWS CONNECTION %s`, b.awsConnection.QualifiedName()))
	}

	// KEY
	if len(b.key) > 0 {
		q.WriteString(fmt.Sprintf(` KEY (%s)`, strings.Join(b.key, ", ")))
	}

	// NOT ENFORCED
	if b.keyNotEnforced {
		q.WriteString(` NOT ENFORCED`)
	}

	// MODE UPSERT is required for Iceberg sinks
	q.WriteString(` MODE UPSERT`)

	// WITH options
	withOptions := []string{}
	if b.commitInterval != "" {
		withOptions = append(withOptions, fmt.Sprintf(`COMMIT INTERVAL = %s`, QuoteString(b.commitInterval)))
	}
	if len(withOptions) > 0 {
		q.WriteString(fmt.Sprintf(` WITH (%s)`, strings.Join(withOptions, ", ")))
	}

	q.WriteString(`;`)

	return b.ddl.exec(q.String())
}
