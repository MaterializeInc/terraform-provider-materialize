package resources

import (
	"context"

	"github.com/MaterializeInc/terraform-provider-materialize/pkg/materialize"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/jmoiron/sqlx"
)

var roleSchema = map[string]*schema.Schema{
	"name":               NameSchema("role", true, false),
	"qualified_sql_name": QualifiedNameSchema("role"),
	"inherit": {
		Description: "Grants the role the ability to inheritance of privileges of other roles.",
		Type:        schema.TypeBool,
		Optional:    true,
		Default:     false,
		ForceNew:    true,
	},
	"create_role": {
		Description: "Grants the role the ability to create, alter, delete roles and the ability to grant and revoke role membership.",
		Type:        schema.TypeBool,
		Optional:    true,
		Default:     false,
	},
	"create_db": {
		Description: "Grants the role the ability to create databases.",
		Type:        schema.TypeBool,
		Optional:    true,
		Default:     false,
	},
	"create_cluster": {
		Description: "Grants the role the ability to create clusters..",
		Type:        schema.TypeBool,
		Optional:    true,
		Default:     false,
	},
}

func Role() *schema.Resource {
	return &schema.Resource{
		Description: "A new role, which is a user account in Materialize.",

		CreateContext: roleCreate,
		ReadContext:   roleRead,
		UpdateContext: roleUpdate,
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
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(i)

	if err := d.Set("name", s.RoleName.String); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("inherit", s.Inherit.Bool); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("create_role", s.CreateRole.Bool); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("create_db", s.CreateDatabase.Bool); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("create_cluster", s.CreateCluster.Bool); err != nil {
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

	if v, ok := d.GetOk("create_role"); ok && v.(bool) {
		b.CreateRole()
	}

	if v, ok := d.GetOk("create_db"); ok && v.(bool) {
		b.CreateDb()
	}

	if v, ok := d.GetOk("create_cluster"); ok && v.(bool) {
		b.CreateCluster()
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

func roleUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	roleName := d.Get("name").(string)

	b := materialize.NewRoleBuilder(meta.(*sqlx.DB), roleName)

	if d.HasChange("create_role") {
		_, nr := d.GetChange("create_role")

		if nr.(bool) {
			b.Alter("CREATEROLE")
		} else {
			b.Alter("NOCREATEROLE")
		}
	}

	if d.HasChange("create_db") {
		_, nd := d.GetChange("create_db")

		if nd.(bool) {
			b.Alter("CREATEDB")
		} else {
			b.Alter("NOCREATEDB")
		}
	}

	if d.HasChange("create_cluster") {
		_, nc := d.GetChange("create_cluster")

		if nc.(bool) {
			b.Alter("CREATECLUSTER")
		} else {
			b.Alter("NOCREATECLUSTER")
		}
	}

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
