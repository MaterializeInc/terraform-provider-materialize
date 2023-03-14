package resources

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/jmoiron/sqlx"
)

func readConnectionId(name, schema, database string) string {
	return fmt.Sprintf(`
		SELECT mz_connections.id
		FROM mz_connections
		JOIN mz_schemas
			ON mz_connections.schema_id = mz_schemas.id
		JOIN mz_databases
			ON mz_schemas.database_id = mz_databases.id
		WHERE mz_connections.name = '%s'
		AND mz_schemas.name = '%s'
		AND mz_databases.name = '%s';
	`, name, schema, database)
}

func readConnectionParams(id string) string {
	return fmt.Sprintf(`
		SELECT
			mz_connections.name,
			mz_schemas.name,
			mz_databases.name,
			mz_connections.type
		FROM mz_connections
		JOIN mz_schemas
			ON mz_connections.schema_id = mz_schemas.id
		JOIN mz_databases
			ON mz_schemas.database_id = mz_databases.id
		WHERE mz_connections.id = '%s';`, id)
}

func ConnectionRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*sqlx.DB)
	i := d.Id()
	q := readConnectionParams(i)

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

	qn := QualifiedName(*database, *schema, *name)
	if err := d.Set("qualified_name", qn); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

type ValueSecretStruct struct {
	Text   string
	Secret string
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
				"secret": {
					Description:   fmt.Sprintf("The %s text value.", elem),
					Type:          schema.TypeString,
					Optional:      true,
					ConflictsWith: []string{fmt.Sprintf("%s.0.text", elem)},
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
