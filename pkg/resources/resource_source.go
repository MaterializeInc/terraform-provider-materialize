package resources

import (
	"context"
	"log"

	"github.com/MaterializeInc/terraform-provider-materialize/pkg/materialize"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/jmoiron/sqlx"
)

func sourceRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sourceName := d.Get("name").(string)
	schemaName := d.Get("schema_name").(string)
	databaseName := d.Get("database_name").(string)

	builder := materialize.NewSource(meta.(*sqlx.DB), sourceName, schemaName, databaseName)

	i := d.Id()
	params, _ := builder.Params(i)

	if err := d.Set("name", params.SourceName); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("schema_name", params.SchemaName); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("database_name", params.DatabaseName); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("size", params.Size); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("cluster_name", params.ClusterName); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("qualified_sql_name", builder.QualifiedName()); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func sourceUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	sourceName := d.Get("name").(string)
	schemaName := d.Get("schema_name").(string)
	databaseName := d.Get("database_name").(string)

	builder := materialize.NewSource(meta.(*sqlx.DB), sourceName, schemaName, databaseName)

	if d.HasChange("size") {
		_, newSize := d.GetChange("size")

		if err := builder.UpdateSize(newSize.(string)); err != nil {
			log.Printf("[ERROR] could not resize source %s", sourceName)
			return diag.FromErr(err)
		}
	}

	if d.HasChange("name") {
		_, newName := d.GetChange("name")

		if err := builder.Rename(newName.(string)); err != nil {
			log.Printf("[ERROR] could not rename source %s", sourceName)
			return diag.FromErr(err)
		}
	}

	return sourceRead(ctx, d, meta)
}

func sourceDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	sourceName := d.Get("name").(string)
	schemaName := d.Get("schema_name").(string)
	databaseName := d.Get("database_name").(string)

	builder := materialize.NewSource(meta.(*sqlx.DB), sourceName, schemaName, databaseName)

	if err := builder.Drop(); err != nil {
		return diag.FromErr(err)
	}

	return nil
}
