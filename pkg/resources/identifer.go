package resources

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

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
