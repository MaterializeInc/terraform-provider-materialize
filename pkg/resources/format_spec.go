package resources

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func FormatSpecSchema(elem string, description string, isRequired bool, isOptional bool) *schema.Schema {
	return &schema.Schema{
		Type: schema.TypeList,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"avro": {
					Description: "Avro format.",
					Type:        schema.TypeList,
					Optional:    true,
					ForceNew:    true,
					MaxItems:    1,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"schema_registry_connection": IdentifierSchema("schema_registry_connection", "The name of a schema registry connection.", true, false),
							"key_strategy": {
								Description:  "How Materialize will define the Avro schema reader key strategy.",
								Type:         schema.TypeString,
								Optional:     true,
								ForceNew:     true,
								ValidateFunc: validation.StringInSlice(strategy, true),
							},
							"value_strategy": {
								Description:  "How Materialize will define the Avro schema reader value strategy.",
								Type:         schema.TypeString,
								Optional:     true,
								ForceNew:     true,
								ValidateFunc: validation.StringInSlice(strategy, true),
							},
						},
					},
				},
				"protobuf": {
					Description: "Protobuf format.",
					Type:        schema.TypeList,
					Optional:    true,
					ForceNew:    true,
					MaxItems:    1,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"schema_registry_connection": IdentifierSchema("schema_registry_connection", "The name of a schema registry connection.", true, false),
							"message": {
								Description: "The name of the Protobuf message to use for the source.",
								Type:        schema.TypeString,
								Required:    true,
								ForceNew:    true,
							},
						},
					},
				},
				"csv": {
					Description: "CSV format.",
					Type:        schema.TypeList,
					Optional:    true,
					ForceNew:    true,
					MaxItems:    2,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"columns": {
								Description: "The columns to use for the source.",
								Type:        schema.TypeInt,
								Optional:    true,
								ForceNew:    true,
							},
							"delimited_by": {
								Description: "The delimiter to use for the source.",
								Type:        schema.TypeString,
								Optional:    true,
								ForceNew:    true,
							},
							"header": {
								Description: "The number of columns and the name of each column using the header row.",
								Type:        schema.TypeList,
								Elem:        &schema.Schema{Type: schema.TypeString},
								Optional:    true,
								ForceNew:    true,
							},
						},
					},
				},
				"json": {
					Description: "JSON format.",
					Type:        schema.TypeBool,
					Optional:    true,
					ForceNew:    true,
				},
				"text": {
					Description: "Text format.",
					Type:        schema.TypeBool,
					Optional:    true,
					ForceNew:    true,
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
