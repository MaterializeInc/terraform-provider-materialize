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

func NameSchema(resource string, required, forceNew bool) *schema.Schema {
	return &schema.Schema{
		Type:        schema.TypeString,
		Description: fmt.Sprintf("The identifier for the %s.", resource),
		Required:    required,
		Optional:    !required,
		ForceNew:    forceNew,
	}
}

func SchemaNameSchema(resource string, required bool) *schema.Schema {
	return &schema.Schema{
		Type:        schema.TypeString,
		Description: fmt.Sprintf("The identifier for the %s schema.", resource),
		Required:    required,
		Optional:    !required,
		ForceNew:    true,
		Default:     defaultSchema,
	}
}

func DatabaseNameSchema(resource string, required bool) *schema.Schema {
	return &schema.Schema{
		Type:        schema.TypeString,
		Description: fmt.Sprintf("The identifier for the %s database.", resource),
		Required:    required,
		Optional:    !required,
		ForceNew:    true,
		DefaultFunc: schema.EnvDefaultFunc("MZ_DATABASE", defaultDatabase),
	}
}

func QualifiedNameSchema(resource string) *schema.Schema {
	return &schema.Schema{
		Type:        schema.TypeString,
		Description: fmt.Sprintf("The fully qualified name of the %s.", resource),
		Computed:    true,
	}
}

func SizeSchema(resource string) *schema.Schema {
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

func ValueSecretSchema(elem string, description string, isRequired bool, isOptional bool) *schema.Schema {
	return &schema.Schema{
		Type: schema.TypeList,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"text": {
					Description:   fmt.Sprintf("The %s text value.", elem),
					Type:          schema.TypeString,
					Optional:      true,
					ConflictsWith: []string{fmt.Sprintf("%s.0.secret", elem)},
				},
				"secret": IdentifierSchema(elem, fmt.Sprintf("The %s secret value.", elem), false),
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
