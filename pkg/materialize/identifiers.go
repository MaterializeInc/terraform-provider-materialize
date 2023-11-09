package materialize

// Any Materialize Object. Will contain name and database and schema
// If no database or schema is provided will inherit those values
type IdentifierSchemaStruct struct {
	Name         string
	SchemaName   string
	DatabaseName string
}

func GetIdentifierSchemaStruct(v interface{}) IdentifierSchemaStruct {
	var i IdentifierSchemaStruct
	u := v.([]interface{})[0].(map[string]interface{})
	if v, ok := u["name"]; ok {
		i.Name = v.(string)
	}
	if v, ok := u["schema_name"]; ok {
		i.SchemaName = v.(string)
	}
	if v, ok := u["database_name"]; ok {
		i.DatabaseName = v.(string)
	}
	return i
}

func (i *IdentifierSchemaStruct) QualifiedName() string {
	return QualifiedName(i.DatabaseName, i.SchemaName, i.Name)
}
