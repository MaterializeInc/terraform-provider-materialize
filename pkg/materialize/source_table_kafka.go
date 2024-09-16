package materialize

import (
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"
)

type SourceTableKafkaBuilder struct {
	*SourceTableBuilder
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
	exposeProgress   IdentifierSchemaStruct
}

func NewSourceTableKafkaBuilder(conn *sqlx.DB, obj MaterializeObject) *SourceTableKafkaBuilder {
	return &SourceTableKafkaBuilder{
		SourceTableBuilder: NewSourceTableBuilder(conn, obj),
	}
}

func (b *SourceTableKafkaBuilder) IncludeKey() *SourceTableKafkaBuilder {
	b.includeKey = true
	return b
}

func (b *SourceTableKafkaBuilder) IncludeHeaders() *SourceTableKafkaBuilder {
	b.includeHeaders = true
	return b
}

func (b *SourceTableKafkaBuilder) IncludePartition() *SourceTableKafkaBuilder {
	b.includePartition = true
	return b
}

func (b *SourceTableKafkaBuilder) IncludeOffset() *SourceTableKafkaBuilder {
	b.includeOffset = true
	return b
}

func (b *SourceTableKafkaBuilder) IncludeTimestamp() *SourceTableKafkaBuilder {
	b.includeTimestamp = true
	return b
}

func (b *SourceTableKafkaBuilder) IncludeKeyAlias(alias string) *SourceTableKafkaBuilder {
	b.includeKey = true
	b.keyAlias = alias
	return b
}

func (b *SourceTableKafkaBuilder) IncludeHeadersAlias(alias string) *SourceTableKafkaBuilder {
	b.includeHeaders = true
	b.headersAlias = alias
	return b
}

func (b *SourceTableKafkaBuilder) IncludePartitionAlias(alias string) *SourceTableKafkaBuilder {
	b.includePartition = true
	b.partitionAlias = alias
	return b
}

func (b *SourceTableKafkaBuilder) IncludeOffsetAlias(alias string) *SourceTableKafkaBuilder {
	b.includeOffset = true
	b.offsetAlias = alias
	return b
}

func (b *SourceTableKafkaBuilder) IncludeTimestampAlias(alias string) *SourceTableKafkaBuilder {
	b.includeTimestamp = true
	b.timestampAlias = alias
	return b
}

func (b *SourceTableKafkaBuilder) Format(f SourceFormatSpecStruct) *SourceTableKafkaBuilder {
	b.format = f
	return b
}

func (b *SourceTableKafkaBuilder) Envelope(e KafkaSourceEnvelopeStruct) *SourceTableKafkaBuilder {
	b.envelope = e
	return b
}

func (b *SourceTableKafkaBuilder) KeyFormat(k SourceFormatSpecStruct) *SourceTableKafkaBuilder {
	b.keyFormat = k
	return b
}

func (b *SourceTableKafkaBuilder) ValueFormat(v SourceFormatSpecStruct) *SourceTableKafkaBuilder {
	b.valueFormat = v
	return b
}

func (b *SourceTableKafkaBuilder) ExposeProgress(e IdentifierSchemaStruct) *SourceTableKafkaBuilder {
	b.exposeProgress = e
	return b
}

