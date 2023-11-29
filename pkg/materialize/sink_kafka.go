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

type AvroColumnStruct struct {
	Key    string
	Value  string
	Column string
}

func GetAvroColumnStruct(v []interface{}) []AvroColumnStruct {
	var comments []AvroColumnStruct
	for _, comment := range v {
		c := comment.(map[string]interface{})
		comments = append(comments, AvroColumnStruct{
			Key:    c["key"].(string),
			Value:  c["value"].(string),
			Column: c["column"].(string),
		})
	}
	return comments
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
	keyNotEnforced  bool
	avroDoc         string
	avroColumnDoc   []AvroColumnStruct
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

func (b *SinkKafkaBuilder) KeyNotEnforced(s bool) *SinkKafkaBuilder {
	b.keyNotEnforced = true
	return b
}

func (b *SinkKafkaBuilder) AvroDoc(a string) *SinkKafkaBuilder {
	b.avroDoc = a
	return b
}

func (b *SinkKafkaBuilder) AvroColumnDoc(a []AvroColumnStruct) *SinkKafkaBuilder {
	b.avroColumnDoc = a
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
		q.WriteString(fmt.Sprintf(` (TOPIC %s)`, QuoteString(b.topic)))
	}

	if len(b.key) > 0 {
		o := strings.Join(b.key[:], ", ")
		q.WriteString(fmt.Sprintf(` KEY (%s)`, o))
	}

	if b.keyNotEnforced {
		q.WriteString(` NOT ENFORCED`)
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
	} else if b.envelope.Upsert {
		q.WriteString(` ENVELOPE UPSERT`)
	}

	// With Options
	withOptions := []string{}
	if b.size != "" {
		withOptions = append(withOptions, fmt.Sprintf(`SIZE = %s`, QuoteString(b.size)))
	}
	if b.snapshot {
		withOptions = append(withOptions, "SNAPSHOT = true")
	}

	if len(withOptions) > 0 {
		q.WriteString(fmt.Sprintf(` WITH (%s)`, strings.Join(withOptions, ", ")))
	}

	// Avro Comments
	var v = []string{}
	if b.avroDoc != "" {
		v = append(v, fmt.Sprintf(`DOC ON TYPE %s = %s`, b.from.QualifiedName(), QuoteString(b.avroDoc)))
	}
	for _, ac := range b.avroColumnDoc {
		c := b.from.QualifiedName() + "." + QuoteIdentifier(ac.Column)
		v = append(v, fmt.Sprintf(`KEY DOC ON COLUMN %s = %s`, c, QuoteString(ac.Key)))
		v = append(v, fmt.Sprintf(`VALUE DOC ON COLUMN TYPE %s = %s`, c, QuoteString(ac.Value)))
	}
	if len(v) > 0 {
		q.WriteString(fmt.Sprintf(` (%s)`, strings.Join(v[:], ", ")))
	}

	return b.ddl.exec(q.String())
}
