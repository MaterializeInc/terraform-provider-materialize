package materialize

import (
	"fmt"
	"strings"
)

type KafkaSinkEnvelopeStruct struct {
	Upsert   bool
	Debezium bool
}

func GetSinkKafkaEnelopeStruct(v interface{}) KafkaSinkEnvelopeStruct {
	var envelope KafkaSinkEnvelopeStruct
	if v, ok := v.([]interface{})[0].(map[string]interface{})["upsert"]; ok {
		envelope.Upsert = v.(bool)
	}
	if v, ok := v.([]interface{})[0].(map[string]interface{})["debezium"]; ok {
		envelope.Debezium = v.(bool)
	}
	return envelope
}

type SinkKafkaBuilder struct {
	Sink
	clusterName     string
	size            string
	from            IdentifierSchemaStruct
	kafkaConnection IdentifierSchemaStruct
	topic           string
	key             []string
	format          SinkFormatSpecStruct
	envelope        KafkaSinkEnvelopeStruct
	snapshot        bool
}

func NewSinkKafkaBuilder(sinkName, schemaName, databaseName string) *SinkKafkaBuilder {
	return &SinkKafkaBuilder{
		Sink: Sink{sinkName, schemaName, databaseName},
	}
}

func (b *SinkKafkaBuilder) ClusterName(c string) *SinkKafkaBuilder {
	b.clusterName = c
	return b
}

func (b *SinkKafkaBuilder) Size(s string) *SinkKafkaBuilder {
	b.size = s
	return b
}

func (b *SinkKafkaBuilder) From(i IdentifierSchemaStruct) *SinkKafkaBuilder {
	b.from = i
	return b
}

func (b *SinkKafkaBuilder) KafkaConnection(k IdentifierSchemaStruct) *SinkKafkaBuilder {
	b.kafkaConnection = k
	return b
}

func (b *SinkKafkaBuilder) Topic(t string) *SinkKafkaBuilder {
	b.topic = t
	return b
}

func (b *SinkKafkaBuilder) Key(k []string) *SinkKafkaBuilder {
	b.key = k
	return b
}

func (b *SinkKafkaBuilder) Format(f SinkFormatSpecStruct) *SinkKafkaBuilder {
	b.format = f
	return b
}

func (b *SinkKafkaBuilder) Envelope(e KafkaSinkEnvelopeStruct) *SinkKafkaBuilder {
	b.envelope = e
	return b
}

func (b *SinkKafkaBuilder) Snapshot(s bool) *SinkKafkaBuilder {
	b.snapshot = s
	return b
}

func (b *SinkKafkaBuilder) Create() string {
	q := strings.Builder{}
	q.WriteString(fmt.Sprintf(`CREATE SINK %s`, b.QualifiedName()))

	if b.clusterName != "" {
		q.WriteString(fmt.Sprintf(` IN CLUSTER %s`, QuoteIdentifier(b.clusterName)))
	}

	q.WriteString(fmt.Sprintf(` FROM %s`, b.from.QualifiedName()))

	// Broker
	if b.kafkaConnection.Name != "" {
		q.WriteString(fmt.Sprintf(` INTO KAFKA CONNECTION %s`, b.kafkaConnection.QualifiedName()))
	}

	if len(b.key) > 0 {
		o := strings.Join(b.key[:], ", ")
		q.WriteString(fmt.Sprintf(` KEY (%s)`, o))
	}

	if b.topic != "" {
		q.WriteString(fmt.Sprintf(` (TOPIC %s)`, QuoteString(b.topic)))
	}

	if b.format.Json {
		q.WriteString(` FORMAT JSON`)
	}

	if b.format.Avro != nil {
		if b.format.Avro.SchemaRegistryConnection.Name != "" {
			q.WriteString(fmt.Sprintf(` FORMAT AVRO USING CONFLUENT SCHEMA REGISTRY CONNECTION %s`, b.format.Avro.SchemaRegistryConnection.QualifiedName()))
		}
		if b.format.Avro.AvroValueFullname != "" && b.format.Avro.AvroKeyFullname != "" {
			q.WriteString(fmt.Sprintf(` WITH (AVRO KEY FULLNAME %s AVRO VALUE FULLNAME %s)`, QuoteString(b.format.Avro.AvroKeyFullname), QuoteString(b.format.Avro.AvroValueFullname)))
		}
	}

	if b.envelope.Debezium {
		q.WriteString(` ENVELOPE DEBEZIUM`)
	}

	if b.envelope.Upsert {
		q.WriteString(` ENVELOPE UPSERT`)
	}

	// With Options
	if b.size != "" || !b.snapshot {
		w := strings.Builder{}

		if b.size != "" {
			w.WriteString(fmt.Sprintf(` SIZE = %s`, QuoteString(b.size)))
		}

		if !b.snapshot {
			w.WriteString(` SNAPSHOT = false`)
		}

		q.WriteString(fmt.Sprintf(` WITH (%s)`, w.String()))
	}

	q.WriteString(`;`)
	return q.String()
}

func (b *SinkKafkaBuilder) Rename(newName string) string {
	n := QualifiedName(b.DatabaseName, b.SchemaName, newName)
	return fmt.Sprintf(`ALTER SINK %s RENAME TO %s;`, b.QualifiedName(), n)
}

func (b *SinkKafkaBuilder) UpdateSize(newSize string) string {
	return fmt.Sprintf(`ALTER SINK %s SET (SIZE = %s);`, b.QualifiedName(), QuoteString(newSize))
}

func (b *SinkKafkaBuilder) Drop() string {
	return fmt.Sprintf(`DROP SINK %s;`, b.QualifiedName())
}

func (b *SinkKafkaBuilder) ReadId() string {
	return ReadSinkId(b.SinkName, b.SchemaName, b.DatabaseName)
}
