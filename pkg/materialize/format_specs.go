package materialize

type AvroFormatSpec struct {
	SchemaRegistryConnection IdentifierSchemaStruct
	KeyStrategy              string
	ValueStrategy            string
}

type ProtobufFormatSpec struct {
	SchemaRegistryConnection IdentifierSchemaStruct
	MessageName              string
}

type CsvFormatSpec struct {
	Columns     int
	DelimitedBy string
	Header      []string
}

type SourceFormatSpecStruct struct {
	Avro     *AvroFormatSpec
	Protobuf *ProtobufFormatSpec
	Csv      *CsvFormatSpec
	Bytes    bool
	Text     bool
}

type SinkAvroFormatSpec struct {
	SchemaRegistryConnection IdentifierSchemaStruct
	AvroKeyFullname          string
	AvroValueFullname        string
}

type SinkFormatSpecStruct struct {
	Avro *SinkAvroFormatSpec
	Json bool
}

func GetFormatSpecStruc(v interface{}) SourceFormatSpecStruct {
	var format SourceFormatSpecStruct
	var databaseName string
	var schemaName string

	u := v.([]interface{})[0].(map[string]interface{})
	if avro, ok := u["avro"]; ok && avro != nil && len(avro.([]interface{})) > 0 {
		if csr, ok := avro.([]interface{})[0].(map[string]interface{})["schema_registry_connection"]; ok {
			if databaseName, ok = avro.([]interface{})[0].(map[string]interface{})["database_name"].(string); !ok {
				databaseName = "materialize"
			}
			if schemaName, ok = avro.([]interface{})[0].(map[string]interface{})["schema_name"].(string); !ok {
				schemaName = "public"
			}
			key := avro.([]interface{})[0].(map[string]interface{})["key_strategy"].(string)
			value := avro.([]interface{})[0].(map[string]interface{})["value_strategy"].(string)
			format.Avro = &AvroFormatSpec{
				SchemaRegistryConnection: GetIdentifierSchemaStruct(databaseName, schemaName, csr),
				KeyStrategy:              key,
				ValueStrategy:            value,
			}
		}
	}
	if protobuf, ok := u["protobuf"]; ok && protobuf != nil && len(protobuf.([]interface{})) > 0 {
		if csr, ok := protobuf.([]interface{})[0].(map[string]interface{})["schema_registry_connection"]; ok {
			if databaseName, ok = protobuf.([]interface{})[0].(map[string]interface{})["database_name"].(string); !ok {
				databaseName = "materialize"
			}
			if schemaName, ok = protobuf.([]interface{})[0].(map[string]interface{})["schema_name"].(string); !ok {
				schemaName = "public"
			}
			message := protobuf.([]interface{})[0].(map[string]interface{})["message_name"].(string)
			format.Protobuf = &ProtobufFormatSpec{
				SchemaRegistryConnection: GetIdentifierSchemaStruct(databaseName, schemaName, csr),
				MessageName:              message,
			}
		}
	}
	if v, ok := u["csv"]; ok && v != nil && len(v.([]interface{})) > 0 {
		csv := v.([]interface{})[0].(map[string]interface{})
		format.Csv = &CsvFormatSpec{
			Columns:     csv["columns"].(int),
			DelimitedBy: csv["delimited_by"].(string),
			Header:      csv["header"].([]string),
		}
	}
	if v, ok := u["bytes"]; ok {
		format.Bytes = v.(bool)
	}
	if v, ok := u["text"]; ok {
		format.Text = v.(bool)
	}
	return format
}

func GetSinkFormatSpecStruc(v interface{}) SinkFormatSpecStruct {
	var format SinkFormatSpecStruct
	var databaseName string
	var schemaName string

	u := v.([]interface{})[0].(map[string]interface{})
	if avro, ok := u["avro"]; ok && avro != nil && len(avro.([]interface{})) > 0 {
		if csr, ok := avro.([]interface{})[0].(map[string]interface{})["schema_registry_connection"]; ok {
			if databaseName, ok = avro.([]interface{})[0].(map[string]interface{})["database_name"].(string); !ok {
				databaseName = "materialize"
			}
			if schemaName, ok = avro.([]interface{})[0].(map[string]interface{})["schema_name"].(string); !ok {
				schemaName = "public"
			}
			key := avro.([]interface{})[0].(map[string]interface{})["avro_key_fullname"].(string)
			value := avro.([]interface{})[0].(map[string]interface{})["avro_value_fullname"].(string)
			format.Avro = &SinkAvroFormatSpec{
				SchemaRegistryConnection: GetIdentifierSchemaStruct(databaseName, schemaName, csr),
				AvroKeyFullname:          key,
				AvroValueFullname:        value,
			}
		}
	}
	if v, ok := u["json"]; ok {
		format.Json = v.(bool)
	}
	return format
}
