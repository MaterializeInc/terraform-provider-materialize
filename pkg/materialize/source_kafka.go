package materialize

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/jmoiron/sqlx"
)

type KafkaSourceEnvelopeStruct struct {
	Debezium      bool
	None          bool
	Upsert        bool
	UpsertOptions *UpsertOptionsStruct
}

type UpsertOptionsStruct struct {
	ValueDecodingErrors struct {
		Inline struct {
			Enabled bool
			Alias   string
		}
	}
}

func GetSourceKafkaEnvelopeStruct(v interface{}) KafkaSourceEnvelopeStruct {
	var envelope KafkaSourceEnvelopeStruct

	data := v.([]interface{})[0].(map[string]interface{})

	if upsert, ok := data["upsert"].(bool); ok {
		envelope.Upsert = upsert
		if options, ok := data["upsert_options"].([]interface{}); ok && len(options) > 0 {
			optionsData := options[0].(map[string]interface{})
			envelope.UpsertOptions = &UpsertOptionsStruct{}
			if valueDecodingErrors, ok := optionsData["value_decoding_errors"].([]interface{}); ok && len(valueDecodingErrors) > 0 {
				vdeData := valueDecodingErrors[0].(map[string]interface{})
				if inline, ok := vdeData["inline"].([]interface{}); ok && len(inline) > 0 {
					inlineData := inline[0].(map[string]interface{})
					if enabled, ok := inlineData["enabled"].(bool); ok {
						envelope.UpsertOptions.ValueDecodingErrors.Inline.Enabled = enabled
					}
					if alias, ok := inlineData["alias"].(string); ok {
						envelope.UpsertOptions.ValueDecodingErrors.Inline.Alias = alias
					}
				}
			}
		}
	}

	if debezium, ok := data["debezium"].(bool); ok {
		envelope.Debezium = debezium
	}

	if none, ok := data["none"].(bool); ok {
		envelope.None = none
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
	keyAlias         string
	headersAlias     string
	partitionAlias   string
	offsetAlias      string
	timestampAlias   string
	format           SourceFormatSpecStruct
	keyFormat        SourceFormatSpecStruct
	valueFormat      SourceFormatSpecStruct
	envelope         KafkaSourceEnvelopeStruct
	startOffset      []int
	startTimestamp   int
	exposeProgress   IdentifierSchemaStruct
}

func NewSourceKafkaBuilder(conn *sqlx.DB, obj MaterializeObject) *SourceKafkaBuilder {
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

func (b *SourceKafkaBuilder) IncludeKeyAlias(alias string) *SourceKafkaBuilder {
	b.includeKey = true
	b.keyAlias = alias
	return b
}

func (b *SourceKafkaBuilder) IncludeHeadersAlias(alias string) *SourceKafkaBuilder {
	b.includeHeaders = true
	b.headersAlias = alias
	return b
}

func (b *SourceKafkaBuilder) IncludePartitionAlias(alias string) *SourceKafkaBuilder {
	b.includePartition = true
	b.partitionAlias = alias
	return b
}

func (b *SourceKafkaBuilder) IncludeOffsetAlias(alias string) *SourceKafkaBuilder {
	b.includeOffset = true
	b.offsetAlias = alias
	return b
}

func (b *SourceKafkaBuilder) IncludeTimestampAlias(alias string) *SourceKafkaBuilder {
	b.includeTimestamp = true
	b.timestampAlias = alias
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

func (b *SourceKafkaBuilder) StartOffset(s []int) *SourceKafkaBuilder {
	b.startOffset = s
	return b
}

func (b *SourceKafkaBuilder) StartTimestamp(s int) *SourceKafkaBuilder {
	b.startTimestamp = s
	return b
}

func (b *SourceKafkaBuilder) ExposeProgress(e IdentifierSchemaStruct) *SourceKafkaBuilder {
	b.exposeProgress = e
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

	// Time-based Offsets
	if b.startTimestamp != 0 {
		q.WriteString(fmt.Sprintf(`, START TIMESTAMP %d`, b.startTimestamp))
	}
	if len(b.startOffset) > 0 {
		o := ""
		for _, v := range b.startOffset {
			if len(o) > 0 {
				o += ","
			}
			o += strconv.Itoa((v))
		}
		q.WriteString(fmt.Sprintf(`, START OFFSET (%s)`, o))
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

	if b.format.Json {
		q.WriteString(` FORMAT JSON`)
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

	if b.keyFormat.Json {
		q.WriteString(` KEY FORMAT JSON`)
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

	if b.valueFormat.Json {
		q.WriteString(` VALUE FORMAT JSON`)
	}

	// Metadata
	var i []string

	if !b.includeKey && b.keyAlias != "" {
		return fmt.Errorf("include_key_alias is set but include_key is false")
	}

	if b.includeKey {
		if b.keyAlias != "" {
			i = append(i, fmt.Sprintf("KEY AS %s", b.keyAlias))
		} else {
			i = append(i, "KEY")
		}
	}

	if !b.includeHeaders && b.headersAlias != "" {
		return fmt.Errorf("include_headers_alias is set but include_headers is false")
	}

	if b.includeHeaders {
		if b.headersAlias != "" {
			i = append(i, fmt.Sprintf("HEADERS AS %s", b.headersAlias))
		} else {
			i = append(i, "HEADERS")
		}
	}

	if !b.includePartition && b.partitionAlias != "" {
		return fmt.Errorf("include_partition_alias is set but include_partition is false")
	}

	if b.includePartition {
		if b.partitionAlias != "" {
			i = append(i, fmt.Sprintf("PARTITION AS %s", b.partitionAlias))
		} else {
			i = append(i, "PARTITION")
		}
	}

	if !b.includeOffset && b.offsetAlias != "" {
		return fmt.Errorf("include_offset_alias is set but include_offset is false")
	}

	if b.includeOffset {
		if b.offsetAlias != "" {
			i = append(i, fmt.Sprintf("OFFSET AS %s", b.offsetAlias))
		} else {
			i = append(i, "OFFSET")
		}
	}

	if !b.includeTimestamp && b.timestampAlias != "" {
		return fmt.Errorf("include_timestamp_alias is set but include_timestamp is false")
	}

	if b.includeTimestamp {
		if b.timestampAlias != "" {
			i = append(i, fmt.Sprintf("TIMESTAMP AS %s", b.timestampAlias))
		} else {
			i = append(i, "TIMESTAMP")
		}
	}

	if len(i) > 0 {
		o := strings.Join(i[:], ", ")
		q.WriteString(fmt.Sprintf(` INCLUDE %s`, o))
	}

	if b.envelope.Debezium {
		q.WriteString(` ENVELOPE DEBEZIUM`)
	}

	if b.envelope.Upsert {
		q.WriteString(" ENVELOPE UPSERT")
		if b.envelope.UpsertOptions != nil {
			inlineOptions := b.envelope.UpsertOptions.ValueDecodingErrors.Inline
			if inlineOptions.Enabled {
				q.WriteString(" (VALUE DECODING ERRORS = (INLINE")
				if inlineOptions.Alias != "" {
					q.WriteString(" AS ")
					q.WriteString(inlineOptions.Alias)
				}
				q.WriteString("))")
			}
		}
	}

	if b.envelope.None {
		q.WriteString(` ENVELOPE NONE`)
	}

	if b.exposeProgress.Name != "" {
		q.WriteString(fmt.Sprintf(` EXPOSE PROGRESS AS %s`, b.exposeProgress.QualifiedName()))
	}

	q.WriteString(`;`)
	return b.ddl.exec(q.String())
}
