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

var schemaSchema = map[string]*schema.Schema{
	"name":               ObjectNameSchema("schema", true, false),
	"database_name":      DatabaseNameSchema("schema", false),
	"qualified_sql_name": QualifiedNameSchema("schema"),
	"comment":            CommentSchema(false),
	"ownership_role":     OwnershipRoleSchema(),
	"region":             RegionSchema(),
}

func Schema() *schema.Resource {
	return &schema.Resource{
		Description: "The second highest level namespace hierarchy in Materialize.",

		CreateContext: schemaCreate,
		ReadContext:   schemaRead,
		UpdateContext: schemaUpdate,
		DeleteContext: schemaDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: schemaSchema,
	}
}

func schemaRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	i := d.Id()
	metaDb, region, err := utils.GetDBClientFromMeta(meta, d)
	if err != nil {
		return diag.FromErr(err)
	}
	s, err := materialize.ScanSchema(metaDb, utils.ExtractId(i))
	if err == sql.ErrNoRows {
		d.SetId("")
		return nil
	} else if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(utils.TransformIdWithRegion(string(region), i))

	if err := d.Set("name", s.SchemaName.String); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("database_name", s.DatabaseName.String); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("ownership_role", s.OwnerName.String); err != nil {
		return diag.FromErr(err)
	}

	qn := materialize.QualifiedName(s.DatabaseName.String, s.SchemaName.String)
	if err := d.Set("qualified_sql_name", qn); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("comment", s.Comment.String); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func schemaCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	schemaName := d.Get("name").(string)
	databaseName := d.Get("database_name").(string)

	metaDb, region, err := utils.GetDBClientFromMeta(meta, d)
	if err != nil {
		return diag.FromErr(err)
	}
	o := materialize.MaterializeObject{ObjectType: "SCHEMA", Name: schemaName, DatabaseName: databaseName}
	b := materialize.NewSchemaBuilder(metaDb, o)

	// create resource
	if err := b.Create(); err != nil {
		return diag.FromErr(err)
	}

	// ownership
	if v, ok := d.GetOk("ownership_role"); ok {
		ownership := materialize.NewOwnershipBuilder(metaDb, o)

		if err := ownership.Alter(v.(string)); err != nil {
			log.Printf("[DEBUG] resource failed ownership, dropping object: %s", o.Name)
			b.Drop()
			return diag.FromErr(err)
		}
	}

	// object comment
	if v, ok := d.GetOk("comment"); ok {
		comment := materialize.NewCommentBuilder(metaDb, o)

		if err := comment.Object(v.(string)); err != nil {
			log.Printf("[DEBUG] resource failed comment, dropping object: %s", o.Name)
			b.Drop()
			return diag.FromErr(err)
		}
	}

	// set id
	i, err := materialize.SchemaId(metaDb, o)
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(utils.TransformIdWithRegion(string(region), i))

	return schemaRead(ctx, d, meta)
}

func schemaUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	schemaName := d.Get("name").(string)
	databaseName := d.Get("database_name").(string)

	metaDb, _, err := utils.GetDBClientFromMeta(meta, d)
	if err != nil {
		return diag.FromErr(err)
	}
	o := materialize.MaterializeObject{ObjectType: "SCHEMA", Name: schemaName, DatabaseName: databaseName}
	b := materialize.NewOwnershipBuilder(metaDb, o)

	if d.HasChange("ownership_role") {
		_, newRole := d.GetChange("ownership_role")
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

	if d.HasChange("name") {
		oldName, newName := d.GetChange("name")
		o := materialize.MaterializeObject{ObjectType: "SCHEMA", Name: oldName.(string), DatabaseName: databaseName}
		b := materialize.NewSchemaBuilder(metaDb, o)
		if err := b.Rename(newName.(string)); err != nil {
			return diag.FromErr(err)
		}
	}

	return schemaRead(ctx, d, meta)
}

func schemaDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	schemaName := d.Get("name").(string)
	databaseName := d.Get("database_name").(string)

	metaDb, _, err := utils.GetDBClientFromMeta(meta, d)
	if err != nil {
		return diag.FromErr(err)
	}
	o := materialize.MaterializeObject{Name: schemaName, DatabaseName: databaseName}
	b := materialize.NewSchemaBuilder(metaDb, o)

	if err := b.Drop(); err != nil {
		return diag.FromErr(err)
	}
	return nil
}
