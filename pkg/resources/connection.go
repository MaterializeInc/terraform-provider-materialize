package resources

import (
	"context"
	"fmt"
	"terraform-materialize/pkg/materialize"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/jmoiron/sqlx"
)

func connectionRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*sqlx.DB)
	i := d.Id()
	q := materialize.ReadConnectionParams(i)

	var name, schema, database, connection_type *string
	if err := conn.QueryRowx(q).Scan(&name, &schema, &database, &connection_type); err != nil {
		return diag.FromErr(err)
	}

	d.SetId(i)

	if err := d.Set("name", name); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("schema_name", schema); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("database_name", database); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("connection_type", connection_type); err != nil {
		return diag.FromErr(err)
	}

	b := materialize.Connection{ConnectionName: *name, SchemaName: *schema, DatabaseName: *database}
	if err := d.Set("qualified_sql_name", b.QualifiedName()); err != nil {
		return diag.FromErr(err)
	}

	return nil
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
