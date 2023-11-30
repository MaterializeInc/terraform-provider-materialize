package materialize

type AvroDocType struct {
	Object IdentifierSchemaStruct
	Doc    string
	Key    bool
	Value  bool
}

type AvroDocColumn struct {
	Object IdentifierSchemaStruct
	Column string
	Doc    string
	Key    bool
	Value  bool
}

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
	Json     bool
}

type SinkAvroFormatSpec struct {
	SchemaRegistryConnection IdentifierSchemaStruct
	AvroKeyFullname          string
	AvroValueFullname        string
	DocType                  AvroDocType
	DocColumn                []AvroDocColumn
}

type SinkFormatSpecStruct struct {
	Avro *SinkAvroFormatSpec
	Json bool
}

func GetFormatSpecStruc(v interface{}) SourceFormatSpecStruct {
	var format SourceFormatSpecStruct

	u := v.([]interface{})[0].(map[string]interface{})
	if avro, ok := u["avro"]; ok && avro != nil && len(avro.([]interface{})) > 0 {
		if csr, ok := avro.([]interface{})[0].(map[string]interface{})["schema_registry_connection"]; ok {
			key := avro.([]interface{})[0].(map[string]interface{})["key_strategy"].(string)
			value := avro.([]interface{})[0].(map[string]interface{})["value_strategy"].(string)
			format.Avro = &AvroFormatSpec{
				SchemaRegistryConnection: GetIdentifierSchemaStruct(csr),
				KeyStrategy:              key,
				ValueStrategy:            value,
			}
		}
	}
	if protobuf, ok := u["protobuf"]; ok && protobuf != nil && len(protobuf.([]interface{})) > 0 {
		if csr, ok := protobuf.([]interface{})[0].(map[string]interface{})["schema_registry_connection"]; ok {
			message := protobuf.([]interface{})[0].(map[string]interface{})["message_name"].(string)
			format.Protobuf = &ProtobufFormatSpec{
				SchemaRegistryConnection: GetIdentifierSchemaStruct(csr),
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
	if v, ok := u["json"]; ok {
		format.Json = v.(bool)
	}
	return format
}

func GetSinkFormatSpecStruc(v interface{}) SinkFormatSpecStruct {
	var format SinkFormatSpecStruct

	u := v.([]interface{})[0].(map[string]interface{})
	if avro, ok := u["avro"]; ok && avro != nil && len(avro.([]interface{})) > 0 {
		if csr, ok := avro.([]interface{})[0].(map[string]interface{})["schema_registry_connection"]; ok {
			key := avro.([]interface{})[0].(map[string]interface{})["avro_key_fullname"].(string)
			value := avro.([]interface{})[0].(map[string]interface{})["avro_value_fullname"].(string)

			var docType AvroDocType
			if adt, ok := avro.([]interface{})[0].(map[string]interface{})["avro_doc_type"].([]interface{}); ok {
				if v, ok := adt[0].(map[string]interface{})["object"]; ok {
					docType.Object = GetIdentifierSchemaStruct(v)
				}
				if v, ok := adt[0].(map[string]interface{})["doc"]; ok {
					docType.Doc = v.(string)
				}
				if v, ok := adt[0].(map[string]interface{})["key"]; ok {
					docType.Key = v.(bool)
				}
				if v, ok := adt[0].(map[string]interface{})["value"]; ok {
					docType.Value = v.(bool)
				}
			}

			var docColumn []AvroDocColumn
			if adc, ok := avro.([]interface{})[0].(map[string]interface{})["avro_doc_column"]; ok {
				for _, column := range adc.([]interface{}) {
					docColumn = append(docColumn, AvroDocColumn{
						Object: GetIdentifierSchemaStruct(column.(map[string]interface{})["object"]),
						Column: column.(map[string]interface{})["column"].(string),
						Doc:    column.(map[string]interface{})["doc"].(string),
						Key:    column.(map[string]interface{})["key"].(bool),
						Value:  column.(map[string]interface{})["value"].(bool),
					})
				}
			}

			format.Avro = &SinkAvroFormatSpec{
				SchemaRegistryConnection: GetIdentifierSchemaStruct(csr),
				AvroKeyFullname:          key,
				AvroValueFullname:        value,
				DocType:                  docType,
				DocColumn:                docColumn,
			}
		}
	}
	if v, ok := u["json"]; ok {
		format.Json = v.(bool)
	}
	return format
}
