package materialize

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
	if v, ok := u["schema_name"]; ok && v.(string) != "" {
		conn.SchemaName = v.(string)
	} else {
		conn.SchemaName = schemaName
	}
	if v, ok := u["database_name"]; ok && v.(string) != "" {
		conn.DatabaseName = v.(string)
	} else {
		conn.DatabaseName = databaseName
	}
	return conn
}
