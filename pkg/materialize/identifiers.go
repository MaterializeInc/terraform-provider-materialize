package materialize

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

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

func IdentifierSchema(elem string, description string, isRequired bool, isOptional bool) *schema.Schema {
	return &schema.Schema{
		Type: schema.TypeList,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"name": {
					Description: fmt.Sprintf("The %s name.", elem),
					Type:        schema.TypeString,
					Required:    true,
				},
				"schema_name": {
					Description: fmt.Sprintf("The %s schema name.", elem),
					Type:        schema.TypeString,
					Optional:    true,
				},
				"database_name": {
					Description: fmt.Sprintf("The %s database name.", elem),
					Type:        schema.TypeString,
					Optional:    true,
				},
			},
		},
		Required:    isRequired,
		Optional:    isOptional,
		MinItems:    1,
		MaxItems:    1,
		ForceNew:    true,
		Description: description,
	}
}
