package resources

import (
	"context"
	"database/sql"

	"github.com/MaterializeInc/terraform-provider-materialize/pkg/materialize"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/jmoiron/sqlx"
)

var roleSchema = map[string]*schema.Schema{
	"name":               NameSchema("role", true, true),
	"qualified_sql_name": QualifiedNameSchema("role"),
	"inherit": {
		Description: "Grants the role the ability to inheritance of privileges of other roles. Unlike PostgreSQL, Materialize does not currently support `NOINHERIT`",
		Type:        schema.TypeBool,
		Computed:    true,
	},
}

func Role() *schema.Resource {
	return &schema.Resource{
		Description: "A new role, which is a user account in Materialize.",

		CreateContext: roleCreate,
		ReadContext:   roleRead,
		DeleteContext: roleDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: roleSchema,
	}
}

func roleRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	i := d.Id()
	s, err := materialize.ScanRole(meta.(*sqlx.DB), i)
	if err == sql.ErrNoRows {
		d.SetId("")
		return nil
	} else if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(i)

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

	return nil
}

func roleCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	roleName := d.Get("name").(string)

	b := materialize.NewRoleBuilder(meta.(*sqlx.DB), roleName)

	if v, ok := d.GetOk("inherit"); ok && v.(bool) {
		b.Inherit()
	}

	// create resource
	if err := b.Create(); err != nil {
		return diag.FromErr(err)
	}

	// set id
	i, err := materialize.RoleId(meta.(*sqlx.DB), roleName)
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(i)

	return roleRead(ctx, d, meta)
}

func roleDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	roleName := d.Get("name").(string)

	b := materialize.NewRoleBuilder(meta.(*sqlx.DB), roleName)

	if err := b.Drop(); err != nil {
		return diag.FromErr(err)
	}
	return nil
}
