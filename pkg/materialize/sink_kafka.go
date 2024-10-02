package materialize

import (
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"
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
	clusterName            string
	size                   string
	from                   IdentifierSchemaStruct
	kafkaConnection        IdentifierSchemaStruct
	topic                  string
	topicReplicationFactor int
	topicPartitionCount    int
	topicConfig            map[string]string
	compressionType        string
	key                    []string
	format                 SinkFormatSpecStruct
	envelope               KafkaSinkEnvelopeStruct
	snapshot               bool
	headers                string
	keyNotEnforced         bool
	partitionBy            string
}

func NewSinkKafkaBuilder(conn *sqlx.DB, obj MaterializeObject) *SinkKafkaBuilder {
	b := Builder{conn, BaseSink}
	return &SinkKafkaBuilder{
		Sink: Sink{b, obj.Name, obj.SchemaName, obj.DatabaseName},
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

func (b *SinkKafkaBuilder) TopicReplicationFactor(factor int) *SinkKafkaBuilder {
	b.topicReplicationFactor = factor
	return b
}

func (b *SinkKafkaBuilder) TopicPartitionCount(count int) *SinkKafkaBuilder {
	b.topicPartitionCount = count
	return b
}

func (b *SinkKafkaBuilder) TopicConfig(config map[string]string) *SinkKafkaBuilder {
	b.topicConfig = config
	return b
}

func (b *SinkKafkaBuilder) CompressionType(c string) *SinkKafkaBuilder {
	b.compressionType = c
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

func (b *SinkKafkaBuilder) Headers(h string) *SinkKafkaBuilder {
	b.headers = h
	return b
}

func (b *SinkKafkaBuilder) KeyNotEnforced(s bool) *SinkKafkaBuilder {
	b.keyNotEnforced = true
	return b
}

func (b *SinkKafkaBuilder) PartitionBy(expr string) *SinkKafkaBuilder {
	b.partitionBy = expr
	return b
}

func (b *SinkKafkaBuilder) Create() error {
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

	if b.topic != "" {
		q.WriteString(fmt.Sprintf(` (TOPIC %s`, QuoteString(b.topic)))
		if b.compressionType != "" {
			q.WriteString(fmt.Sprintf(`, COMPRESSION TYPE = %s`, b.compressionType))
		}
		if b.topicReplicationFactor > 0 {
			q.WriteString(fmt.Sprintf(`, TOPIC REPLICATION FACTOR = %d`, b.topicReplicationFactor))
		}
		if b.topicPartitionCount > 0 {
			q.WriteString(fmt.Sprintf(`, TOPIC PARTITION COUNT = %d`, b.topicPartitionCount))
		}
		if len(b.topicConfig) > 0 {
			configItems := make([]string, 0, len(b.topicConfig))
			for k, v := range b.topicConfig {
				configItems = append(configItems, fmt.Sprintf("%s => %s", QuoteString(k), QuoteString(v)))
			}
			q.WriteString(fmt.Sprintf(`, TOPIC CONFIG MAP[%s]`, strings.Join(configItems, ", ")))
		}

		if b.partitionBy != "" {
			q.WriteString(fmt.Sprintf(`, PARTITION BY %s`, b.partitionBy))
		}

		q.WriteString(")")
	}

	if len(b.key) > 0 {
		o := strings.Join(b.key[:], ", ")
		q.WriteString(fmt.Sprintf(` KEY (%s)`, o))
	}

	if b.keyNotEnforced {
		q.WriteString(` NOT ENFORCED`)
	}

	if b.headers != "" {
		q.WriteString(fmt.Sprintf(` HEADERS %s`, b.headers))
	}

	if b.format.Json {
		q.WriteString(` FORMAT JSON`)
	}

	if b.format.Avro != nil {
		if b.format.Avro.SchemaRegistryConnection.Name != "" {
			q.WriteString(fmt.Sprintf(` FORMAT AVRO USING CONFLUENT SCHEMA REGISTRY CONNECTION %s`, b.format.Avro.SchemaRegistryConnection.QualifiedName()))
		}

		// CSR Connection Options
		var v = []string{}
		if b.format.Avro.AvroValueFullname != "" && b.format.Avro.AvroKeyFullname != "" {
			v = append(v, fmt.Sprintf(`AVRO KEY FULLNAME %s AVRO VALUE FULLNAME %s`,
				QuoteString(b.format.Avro.AvroKeyFullname),
				QuoteString(b.format.Avro.AvroValueFullname)),
			)
		}

		// Doc Type
		if b.format.Avro.DocType.Object.Name != "" {
			c := strings.Builder{}
			if b.format.Avro.DocType.Key {
				c.WriteString("KEY ")
			} else if b.format.Avro.DocType.Value {
				c.WriteString("VALUE ")
			}
			c.WriteString(fmt.Sprintf("DOC ON TYPE %[1]s = %[2]s",
				b.format.Avro.DocType.Object.QualifiedName(),
				QuoteString(b.format.Avro.DocType.Doc),
			))
			v = append(v, c.String())
		}

		// Doc Column
		for _, ac := range b.format.Avro.DocColumn {
			c := strings.Builder{}
			if ac.Key {
				c.WriteString("KEY")
			} else if ac.Value {
				c.WriteString("VALUE")
			}
			f := b.from.QualifiedName() + "." + QuoteIdentifier(ac.Column)
			c.WriteString(fmt.Sprintf(" DOC ON COLUMN %[1]s = %[2]s", f, QuoteString(ac.Doc)))
			v = append(v, c.String())
		}

		if b.format.Avro.KeyCompatibilityLevel != "" {
			v = append(v, fmt.Sprintf("KEY COMPATIBILITY LEVEL %s", QuoteString(b.format.Avro.KeyCompatibilityLevel)))
		}
		if b.format.Avro.ValueCompatibilityLevel != "" {
			v = append(v, fmt.Sprintf("VALUE COMPATIBILITY LEVEL %s", QuoteString(b.format.Avro.ValueCompatibilityLevel)))
		}

		if len(v) > 0 {
			q.WriteString(fmt.Sprintf(` (%s)`, strings.Join(v[:], ", ")))
		}
	}

	if b.envelope.Debezium {
		q.WriteString(` ENVELOPE DEBEZIUM`)
	} else if b.envelope.Upsert {
		q.WriteString(` ENVELOPE UPSERT`)
	}

	// With Options
	withOptions := []string{}
	if b.snapshot {
		withOptions = append(withOptions, "SNAPSHOT = true")
	}

	if len(withOptions) > 0 {
		q.WriteString(fmt.Sprintf(` WITH (%s)`, strings.Join(withOptions, ", ")))
	}

	return b.ddl.exec(q.String())
}
