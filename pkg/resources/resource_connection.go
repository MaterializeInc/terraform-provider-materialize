package resources

import (
	"context"
	"log"

	"github.com/MaterializeInc/terraform-provider-materialize/pkg/materialize"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/jmoiron/sqlx"
)

func connectionRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	connectionName := d.Get("name").(string)
	schemaName := d.Get("schema_name").(string)
	databaseName := d.Get("database_name").(string)

	builder := materialize.NewConnection(meta.(*sqlx.DB), connectionName, schemaName, databaseName)

	i := d.Id()
	params, err := builder.Params(i)
	if err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("name", params.ConnectionName.String); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("schema_name", params.SchemaName.String); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("database_name", params.DatabaseName.String); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("qualified_sql_name", builder.QualifiedName()); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func connectionUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	connectionName := d.Get("name").(string)
	schemaName := d.Get("schema_name").(string)
	databaseName := d.Get("database_name").(string)

	builder := materialize.NewConnection(meta.(*sqlx.DB), connectionName, schemaName, databaseName)

	if d.HasChange("name") {
		_, newName := d.GetChange("name")

		if err := builder.Rename(newName.(string)); err != nil {
			log.Printf("[ERROR] could not rename connection %s", connectionName)
			return diag.FromErr(err)
		}
	}

	return connectionRead(ctx, d, meta)
}

func connectionDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	connectionName := d.Get("name").(string)
	schemaName := d.Get("schema_name").(string)
	databaseName := d.Get("database_name").(string)

	builder := materialize.NewConnection(meta.(*sqlx.DB), connectionName, schemaName, databaseName)

	if err := builder.Drop(); err != nil {
		return diag.FromErr(err)
	}

	return nil
}
