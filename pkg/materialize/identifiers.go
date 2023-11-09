package materialize

// Any Materialize Object. Will contain name and database and schema
// If no database or schema is provided will inherit those values
type IdentifierSchemaStruct struct {
	Name         string
	SchemaName   string
	DatabaseName string
}

func GetIdentifierSchemaStruct(databaseName string, schemaName string, v interface{}) IdentifierSchemaStruct {
	var conn IdentifierSchemaStruct
	u := v.([]interface{})[0].(map[string]interface{})
	if v, ok := u["name"]; ok {
		conn.Name = v.(string)
	}
	if v, ok := u["schema_name"]; ok {
		conn.SchemaName = v.(string)
	}
	if v, ok := u["database_name"]; ok {
		conn.DatabaseName = v.(string)
	}
	return conn
}

func (i *IdentifierSchemaStruct) QualifiedName() string {
	return QualifiedName(i.DatabaseName, i.SchemaName, i.Name)
}
