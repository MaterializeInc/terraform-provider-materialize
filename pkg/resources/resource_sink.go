package resources

import (
	"context"
	"log"

	"github.com/MaterializeInc/terraform-provider-materialize/pkg/materialize"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/jmoiron/sqlx"
)

func sinkRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sinkName := d.Get("name").(string)
	schemaName := d.Get("schema_name").(string)
	databaseName := d.Get("database_name").(string)

	builder := materialize.NewSink(meta.(*sqlx.DB), sinkName, schemaName, databaseName)

	i := d.Id()
	params, err := builder.Params(i)
	if err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("name", params.SinkName.String); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("schema_name", params.SchemaName.String); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("database_name", params.DatabaseName.String); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("size", params.Size.String); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("cluster_name", params.ClusterName.String); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("qualified_sql_name", builder.QualifiedName()); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func sinkUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	sinkName := d.Get("name").(string)
	schemaName := d.Get("schema_name").(string)
	databaseName := d.Get("database_name").(string)

	builder := materialize.NewSink(meta.(*sqlx.DB), sinkName, schemaName, databaseName)

	if d.HasChange("size") {
		_, newSize := d.GetChange("size")

		if err := builder.UpdateSize(newSize.(string)); err != nil {
			log.Printf("[ERROR] could not resize sink %s", sinkName)
			return diag.FromErr(err)
		}
	}

	if d.HasChange("name") {
		_, newName := d.GetChange("name")

		if err := builder.Rename(newName.(string)); err != nil {
			log.Printf("[ERROR] could not rename sink %s", sinkName)
			return diag.FromErr(err)
		}
	}

	return sinkRead(ctx, d, meta)
}

func sinkDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	sinkName := d.Get("name").(string)
	schemaName := d.Get("schema_name").(string)
	databaseName := d.Get("database_name").(string)

	builder := materialize.NewSink(meta.(*sqlx.DB), sinkName, schemaName, databaseName)

	if err := builder.Drop(); err != nil {
		return diag.FromErr(err)
	}

	return nil
}