func (b *SourceTableKafkaBuilder) Create() error {
	return b.BaseCreate("kafka", func() string {
		q := strings.Builder{}
		var options []string

		// Format
		if b.format.Avro != nil {
			if b.format.Avro.SchemaRegistryConnection.Name != "" {
				options = append(options, fmt.Sprintf(`FORMAT AVRO USING CONFLUENT SCHEMA REGISTRY CONNECTION %s`, QualifiedName(b.format.Avro.SchemaRegistryConnection.DatabaseName, b.format.Avro.SchemaRegistryConnection.SchemaName, b.format.Avro.SchemaRegistryConnection.Name)))
			}
			if b.format.Avro.KeyStrategy != "" {
				options = append(options, fmt.Sprintf(`KEY STRATEGY %s`, b.format.Avro.KeyStrategy))
			}
			if b.format.Avro.ValueStrategy != "" {
				options = append(options, fmt.Sprintf(`VALUE STRATEGY %s`, b.format.Avro.ValueStrategy))
			}
		}

		if b.format.Protobuf != nil {
			if b.format.Protobuf.SchemaRegistryConnection.Name != "" && b.format.Protobuf.MessageName != "" {
				options = append(options, fmt.Sprintf(`FORMAT PROTOBUF MESSAGE '%s' USING CONFLUENT SCHEMA REGISTRY CONNECTION %s`, b.format.Protobuf.MessageName, QualifiedName(b.format.Protobuf.SchemaRegistryConnection.DatabaseName, b.format.Protobuf.SchemaRegistryConnection.SchemaName, b.format.Protobuf.SchemaRegistryConnection.Name)))
			} else if b.format.Protobuf.SchemaRegistryConnection.Name != "" {
				options = append(options, fmt.Sprintf(`FORMAT PROTOBUF USING CONFLUENT SCHEMA REGISTRY CONNECTION %s`, QualifiedName(b.format.Protobuf.SchemaRegistryConnection.DatabaseName, b.format.Protobuf.SchemaRegistryConnection.SchemaName, b.format.Protobuf.SchemaRegistryConnection.Name)))
			}
		}

		if b.format.Csv != nil {
			if b.format.Csv.Columns > 0 {
				options = append(options, fmt.Sprintf(`FORMAT CSV WITH %d COLUMNS`, b.format.Csv.Columns))
			}
			if b.format.Csv.Header != nil {
				options = append(options, fmt.Sprintf(`FORMAT CSV WITH HEADER ( %s )`, strings.Join(b.format.Csv.Header, ", ")))
			}
			if b.format.Csv.DelimitedBy != "" {
				options = append(options, fmt.Sprintf(`DELIMITER '%s'`, b.format.Csv.DelimitedBy))
			}
		}

		if b.format.Bytes {
			options = append(options, `FORMAT BYTES`)
		}
		if b.format.Text {
			options = append(options, `FORMAT TEXT`)
		}
		if b.format.Json {
			options = append(options, `FORMAT JSON`)
		}

		// Key Format
		if b.keyFormat.Avro != nil {
			if b.keyFormat.Avro.SchemaRegistryConnection.Name != "" {
				options = append(options, fmt.Sprintf(`KEY FORMAT AVRO USING CONFLUENT SCHEMA REGISTRY CONNECTION %s`, QualifiedName(b.keyFormat.Avro.SchemaRegistryConnection.DatabaseName, b.keyFormat.Avro.SchemaRegistryConnection.SchemaName, b.keyFormat.Avro.SchemaRegistryConnection.Name)))
			}
			if b.keyFormat.Avro.KeyStrategy != "" {
				options = append(options, fmt.Sprintf(`KEY STRATEGY %s`, b.keyFormat.Avro.KeyStrategy))
			}
			if b.keyFormat.Avro.ValueStrategy != "" {
				options = append(options, fmt.Sprintf(`VALUE STRATEGY %s`, b.keyFormat.Avro.ValueStrategy))
			}
		}

		if b.keyFormat.Protobuf != nil {
			if b.keyFormat.Protobuf.SchemaRegistryConnection.Name != "" && b.keyFormat.Protobuf.MessageName != "" {
				options = append(options, fmt.Sprintf(`KEY FORMAT PROTOBUF MESSAGE '%s' USING CONFLUENT SCHEMA REGISTRY CONNECTION %s`, b.keyFormat.Protobuf.MessageName, QualifiedName(b.keyFormat.Protobuf.SchemaRegistryConnection.DatabaseName, b.keyFormat.Protobuf.SchemaRegistryConnection.SchemaName, b.keyFormat.Protobuf.SchemaRegistryConnection.Name)))
			} else if b.keyFormat.Protobuf.SchemaRegistryConnection.Name != "" {
				options = append(options, fmt.Sprintf(`KEY FORMAT PROTOBUF USING CONFLUENT SCHEMA REGISTRY CONNECTION %s`, QualifiedName(b.keyFormat.Protobuf.SchemaRegistryConnection.DatabaseName, b.keyFormat.Protobuf.SchemaRegistryConnection.SchemaName, b.keyFormat.Protobuf.SchemaRegistryConnection.Name)))
			}
		}

		if b.keyFormat.Csv != nil {
			if b.keyFormat.Csv.Columns > 0 {
				options = append(options, fmt.Sprintf(`KEY FORMAT CSV WITH %d COLUMNS`, b.keyFormat.Csv.Columns))
			}
			if b.keyFormat.Csv.Header != nil {
				options = append(options, fmt.Sprintf(`KEY FORMAT CSV WITH HEADER ( %s )`, strings.Join(b.keyFormat.Csv.Header, ", ")))
			}
			if b.keyFormat.Csv.DelimitedBy != "" {
				options = append(options, fmt.Sprintf(`KEY DELIMITER '%s'`, b.keyFormat.Csv.DelimitedBy))
			}
		}

		if b.keyFormat.Bytes {
			options = append(options, `KEY FORMAT BYTES`)
		}
		if b.keyFormat.Text {
			options = append(options, `KEY FORMAT TEXT`)
		}
		if b.keyFormat.Json {
			options = append(options, `KEY FORMAT JSON`)
		}

		// Value Format
		if b.valueFormat.Avro != nil {
			if b.valueFormat.Avro.SchemaRegistryConnection.Name != "" {
				options = append(options, fmt.Sprintf(`VALUE FORMAT AVRO USING CONFLUENT SCHEMA REGISTRY CONNECTION %s`, QualifiedName(b.valueFormat.Avro.SchemaRegistryConnection.DatabaseName, b.valueFormat.Avro.SchemaRegistryConnection.SchemaName, b.valueFormat.Avro.SchemaRegistryConnection.Name)))
			}
			if b.valueFormat.Avro.KeyStrategy != "" {
				options = append(options, fmt.Sprintf(`VALUE STRATEGY %s`, b.valueFormat.Avro.KeyStrategy))
			}
			if b.valueFormat.Avro.ValueStrategy != "" {
				options = append(options, fmt.Sprintf(`VALUE STRATEGY %s`, b.valueFormat.Avro.ValueStrategy))
			}
		}

		if b.valueFormat.Protobuf != nil {
			if b.valueFormat.Protobuf.SchemaRegistryConnection.Name != "" && b.valueFormat.Protobuf.MessageName != "" {
				options = append(options, fmt.Sprintf(`VALUE FORMAT PROTOBUF MESSAGE '%s' USING CONFLUENT SCHEMA REGISTRY CONNECTION %s`, b.valueFormat.Protobuf.MessageName, QualifiedName(b.valueFormat.Protobuf.SchemaRegistryConnection.DatabaseName, b.valueFormat.Protobuf.SchemaRegistryConnection.SchemaName, b.valueFormat.Protobuf.SchemaRegistryConnection.Name)))
			} else if b.valueFormat.Protobuf.SchemaRegistryConnection.Name != "" {
				options = append(options, fmt.Sprintf(`VALUE FORMAT PROTOBUF USING CONFLUENT SCHEMA REGISTRY CONNECTION %s`, QualifiedName(b.valueFormat.Protobuf.SchemaRegistryConnection.DatabaseName, b.valueFormat.Protobuf.SchemaRegistryConnection.SchemaName, b.valueFormat.Protobuf.SchemaRegistryConnection.Name)))
			}
		}

		if b.valueFormat.Csv != nil {
			if b.valueFormat.Csv.Columns > 0 {
				options = append(options, fmt.Sprintf(`VALUE FORMAT CSV WITH %d COLUMNS`, b.valueFormat.Csv.Columns))
			}
			if b.valueFormat.Csv.Header != nil {
				options = append(options, fmt.Sprintf(`VALUE FORMAT CSV WITH HEADER ( %s )`, strings.Join(b.valueFormat.Csv.Header, ", ")))
			}
			if b.valueFormat.Csv.DelimitedBy != "" {
				options = append(options, fmt.Sprintf(`VALUE DELIMITER '%s'`, b.valueFormat.Csv.DelimitedBy))
			}
		}

		if b.valueFormat.Bytes {
			options = append(options, `VALUE FORMAT BYTES`)
		}
		if b.valueFormat.Text {
			options = append(options, `VALUE FORMAT TEXT`)
		}
		if b.valueFormat.Json {
			options = append(options, `VALUE FORMAT JSON`)
		}

		// Metadata
		var metadataOptions []string
		if b.includeKey {
			if b.keyAlias != "" {
				metadataOptions = append(metadataOptions, fmt.Sprintf("KEY AS %s", b.keyAlias))
			} else {
				metadataOptions = append(metadataOptions, "KEY")
			}
		}
		if b.includeHeaders {
			if b.headersAlias != "" {
				metadataOptions = append(metadataOptions, fmt.Sprintf("HEADERS AS %s", b.headersAlias))
			} else {
				metadataOptions = append(metadataOptions, "HEADERS")
			}
		}
		if b.includePartition {
			if b.partitionAlias != "" {
				metadataOptions = append(metadataOptions, fmt.Sprintf("PARTITION AS %s", b.partitionAlias))
			} else {
				metadataOptions = append(metadataOptions, "PARTITION")
			}
		}
		if b.includeOffset {
			if b.offsetAlias != "" {
				metadataOptions = append(metadataOptions, fmt.Sprintf("OFFSET AS %s", b.offsetAlias))
			} else {
				metadataOptions = append(metadataOptions, "OFFSET")
			}
		}
		if b.includeTimestamp {
			if b.timestampAlias != "" {
				metadataOptions = append(metadataOptions, fmt.Sprintf("TIMESTAMP AS %s", b.timestampAlias))
			} else {
				metadataOptions = append(metadataOptions, "TIMESTAMP")
			}
		}
		if len(metadataOptions) > 0 {
			options = append(options, fmt.Sprintf(`INCLUDE %s`, strings.Join(metadataOptions, ", ")))
		}

		// Envelope
		if b.envelope.Debezium {
			options = append(options, `ENVELOPE DEBEZIUM`)
		}
		if b.envelope.Upsert {
			upsertOption := "ENVELOPE UPSERT"
			if b.envelope.UpsertOptions != nil {
				inlineOptions := b.envelope.UpsertOptions.ValueDecodingErrors.Inline
				if inlineOptions.Enabled {
					upsertOption += " (VALUE DECODING ERRORS = (INLINE"
					if inlineOptions.Alias != "" {
						upsertOption += fmt.Sprintf(" AS %s", inlineOptions.Alias)
					}
					upsertOption += "))"
				}
			}
			options = append(options, upsertOption)
		}
		if b.envelope.None {
			options = append(options, `ENVELOPE NONE`)
		}

		// Expose Progress
		if b.exposeProgress.Name != "" {
			options = append(options, fmt.Sprintf(`EXPOSE PROGRESS AS %s`, b.exposeProgress.QualifiedName()))
		}

		q.WriteString(strings.Join(options, " "))
		return " " + q.String()
	})
}
