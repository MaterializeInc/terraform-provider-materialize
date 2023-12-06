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

var roleSchema = map[string]*schema.Schema{
	"name":               ObjectNameSchema("role", true, true),
	"qualified_sql_name": QualifiedNameSchema("role"),
	"comment":            CommentSchema(false),
	"inherit": {
		Description: "Grants the role the ability to inheritance of privileges of other roles. Unlike PostgreSQL, Materialize does not currently support `NOINHERIT`",
		Type:        schema.TypeBool,
		Computed:    true,
	},
}

// Define the V0 schema function
func roleSchemaV0() *schema.Resource {
	return &schema.Resource{
		Schema: roleSchema,
	}
}

func Role() *schema.Resource {
	return &schema.Resource{
		Description: "A role is a collection of privileges you can apply to users.",

		CreateContext: roleCreate,
		ReadContext:   roleRead,
		UpdateContext: roleUpdate,
		DeleteContext: roleDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema:        roleSchema,
		SchemaVersion: 1,
		StateUpgraders: []schema.StateUpgrader{
			{
				Type:    roleSchemaV0().CoreConfigSchema().ImpliedType(),
				Upgrade: utils.IdStateUpgradeV0,
				Version: 0,
			},
		},
	}
}

func roleRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	i := d.Id()
	s, err := materialize.ScanRole(meta.(*sqlx.DB), utils.ExtractId(i))
	if err == sql.ErrNoRows {
		d.SetId("")
		return nil
	} else if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(utils.TransformIdWithRegion(i))

	if err := d.Set("name", s.RoleName.String); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("inherit", s.Inherit.Bool); err != nil {
		return diag.FromErr(err)
	}

	qn := materialize.QualifiedName(s.RoleName.String)
	if err := d.Set("qualified_sql_name", qn); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("comment", s.Comment.String); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func roleCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	roleName := d.Get("name").(string)

	o := materialize.MaterializeObject{ObjectType: "ROLE", Name: roleName}
	b := materialize.NewRoleBuilder(meta.(*sqlx.DB), o)

	if v, ok := d.GetOk("inherit"); ok && v.(bool) {
		b.Inherit()
	}

	// create resource
	if err := b.Create(); err != nil {
		return diag.FromErr(err)
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
	i, err := materialize.RoleId(meta.(*sqlx.DB), roleName)
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(utils.TransformIdWithRegion(i))

	return roleRead(ctx, d, meta)
}

func roleUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	roleName := d.Get("name").(string)

	o := materialize.MaterializeObject{ObjectType: "ROLE", Name: roleName}

	if d.HasChange("comment") {
		_, newComment := d.GetChange("comment")
		b := materialize.NewCommentBuilder(meta.(*sqlx.DB), o)

		if err := b.Object(newComment.(string)); err != nil {
			return diag.FromErr(err)
		}
	}

	return roleRead(ctx, d, meta)
}

func roleDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	roleName := d.Get("name").(string)

	o := materialize.MaterializeObject{ObjectType: "ROLE", Name: roleName}
	b := materialize.NewRoleBuilder(meta.(*sqlx.DB), o)

	if err := b.Drop(); err != nil {
		return diag.FromErr(err)
	}
	return nil
}
