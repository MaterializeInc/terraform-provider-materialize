package materialize

import (
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"
)

type KafkaSourceEnvelopeStruct struct {
	Debezium bool
	None     bool
	Upsert   bool
}

func GetSourceKafkaEnelopeStruct(v interface{}) KafkaSourceEnvelopeStruct {
	var envelope KafkaSourceEnvelopeStruct
	if v, ok := v.([]interface{})[0].(map[string]interface{})["upsert"]; ok {
		envelope.Upsert = v.(bool)
	}
	if v, ok := v.([]interface{})[0].(map[string]interface{})["debezium"]; ok {
		envelope.Debezium = v.(bool)
	}
	if v, ok := v.([]interface{})[0].(map[string]interface{})["none"]; ok {
		envelope.None = v.(bool)
	}
	return envelope
}

type SourceKafkaBuilder struct {
	Source
	clusterName      string
	size             string
	kafkaConnection  IdentifierSchemaStruct
	topic            string
	includeKey       bool
	includeHeaders   bool
	includePartition bool
	includeOffset    bool
	includeTimestamp bool
	format           SourceFormatSpecStruct
	keyFormat        SourceFormatSpecStruct
	valueFormat      SourceFormatSpecStruct
	envelope         KafkaSourceEnvelopeStruct
	primaryKey       []string
	startOffset      []int
	startTimestamp   int
	exposeProgress   string
}

