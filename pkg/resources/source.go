package resources

import (
	"context"
	"terraform-materialize/pkg/materialize"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/jmoiron/sqlx"
)

func sourceRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*sqlx.DB)
	i := d.Id()
	q := materialize.ReadSourceParams(i)

	var name, schema, database, source_type, size, connection_name, cluster_name *string
	if err := conn.QueryRowx(q).Scan(&name, &schema, &database, &source_type, &size, &connection_name, &cluster_name); err != nil {
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

	if err := d.Set("source_type", source_type); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("size", size); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("cluster_name", cluster_name); err != nil {
		return diag.FromErr(err)
	}

	b := materialize.Source{SourceName: *name, SchemaName: *schema, DatabaseName: *database}
	if err := d.Set("qualified_sql_name", b.QualifiedName()); err != nil {
		return diag.FromErr(err)
	}

	return nil
}
