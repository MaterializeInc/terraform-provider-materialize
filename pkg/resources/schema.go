package resources

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

const (
	defaultSchema   = "public"
	defaultDatabase = "materialize"
)

func SchemaResourceName(resource string, required, forceNew bool) *schema.Schema {
	return &schema.Schema{
		Type:        schema.TypeString,
		Description: fmt.Sprintf("The identifier for the %s.", resource),
		Required:    required,
		Optional:    !required,
		ForceNew:    forceNew,
	}
}

func SchemaResourceSchemaName(resource string, required bool) *schema.Schema {
	return &schema.Schema{
		Type:        schema.TypeString,
		Description: fmt.Sprintf("The identifier for the %s schema.", resource),
		Required:    required,
		Optional:    !required,
		ForceNew:    true,
		Default:     defaultSchema,
	}
}

func SchemaResourceDatabaseName(resource string, required bool) *schema.Schema {
	return &schema.Schema{
		Type:        schema.TypeString,
		Description: fmt.Sprintf("The identifier for the %s database.", resource),
		Required:    required,
		Optional:    !required,
		ForceNew:    true,
		DefaultFunc: schema.EnvDefaultFunc("MZ_DATABASE", defaultDatabase),
	}
}

func SchemaResourceQualifiedName(resource string) *schema.Schema {
	return &schema.Schema{
		Type:        schema.TypeString,
		Description: fmt.Sprintf("The fully qualified name of the %s.", resource),
		Computed:    true,
	}
}

func SchemaSize(resource string) *schema.Schema {
	return &schema.Schema{
		Type:         schema.TypeString,
		Description:  fmt.Sprintf("The size of the %s.", resource),
		Required:     true,
		ForceNew:     true,
		ValidateFunc: validation.StringInSlice(append(replicaSizes, localSizes...), true),
	}
}

func IdentifierSchema(elem, description string, required bool) *schema.Schema {
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
		Required:    required,
		Optional:    !required,
		MinItems:    1,
		MaxItems:    1,
		ForceNew:    true,
		Description: description,
	}
}
