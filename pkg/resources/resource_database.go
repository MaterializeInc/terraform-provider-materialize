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

var databaseSchema = map[string]*schema.Schema{
	"name": ObjectNameSchema("database", true, true),
	"no_public_schema": {
		Description: "When set to true, the default 'public' schema will be dropped from the database.",
		Type:        schema.TypeBool,
		Optional:    true,
		Default:     false,
	},
	"comment":        CommentSchema(false),
	"ownership_role": OwnershipRoleSchema(),
	"region":         RegionSchema(),
}

func Database() *schema.Resource {
	return &schema.Resource{
		Description: "The highest level namespace hierarchy in Materialize.",

		CreateContext: databaseCreate,
		ReadContext:   databaseRead,
		UpdateContext: databaseUpdate,
		DeleteContext: databaseDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: databaseSchema,
	}
}

func databaseRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	i := d.Id()

	metaDb, region, err := utils.GetDBClientFromMeta(meta, d)
	if err != nil {
		return diag.FromErr(err)
	}

	s, err := materialize.ScanDatabase(metaDb, utils.ExtractId(i))
	if err == sql.ErrNoRows {
		d.SetId("")
		return nil
	} else if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(utils.TransformIdWithRegion(string(region), i))

	if err := d.Set("name", s.DatabaseName.String); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("ownership_role", s.OwnerName.String); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("comment", s.Comment.String); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func databaseCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	databaseName := d.Get("name").(string)

	metaDb, region, err := utils.GetDBClientFromMeta(meta, d)
	if err != nil {
		return diag.FromErr(err)
	}

	o := materialize.MaterializeObject{ObjectType: "DATABASE", Name: databaseName}
	b := materialize.NewDatabaseBuilder(metaDb, o)

	// create resource
	if err := b.Create(); err != nil {
		return diag.FromErr(err)
	}

	if d.Get("no_public_schema").(bool) {
		if err := b.DropPublicSchema(); err != nil {
			return diag.FromErr(err)
		}
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
	i, err := materialize.DatabaseId(metaDb, o)
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(utils.TransformIdWithRegion(string(region), i))

	return databaseRead(ctx, d, meta)
}

func databaseUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	databaseName := d.Get("name").(string)

	metaDb, _, err := utils.GetDBClientFromMeta(meta, d)
	if err != nil {
		return diag.FromErr(err)
	}

	o := materialize.MaterializeObject{ObjectType: "DATABASE", Name: databaseName}
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

	if d.HasChange("no_public_schema") {
		b := materialize.NewDatabaseBuilder(metaDb, o)
		if _, newNoPublicSchema := d.GetChange("no_public_schema"); newNoPublicSchema.(bool) {
			if err := b.DropPublicSchema(); err != nil {
				return diag.FromErr(err)
			}
		}
	}

	return databaseRead(ctx, d, meta)
}

func databaseDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	databaseName := d.Get("name").(string)

	metaDb, _, err := utils.GetDBClientFromMeta(meta, d)
	if err != nil {
		return diag.FromErr(err)
	}

	o := materialize.MaterializeObject{Name: databaseName}
	b := materialize.NewDatabaseBuilder(metaDb, o)

	if err := b.Drop(); err != nil {
		return diag.FromErr(err)
	}
	return nil
}
