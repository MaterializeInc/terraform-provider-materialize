package resources

import (
	"context"
	"database/sql"
	"log"

	"github.com/MaterializeInc/terraform-provider-materialize/pkg/materialize"
	"github.com/MaterializeInc/terraform-provider-materialize/pkg/utils"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/jmoiron/sqlx"
)

var databaseSchema = map[string]*schema.Schema{
	"name":           ObjectNameSchema("database", true, true),
	"comment":        CommentSchema(false),
	"ownership_role": OwnershipRoleSchema(),
}

// Define the V0 schema function
func databaseSchemaV0() *schema.Resource {
	return &schema.Resource{
		Schema: databaseSchema,
	}
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

		Schema:        databaseSchema,
		SchemaVersion: 1,
		StateUpgraders: []schema.StateUpgrader{
			{
				Type:    databaseSchemaV0().CoreConfigSchema().ImpliedType(),
				Upgrade: utils.IdStateUpgradeV0,
				Version: 0,
			},
		},
	}
}

func databaseRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	i := d.Id()
	i, err := utils.ExtractId(i)
	if err != nil {
		return diag.FromErr(err)
	}

	s, err := materialize.ScanDatabase(meta.(*sqlx.DB), i)
	if err == sql.ErrNoRows {
		d.SetId("")
		return nil
	} else if err != nil {
		return diag.FromErr(err)
	}

	// set id with the region
	i, err = utils.TransformIdWithRegion(i)
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(i)

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

	o := materialize.MaterializeObject{ObjectType: "DATABASE", Name: databaseName}
	b := materialize.NewDatabaseBuilder(meta.(*sqlx.DB), o)

	// create resource
	if err := b.Create(); err != nil {
		return diag.FromErr(err)
	}

	// ownership
	if v, ok := d.GetOk("ownership_role"); ok {
		ownership := materialize.NewOwnershipBuilder(meta.(*sqlx.DB), o)

		if err := ownership.Alter(v.(string)); err != nil {
			log.Printf("[DEBUG] resource failed ownership, dropping object: %s", o.Name)
			b.Drop()
			return diag.FromErr(err)
		}
	}

	// object comment
	if v, ok := d.GetOk("comment"); ok {
		comment := materialize.NewCommentBuilder(meta.(*sqlx.DB), o)

		if err := comment.Object(v.(string)); err != nil {
			log.Printf("[DEBUG] resource failed comment, dropping object: %s", o.Name)
			b.Drop()
			return diag.FromErr(err)
		}
	}

	// set id
	i, err := materialize.DatabaseId(meta.(*sqlx.DB), o)
	if err != nil {
		return diag.FromErr(err)
	}

	// set id with the region
	i, err = utils.TransformIdWithRegion(i)
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(i)

	return databaseRead(ctx, d, meta)
}

func databaseUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	databaseName := d.Get("name").(string)

	o := materialize.MaterializeObject{ObjectType: "DATABASE", Name: databaseName}
	b := materialize.NewOwnershipBuilder(meta.(*sqlx.DB), o)

	if d.HasChange("ownership_role") {
		_, newRole := d.GetChange("ownership_role")
		if err := b.Alter(newRole.(string)); err != nil {
			return diag.FromErr(err)
		}
	}

	if d.HasChange("comment") {
		_, newComment := d.GetChange("comment")
		b := materialize.NewCommentBuilder(meta.(*sqlx.DB), o)

		if err := b.Object(newComment.(string)); err != nil {
			return diag.FromErr(err)
		}
	}

	return databaseRead(ctx, d, meta)
}

func databaseDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	databaseName := d.Get("name").(string)

	o := materialize.MaterializeObject{Name: databaseName}
	b := materialize.NewDatabaseBuilder(meta.(*sqlx.DB), o)

	if err := b.Drop(); err != nil {
		return diag.FromErr(err)
	}
	return nil
}
