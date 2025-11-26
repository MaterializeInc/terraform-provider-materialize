package materialize

import (
	"fmt"
	"strings"
)

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

// ColumnReferenceStruct represents a column reference with optional schema and table
type ColumnReferenceStruct struct {
	ColumnName string
	TableName  string
	SchemaName string
}

func GetColumnReferenceStruct(v interface{}) ColumnReferenceStruct {
	var c ColumnReferenceStruct
	u := v.(map[string]interface{})
	if v, ok := u["column_name"]; ok {
		c.ColumnName = v.(string)
	}
	if v, ok := u["table_name"]; ok {
		c.TableName = v.(string)
	}
	if v, ok := u["schema_name"]; ok {
		c.SchemaName = v.(string)
	}
	return c
}

func GetColumnReferenceStructSlice(attrName string, v []interface{}) ([]ColumnReferenceStruct, error) {
	var columns []ColumnReferenceStruct
	for _, item := range v {
		col := GetColumnReferenceStruct(item.(map[string]interface{}))
		if col.ColumnName == "" {
			return nil, fmt.Errorf("column_name is required for %s", attrName)
		}
		columns = append(columns, col)
	}
	return columns, nil
}

func (c *ColumnReferenceStruct) QualifiedColumnName() string {
	var parts []string
	if c.SchemaName != "" {
		parts = append(parts, QuoteIdentifier(c.SchemaName))
	}
	if c.TableName != "" {
		parts = append(parts, QuoteIdentifier(c.TableName))
	}
	parts = append(parts, QuoteIdentifier(c.ColumnName))
	return strings.Join(parts, ".")
}