func NewSourceKafkaBuilder(conn *sqlx.DB, obj ObjectSchemaStruct) *SourceKafkaBuilder {
	b := Builder{conn, BaseSink}
	return &SourceKafkaBuilder{
		Source: Source{b, obj.Name, obj.SchemaName, obj.DatabaseName},
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

func (b *SourceKafkaBuilder) IncludeKey() *SourceKafkaBuilder {
	b.includeKey = true
	return b
}

func (b *SourceKafkaBuilder) IncludeHeaders() *SourceKafkaBuilder {
	b.includeHeaders = true
	return b
}

func (b *SourceKafkaBuilder) IncludePartition() *SourceKafkaBuilder {
	b.includePartition = true
	return b
}

func (b *SourceKafkaBuilder) IncludeOffset() *SourceKafkaBuilder {
	b.includeOffset = true
	return b
}

func (b *SourceKafkaBuilder) IncludeTimestamp() *SourceKafkaBuilder {
	b.includeTimestamp = true
	return b
}

func (b *SourceKafkaBuilder) Format(f SourceFormatSpecStruct) *SourceKafkaBuilder {
	b.format = f
	return b
}

func (b *SourceKafkaBuilder) Envelope(e KafkaSourceEnvelopeStruct) *SourceKafkaBuilder {
	b.envelope = e
	return b
}

func (b *SourceKafkaBuilder) KeyFormat(k SourceFormatSpecStruct) *SourceKafkaBuilder {
	b.keyFormat = k
	return b
}

func (b *SourceKafkaBuilder) ValueFormat(v SourceFormatSpecStruct) *SourceKafkaBuilder {
	b.valueFormat = v
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

func (b *SourceKafkaBuilder) ExposeProgress(s string) *SourceKafkaBuilder {
	b.exposeProgress = s
	return b
}

func (b *SourceKafkaBuilder) Create() error {
	q := strings.Builder{}
	q.WriteString(fmt.Sprintf(`CREATE SOURCE %s`, b.QualifiedName()))

	if b.clusterName != "" {
		q.WriteString(fmt.Sprintf(` IN CLUSTER %s`, QuoteIdentifier(b.clusterName)))
	}

	q.WriteString(fmt.Sprintf(` FROM KAFKA CONNECTION %s`, b.kafkaConnection.QualifiedName()))
	q.WriteString(fmt.Sprintf(` (TOPIC %s`, QuoteString(b.topic)))

	if b.startTimestamp != 0 {
		q.WriteString(fmt.Sprintf(`, START TIMESTAMP %d`, b.startTimestamp))
	}

	q.WriteString(`)`)

	// Format
	if b.format.Avro != nil {
		if b.format.Avro.SchemaRegistryConnection.Name != "" {
			q.WriteString(fmt.Sprintf(` FORMAT AVRO USING CONFLUENT SCHEMA REGISTRY CONNECTION %s`, QualifiedName(b.format.Avro.SchemaRegistryConnection.DatabaseName, b.format.Avro.SchemaRegistryConnection.SchemaName, b.format.Avro.SchemaRegistryConnection.Name)))
		}
		if b.format.Avro.KeyStrategy != "" {
			q.WriteString(fmt.Sprintf(` KEY STRATEGY %s`, b.format.Avro.KeyStrategy))
		}
		if b.format.Avro.ValueStrategy != "" {
			q.WriteString(fmt.Sprintf(` VALUE STRATEGY %s`, b.format.Avro.ValueStrategy))
		}
	}

	if b.format.Protobuf != nil {
		if b.format.Protobuf.SchemaRegistryConnection.Name != "" && b.format.Protobuf.MessageName != "" {
			q.WriteString(fmt.Sprintf(` FORMAT PROTOBUF MESSAGE '%s' USING CONFLUENT SCHEMA REGISTRY CONNECTION %s`, b.format.Protobuf.MessageName, QualifiedName(b.format.Protobuf.SchemaRegistryConnection.DatabaseName, b.format.Protobuf.SchemaRegistryConnection.SchemaName, b.format.Protobuf.SchemaRegistryConnection.Name)))
		}

		if b.format.Protobuf.SchemaRegistryConnection.Name != "" {
			q.WriteString(fmt.Sprintf(` FORMAT PROTOBUF USING CONFLUENT SCHEMA REGISTRY CONNECTION %s`, QualifiedName(b.format.Protobuf.SchemaRegistryConnection.DatabaseName, b.format.Protobuf.SchemaRegistryConnection.SchemaName, b.format.Protobuf.SchemaRegistryConnection.Name)))
		}
	}

	if b.format.Csv != nil {
		if b.format.Csv.Columns > 0 {
			q.WriteString(fmt.Sprintf(` FORMAT CSV WITH %d COLUMNS`, b.format.Csv.Columns))
		}

		if b.format.Csv.Header != nil {
			q.WriteString(fmt.Sprintf(` FORMAT CSV WITH HEADER ( %s )`, strings.Join(b.format.Csv.Header, ", ")))
		}

		if b.format.Csv.DelimitedBy != "" {
			q.WriteString(fmt.Sprintf(` DELIMITER '%s'`, b.format.Csv.DelimitedBy))
		}
	}

	if b.format.Bytes {
		q.WriteString(` FORMAT BYTES`)
	}

	if b.format.Text {
		q.WriteString(` FORMAT TEXT`)
	}

	if b.keyFormat.Avro != nil {
		if b.keyFormat.Avro.SchemaRegistryConnection.Name != "" {
			q.WriteString(fmt.Sprintf(` KEY FORMAT AVRO USING CONFLUENT SCHEMA REGISTRY CONNECTION %s`, QualifiedName(b.keyFormat.Avro.SchemaRegistryConnection.DatabaseName, b.keyFormat.Avro.SchemaRegistryConnection.SchemaName, b.keyFormat.Avro.SchemaRegistryConnection.Name)))
		}
		if b.keyFormat.Avro.KeyStrategy != "" {
			q.WriteString(fmt.Sprintf(` KEY STRATEGY %s`, b.keyFormat.Avro.KeyStrategy))
		}
		if b.keyFormat.Avro.ValueStrategy != "" {
			q.WriteString(fmt.Sprintf(` VALUE STRATEGY %s`, b.keyFormat.Avro.ValueStrategy))
		}
	}

	if b.keyFormat.Protobuf != nil {
		if b.keyFormat.Protobuf.SchemaRegistryConnection.Name != "" && b.keyFormat.Protobuf.MessageName != "" {
			q.WriteString(fmt.Sprintf(` KEY FORMAT PROTOBUF MESSAGE '%s' USING CONFLUENT SCHEMA REGISTRY CONNECTION %s`, b.keyFormat.Protobuf.MessageName, QualifiedName(b.keyFormat.Protobuf.SchemaRegistryConnection.DatabaseName, b.keyFormat.Protobuf.SchemaRegistryConnection.SchemaName, b.keyFormat.Protobuf.SchemaRegistryConnection.Name)))
		}

		if b.keyFormat.Protobuf.SchemaRegistryConnection.Name != "" {
			q.WriteString(fmt.Sprintf(` KEY FORMAT PROTOBUF USING CONFLUENT SCHEMA REGISTRY CONNECTION %s`, QualifiedName(b.keyFormat.Protobuf.SchemaRegistryConnection.DatabaseName, b.keyFormat.Protobuf.SchemaRegistryConnection.SchemaName, b.keyFormat.Protobuf.SchemaRegistryConnection.Name)))
		}
	}

	if b.keyFormat.Csv != nil {
		if b.keyFormat.Csv.Columns > 0 {
			q.WriteString(fmt.Sprintf(` KEY FORMAT CSV WITH %d COLUMNS`, b.keyFormat.Csv.Columns))
		}

		if b.keyFormat.Csv.Header != nil {
			q.WriteString(fmt.Sprintf(` KEY FORMAT CSV WITH HEADER ( %s )`, strings.Join(b.keyFormat.Csv.Header, ", ")))
		}

		if b.keyFormat.Csv.DelimitedBy != "" {
			q.WriteString(fmt.Sprintf(` DELIMITER '%s'`, b.keyFormat.Csv.DelimitedBy))
		}
	}

	if b.keyFormat.Bytes {
		q.WriteString(` KEY FORMAT BYTES`)
	}

	if b.keyFormat.Text {
		q.WriteString(` KEY FORMAT TEXT`)
	}

	if b.valueFormat.Avro != nil {
		if b.valueFormat.Avro.SchemaRegistryConnection.Name != "" {
			q.WriteString(fmt.Sprintf(` VALUE FORMAT AVRO USING CONFLUENT SCHEMA REGISTRY CONNECTION %s`, QualifiedName(b.valueFormat.Avro.SchemaRegistryConnection.DatabaseName, b.valueFormat.Avro.SchemaRegistryConnection.SchemaName, b.valueFormat.Avro.SchemaRegistryConnection.Name)))
		}
		if b.valueFormat.Avro.KeyStrategy != "" {
			q.WriteString(fmt.Sprintf(` VALUE STRATEGY %s`, b.valueFormat.Avro.KeyStrategy))
		}
		if b.valueFormat.Avro.ValueStrategy != "" {
			q.WriteString(fmt.Sprintf(` VALUE STRATEGY %s`, b.valueFormat.Avro.ValueStrategy))
		}
	}

	if b.valueFormat.Protobuf != nil {
		if b.valueFormat.Protobuf.SchemaRegistryConnection.Name != "" && b.valueFormat.Protobuf.MessageName != "" {
			q.WriteString(fmt.Sprintf(` VALUE FORMAT PROTOBUF MESSAGE '%s' USING CONFLUENT SCHEMA REGISTRY CONNECTION %s`, b.valueFormat.Protobuf.MessageName, QualifiedName(b.valueFormat.Protobuf.SchemaRegistryConnection.DatabaseName, b.valueFormat.Protobuf.SchemaRegistryConnection.SchemaName, b.valueFormat.Protobuf.SchemaRegistryConnection.Name)))
		}

		if b.valueFormat.Protobuf.SchemaRegistryConnection.Name != "" {
			q.WriteString(fmt.Sprintf(` VALUE FORMAT PROTOBUF USING CONFLUENT SCHEMA REGISTRY CONNECTION %s`, QualifiedName(b.valueFormat.Protobuf.SchemaRegistryConnection.DatabaseName, b.valueFormat.Protobuf.SchemaRegistryConnection.SchemaName, b.valueFormat.Protobuf.SchemaRegistryConnection.Name)))
		}
	}

	if b.valueFormat.Csv != nil {
		if b.valueFormat.Csv.Columns > 0 {
			q.WriteString(fmt.Sprintf(` VALUE FORMAT CSV WITH %d COLUMNS`, b.valueFormat.Csv.Columns))
		}

		if b.valueFormat.Csv.Header != nil {
			q.WriteString(fmt.Sprintf(` VALUE FORMAT CSV WITH HEADER ( %s )`, strings.Join(b.valueFormat.Csv.Header, ", ")))
		}

		if b.valueFormat.Csv.DelimitedBy != "" {
			q.WriteString(fmt.Sprintf(` DELIMITER '%s'`, b.valueFormat.Csv.DelimitedBy))
		}
	}

	if b.valueFormat.Bytes {
		q.WriteString(` VALUE FORMAT BYTES`)
	}

	if b.valueFormat.Text {
		q.WriteString(` VALUE FORMAT TEXT`)
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

	// Metadata
	var i []string

	if b.includeKey {
		i = append(i, "KEY")
	}

	if b.includeHeaders {
		i = append(i, "HEADERS")
	}

	if b.includePartition {
		i = append(i, "PARTITION")
	}

	if b.includeOffset {
		i = append(i, "OFFSET")
	}

	if b.includeTimestamp {
		i = append(i, "TIMESTAMP")
	}

	if len(i) > 0 {
		o := strings.Join(i[:], ", ")
		q.WriteString(fmt.Sprintf(` INCLUDE %s`, o))
	}

	if b.envelope.Debezium {
		q.WriteString(` ENVELOPE DEBEZIUM`)
	}

	if b.envelope.Upsert {
		q.WriteString(` ENVELOPE UPSERT`)
	}

	if b.envelope.None {
		q.WriteString(` ENVELOPE NONE`)
	}

	if b.exposeProgress != "" {
		q.WriteString(fmt.Sprintf(` EXPOSE PROGRESS AS %s`, b.exposeProgress))
	}

	if b.size != "" {
		q.WriteString(fmt.Sprintf(` WITH (SIZE = %s)`, QuoteString(b.size)))
	}

	q.WriteString(`;`)
	return b.ddl.exec(q.String())
}
