package materialize

// Any Materialize Object. Will contain name and optionally database and schema
type ObjectSchemaStruct struct {
	Name         string
	SchemaName   string
	DatabaseName string
}

func GetObjectSchemaStruct(v interface{}) ObjectSchemaStruct {
	var conn ObjectSchemaStruct
	u := v.([]interface{})[0].(map[string]interface{})
	if v, ok := u["name"]; ok {
		conn.Name = v.(string)
	}
	if v, ok := u["schema_name"]; ok && v.(string) != "" {
		conn.SchemaName = v.(string)
	}

	if v, ok := u["database_name"]; ok && v.(string) != "" {
		conn.DatabaseName = v.(string)
	}
	return conn
}

func (g *ObjectSchemaStruct) QualifiedName() string {
	fields := []string{}

	if g.DatabaseName != "" {
		fields = append(fields, g.DatabaseName)
	}

	if g.SchemaName != "" {
		fields = append(fields, g.SchemaName)
	}

	fields = append(fields, g.Name)
	return QualifiedName(fields...)
}
