package materialize

import (
	"fmt"
	"strings"
)

type SourceKafkaBuilder struct {
	Source
	clusterName              string
	size                     string
	kafkaConnection          IdentifierSchemaStruct
	topic                    string
	includeKey               string
	includeHeaders           bool
	includePartition         string
	includeOffset            string
	includeTimestamp         string
	format                   string
	keyFormat                string
	envelope                 string
	schemaRegistryConnection IdentifierSchemaStruct
	keyStrategy              string
	valueStrategy            string
	primaryKey               []string
	startOffset              []int
	startTimestamp           int
}

func NewSourceKafkaBuilder(sourceName, schemaName, databaseName string) *SourceKafkaBuilder {
	return &SourceKafkaBuilder{
		Source: Source{sourceName, schemaName, databaseName},
	}
}

func (b *SourceKafkaBuilder) ClusterName(c string) *SourceKafkaBuilder {
	b.clusterName = c
	return b
}

func (b *SourceKafkaBuilder) Size(s string) *SourceKafkaBuilder {
	b.size = s
	return b
}

func (b *SourceKafkaBuilder) KafkaConnection(k IdentifierSchemaStruct) *SourceKafkaBuilder {
	b.kafkaConnection = k
	return b
}

func (b *SourceKafkaBuilder) Topic(t string) *SourceKafkaBuilder {
	b.topic = t
	return b
}

func (b *SourceKafkaBuilder) IncludeKey(i string) *SourceKafkaBuilder {
	b.includeKey = i
	return b
}

func (b *SourceKafkaBuilder) IncludeHeaders() *SourceKafkaBuilder {
	b.includeHeaders = true
	return b
}

func (b *SourceKafkaBuilder) IncludePartition(i string) *SourceKafkaBuilder {
	b.includePartition = i
	return b
}

func (b *SourceKafkaBuilder) IncludeOffset(i string) *SourceKafkaBuilder {
	b.includeOffset = i
	return b
}

func (b *SourceKafkaBuilder) IncludeTimestamp(i string) *SourceKafkaBuilder {
	b.includeTimestamp = i
	return b
}

func (b *SourceKafkaBuilder) Format(f string) *SourceKafkaBuilder {
	b.format = f
	return b
}

func (b *SourceKafkaBuilder) Envelope(e string) *SourceKafkaBuilder {
	b.envelope = e
	return b
}

func (b *SourceKafkaBuilder) SchemaRegistryConnection(s IdentifierSchemaStruct) *SourceKafkaBuilder {
	b.schemaRegistryConnection = s
	return b
}

func (b *SourceKafkaBuilder) KeyFormat(k string) *SourceKafkaBuilder {
	b.keyFormat = k
	return b
}

func (b *SourceKafkaBuilder) KeyStrategy(k string) *SourceKafkaBuilder {
	b.keyStrategy = k
	return b
}

func (b *SourceKafkaBuilder) ValueStrategy(v string) *SourceKafkaBuilder {
	b.valueStrategy = v
	return b
}

func (b *SourceKafkaBuilder) PrimaryKey(p []string) *SourceKafkaBuilder {
	b.primaryKey = p
	return b
}

func (b *SourceKafkaBuilder) StartOffset(s []int) *SourceKafkaBuilder {
	b.startOffset = s
	return b
}

func (b *SourceKafkaBuilder) StartTimestamp(s int) *SourceKafkaBuilder {
	b.startTimestamp = s
	return b
}

func (b *SourceKafkaBuilder) Create() string {
	q := strings.Builder{}
	q.WriteString(fmt.Sprintf(`CREATE SOURCE %s`, b.QualifiedName()))

	if b.clusterName != "" {
		q.WriteString(fmt.Sprintf(` IN CLUSTER %s`, QuoteIdentifier(b.clusterName)))
	}

	q.WriteString(fmt.Sprintf(` FROM KAFKA CONNECTION %s`, b.kafkaConnection.QualifiedName()))
	q.WriteString(fmt.Sprintf(` (TOPIC %s)`, QuoteString(b.topic)))

	// Format
	if b.keyFormat != "" {
		q.WriteString(fmt.Sprintf(` KEY FORMAT %s VALUE FORMAT %s`, b.keyFormat, b.format))
	} else {
		q.WriteString(fmt.Sprintf(` FORMAT %s`, b.format))
	}

	if b.schemaRegistryConnection.Name != "" {
		q.WriteString(fmt.Sprintf(` USING CONFLUENT SCHEMA REGISTRY CONNECTION %s`, b.schemaRegistryConnection.QualifiedName()))
	}

	// Key Constraint
	if len(b.primaryKey) > 0 {
		k := strings.Join(b.primaryKey[:], ", ")
		q.WriteString(fmt.Sprintf(` PRIMARY KEY (%s) NOT ENFORCED`, k))
	}

	// Time-based Offsets
	if len(b.startOffset) > 0 {
		k := strings.Join(strings.Fields(fmt.Sprint(b.startOffset)), ", ")
		q.WriteString(fmt.Sprintf(` START OFFSET %s`, k))
	}

	if b.startTimestamp != 0 {
		q.WriteString(fmt.Sprintf(` START TIMESTAMP %d`, b.startTimestamp))
	}

	// Strategy
	if b.keyStrategy != "" {
		q.WriteString(fmt.Sprintf(` KEY STRATEGY %s`, b.keyStrategy))
	}

	if b.valueStrategy != "" {
		q.WriteString(fmt.Sprintf(` VALUE STRATEGY %s`, b.valueStrategy))
	}

	// Metadata
	var i []string

	if b.includeKey != "" {
		i = append(i, b.includeKey)
	}

	if b.includeHeaders {
		i = append(i, "HEADERS")
	}

	if b.includePartition != "" {
		i = append(i, b.includePartition)
	}

	if b.includeOffset != "" {
		i = append(i, b.includeOffset)
	}

	if b.includeTimestamp != "" {
		i = append(i, b.includeTimestamp)
	}

	if len(i) > 0 {
		o := strings.Join(i[:], ", ")
		q.WriteString(fmt.Sprintf(` INCLUDE %s`, o))
	}

	if b.envelope != "" {
		q.WriteString(fmt.Sprintf(` ENVELOPE %s`, b.envelope))
	}

	if b.size != "" {
		q.WriteString(fmt.Sprintf(` WITH (SIZE = %s)`, QuoteString(b.size)))
	}

	q.WriteString(`;`)
	return q.String()
}

func (b *SourceKafkaBuilder) Rename(newName string) string {
	n := QualifiedName(b.DatabaseName, b.SchemaName, newName)
	return fmt.Sprintf(`ALTER SOURCE %s RENAME TO %s;`, b.QualifiedName(), n)
}

func (b *SourceKafkaBuilder) UpdateSize(newSize string) string {
	return fmt.Sprintf(`ALTER SOURCE %s SET (SIZE = %s);`, b.QualifiedName(), QuoteString(newSize))
}

func (b *SourceKafkaBuilder) Drop() string {
	return fmt.Sprintf(`DROP SOURCE %s;`, b.QualifiedName())
}

func (b *SourceKafkaBuilder) ReadId() string {
	return ReadSourceId(b.SourceName, b.SchemaName, b.DatabaseName)
}
