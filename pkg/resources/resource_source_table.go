package resources

import (
	"context"
	"database/sql"

	"github.com/MaterializeInc/terraform-provider-materialize/pkg/materialize"
	"github.com/MaterializeInc/terraform-provider-materialize/pkg/utils"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func sourceTableRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	i := d.Id()

	metaDb, region, err := utils.GetDBClientFromMeta(meta, d)
	if err != nil {
		return diag.FromErr(err)
	}

	t, err := materialize.ScanSourceTable(metaDb, utils.ExtractId(i))
	if err == sql.ErrNoRows {
		d.SetId("")
		return nil
	} else if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(utils.TransformIdWithRegion(string(region), i))

	if err := d.Set("name", t.TableName.String); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("schema_name", t.SchemaName.String); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("database_name", t.DatabaseName.String); err != nil {
		return diag.FromErr(err)
	}

	source := []interface{}{
		map[string]interface{}{
			"name":          t.SourceName.String,
			"schema_name":   t.SourceSchemaName.String,
			"database_name": t.SourceDatabaseName.String,
		},
	}
	if err := d.Set("source", source); err != nil {
		return diag.FromErr(err)
	}

	// TODO: Set the upstream_name and upstream_schema_name once supported on the Materialize side
	// if err := d.Set("upstream_name", t.UpstreamName.String); err != nil {
	// 	return diag.FromErr(err)
	// }

	// if err := d.Set("upstream_schema_name", t.UpstreamSchemaName.String); err != nil {
	// 	return diag.FromErr(err)
	// }

	if err := d.Set("ownership_role", t.OwnerName.String); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("comment", t.Comment.String); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func sourceTableUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	tableName := d.Get("name").(string)
	schemaName := d.Get("schema_name").(string)
	databaseName := d.Get("database_name").(string)

	metaDb, _, err := utils.GetDBClientFromMeta(meta, d)
	if err != nil {
		return diag.FromErr(err)
	}

	o := materialize.MaterializeObject{ObjectType: "TABLE", Name: tableName, SchemaName: schemaName, DatabaseName: databaseName}

	if d.HasChange("name") {
		oldName, newName := d.GetChange("name")
		o := materialize.MaterializeObject{ObjectType: "TABLE", Name: oldName.(string), SchemaName: schemaName, DatabaseName: databaseName}
		b := materialize.NewSourceTableBuilder(metaDb, o)
		if err := b.Rename(newName.(string)); err != nil {
			return diag.FromErr(err)
		}
	}

	// TODO: Handle source and text_columns changes once supported on the Materialize side

	if d.HasChange("ownership_role") {
		_, newRole := d.GetChange("ownership_role")
		b := materialize.NewOwnershipBuilder(metaDb, o)

		if err := b.Alter(newRole.(string)); err != nil {
			return diag.FromErr(err)
		}
	}

	if d.HasChange("comment") {
		_, newComment := d.GetChange("comment")
		b := materialize.NewCommentBuilder(metaDb, o)

		if err := b.Object(newComment.(string)); err != nil {
			return diag.FromErr(err)
		}
	}

	return sourceTableRead(ctx, d, meta)
}

func sourceTableDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	tableName := d.Get("name").(string)
	schemaName := d.Get("schema_name").(string)
	databaseName := d.Get("database_name").(string)

	metaDb, _, err := utils.GetDBClientFromMeta(meta, d)
	if err != nil {
		return diag.FromErr(err)
	}

	o := materialize.MaterializeObject{ObjectType: "TABLE", Name: tableName, SchemaName: schemaName, DatabaseName: databaseName}
	b := materialize.NewSourceTableBuilder(metaDb, o)

	if err := b.Drop(); err != nil {
		return diag.FromErr(err)
	}

	return nil
}
