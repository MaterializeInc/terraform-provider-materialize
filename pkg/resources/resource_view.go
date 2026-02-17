package resources

import (
	"context"
	"database/sql"

	"github.com/MaterializeInc/terraform-provider-materialize/pkg/materialize"
	"github.com/MaterializeInc/terraform-provider-materialize/pkg/utils"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

var viewSchema = map[string]*schema.Schema{
	"name":               ObjectNameSchema("view", true, false),
	"schema_name":        SchemaNameSchema("view", false),
	"database_name":      DatabaseNameSchema("view", false),
	"qualified_sql_name": QualifiedNameSchema("view"),
	"comment":            CommentSchema(false),
	"statement": {
		Description: "The SQL statement for the view.",
		Type:        schema.TypeString,
		Required:    true,
		ForceNew:    true,
	},
	"create_sql": {
		Description: "The SQL statement used to create the view.",
		Type:        schema.TypeString,
		Computed:    true,
	},
	"ownership_role": OwnershipRoleSchema(),
	"region":         RegionSchema(),
}

func View() *schema.Resource {
	return &schema.Resource{
		Description: "Views represent queries of sources and other views that you want to save for repeated execution.",

		CreateContext: viewCreate,
		ReadContext:   viewRead,
		UpdateContext: viewUpdate,
		DeleteContext: viewDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: viewSchema,
	}
}

func viewRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	i := d.Id()

	metaDb, region, err := utils.GetDBClientFromMeta(meta, d)
	if err != nil {
		return diag.FromErr(err)
	}
	s, err := materialize.ScanView(metaDb, utils.ExtractId(i))
	if err == sql.ErrNoRows {
		d.SetId("")
		return nil
	} else if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(utils.TransformIdWithRegion(string(region), i))

	if err := d.Set("name", s.ViewName.String); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("schema_name", s.SchemaName.String); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("database_name", s.DatabaseName.String); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("ownership_role", s.OwnerName.String); err != nil {
		return diag.FromErr(err)
	}

	qn := materialize.QualifiedName(s.DatabaseName.String, s.SchemaName.String, s.ViewName.String)
	if err := d.Set("qualified_sql_name", qn); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("comment", s.Comment.String); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("create_sql", s.CreateSQL.String); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func viewCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	viewName := d.Get("name").(string)
	schemaName := d.Get("schema_name").(string)
	databaseName := d.Get("database_name").(string)

	metaDb, region, err := utils.GetDBClientFromMeta(meta, d)
	if err != nil {
		return diag.FromErr(err)
	}
	o := materialize.MaterializeObject{ObjectType: materialize.View, Name: viewName, SchemaName: schemaName, DatabaseName: databaseName}
	b := materialize.NewViewBuilder(metaDb, o)

	if v, ok := d.GetOk("statement"); ok && v.(string) != "" {
		b.SelectStmt(v.(string))
	}

	// create resource
	if err := b.Create(); err != nil {
		return diag.FromErr(err)
	}

	// ownership
	if diags := applyOwnership(d, metaDb, o, b); diags != nil {
		return diags
	}

	// object comment
	if diags := applyComment(d, metaDb, o, b); diags != nil {
		return diags
	}

	// set id
	i, err := materialize.ViewId(metaDb, o)
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(utils.TransformIdWithRegion(string(region), i))

	return viewRead(ctx, d, meta)
}

func viewUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	viewName := d.Get("name").(string)
	schemaName := d.Get("schema_name").(string)
	databaseName := d.Get("database_name").(string)

	metaDb, _, err := utils.GetDBClientFromMeta(meta, d)
	if err != nil {
		return diag.FromErr(err)
	}
	o := materialize.MaterializeObject{ObjectType: materialize.View, Name: viewName, SchemaName: schemaName, DatabaseName: databaseName}

	if d.HasChange("name") {
		oldName, newViewName := d.GetChange("name")
		o := materialize.MaterializeObject{ObjectType: materialize.View, Name: oldName.(string), SchemaName: schemaName, DatabaseName: databaseName}
		b := materialize.NewViewBuilder(metaDb, o)

		if err := b.Rename(newViewName.(string)); err != nil {
			return diag.FromErr(err)
		}
	}

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

	return viewRead(ctx, d, meta)
}

func viewDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	viewName := d.Get("name").(string)
	schemaName := d.Get("schema_name").(string)
	databaseName := d.Get("database_name").(string)

	metaDb, _, err := utils.GetDBClientFromMeta(meta, d)
	if err != nil {
		return diag.FromErr(err)
	}
	o := materialize.MaterializeObject{Name: viewName, SchemaName: schemaName, DatabaseName: databaseName}
	b := materialize.NewViewBuilder(metaDb, o)

	if err := b.Drop(); err != nil {
		return diag.FromErr(err)
	}
	return nil
}
