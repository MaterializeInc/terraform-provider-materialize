package resources

import (
	"context"
	"database/sql"
	"log"

	"github.com/MaterializeInc/terraform-provider-materialize/pkg/materialize"
	"github.com/MaterializeInc/terraform-provider-materialize/pkg/utils"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

var sourceTableSchema = map[string]*schema.Schema{
	"name":               ObjectNameSchema("table", true, false),
	"schema_name":        SchemaNameSchema("table", false),
	"database_name":      DatabaseNameSchema("table", false),
	"qualified_sql_name": QualifiedNameSchema("table"),
	"source": IdentifierSchema(IdentifierSchemaParams{
		Elem:        "source",
		Description: "The source this table is created from.",
		Required:    true,
		ForceNew:    true,
	}),
	"upstream_name": {
		Type:        schema.TypeString,
		Required:    true,
		ForceNew:    true,
		Description: "The name of the table in the upstream database.",
	},
	"upstream_schema_name": {
		Type:        schema.TypeString,
		Optional:    true,
		ForceNew:    true,
		Description: "The schema of the table in the upstream database.",
	},
	"text_columns": {
		Description: "Columns to be decoded as text.",
		Type:        schema.TypeList,
		Elem:        &schema.Schema{Type: schema.TypeString},
		Optional:    true,
		ForceNew:    true,
	},
	"ignore_columns": {
		Description: "Ignore specific columns when reading data from MySQL. Only compatible with MySQL sources, if the source is not MySQL, the attribute will be ignored.",
		Type:        schema.TypeList,
		Elem:        &schema.Schema{Type: schema.TypeString},
		Optional:    true,
	},
	"comment":        CommentSchema(false),
	"ownership_role": OwnershipRoleSchema(),
	"region":         RegionSchema(),
}

func SourceTable() *schema.Resource {
	return &schema.Resource{
		CreateContext: sourceTableCreate,
		ReadContext:   sourceTableRead,
		UpdateContext: sourceTableUpdate,
		DeleteContext: sourceTableDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: sourceTableSchema,
	}
}

func sourceTableCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	tableName := d.Get("name").(string)
	schemaName := d.Get("schema_name").(string)
	databaseName := d.Get("database_name").(string)

	metaDb, region, err := utils.GetDBClientFromMeta(meta, d)
	if err != nil {
		return diag.FromErr(err)
	}

	o := materialize.MaterializeObject{ObjectType: "TABLE", Name: tableName, SchemaName: schemaName, DatabaseName: databaseName}
	b := materialize.NewSourceTableBuilder(metaDb, o)

	source := materialize.GetIdentifierSchemaStruct(d.Get("source"))
	b.Source(source)

	b.UpstreamName(d.Get("upstream_name").(string))

	if v, ok := d.GetOk("upstream_schema_name"); ok {
		b.UpstreamSchemaName(v.(string))
	}

	if v, ok := d.GetOk("text_columns"); ok {
		textColumns, err := materialize.GetSliceValueString("text_columns", v.([]interface{}))
		if err != nil {
			return diag.FromErr(err)
		}
		b.TextColumns(textColumns)
	}

	if v, ok := d.GetOk("ignore_columns"); ok && len(v.([]interface{})) > 0 {
		columns, err := materialize.GetSliceValueString("ignore_columns", v.([]interface{}))
		if err != nil {
			return diag.FromErr(err)
		}
		b.IgnoreColumns(columns)
	}

	if err := b.Create(); err != nil {
		return diag.FromErr(err)
	}

	// Handle ownership
	if v, ok := d.GetOk("ownership_role"); ok {
		ownership := materialize.NewOwnershipBuilder(metaDb, o)
		if err := ownership.Alter(v.(string)); err != nil {
			log.Printf("[DEBUG] resource failed ownership, dropping object: %s", o.Name)
			b.Drop()
			return diag.FromErr(err)
		}
	}

	// Handle comments
	if v, ok := d.GetOk("comment"); ok {
		comment := materialize.NewCommentBuilder(metaDb, o)
		if err := comment.Object(v.(string)); err != nil {
			log.Printf("[DEBUG] resource failed comment, dropping object: %s", o.Name)
			b.Drop()
			return diag.FromErr(err)
		}
	}

	i, err := materialize.SourceTableId(metaDb, o)
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(utils.TransformIdWithRegion(string(region), i))

	return sourceTableRead(ctx, d, meta)
}

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

	// TODO: Set source once the source_id is available in the mz_tables table

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
