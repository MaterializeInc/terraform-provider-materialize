package resources

import (
	"context"

	"github.com/MaterializeInc/terraform-provider-materialize/pkg/materialize"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/jmoiron/sqlx"
)

func connectionRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*sqlx.DB)
	i := d.Id()
	q := materialize.ReadConnectionParams(i)

	var name, schema, database *string
	if err := conn.QueryRowx(q).Scan(&name, &schema, &database); err != nil {
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

	b := materialize.Connection{ConnectionName: *name, SchemaName: *schema, DatabaseName: *database}
	if err := d.Set("qualified_sql_name", b.QualifiedName()); err != nil {
		return diag.FromErr(err)
	}

	return nil
}
