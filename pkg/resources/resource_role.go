package resources

import (
	"context"
	"database/sql"
	"log"

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

type RoleParams struct {
	RoleName       sql.NullString `db:"name"`
	Inherit        sql.NullBool   `db:"inherit"`
	CreateRole     sql.NullBool   `db:"create_role"`
	CreateDatabase sql.NullBool   `db:"create_db"`
	CreateCluster  sql.NullBool   `db:"create_cluster"`
}

func roleRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*sqlx.DB)
	i := d.Id()
	q := materialize.ReadRoleParams(i)

	var s RoleParams
	if err := conn.Get(&s, q); err != nil {
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
	conn := meta.(*sqlx.DB)
	roleName := d.Get("name").(string)

	builder := materialize.NewRoleBuilder(roleName)

	if v, ok := d.GetOk("inherit"); ok && v.(bool) {
		builder.Inherit()
	}

	if v, ok := d.GetOk("create_role"); ok && v.(bool) {
		builder.CreateRole()
	}

	if v, ok := d.GetOk("create_db"); ok && v.(bool) {
		builder.CreateDb()
	}

	if v, ok := d.GetOk("create_cluster"); ok && v.(bool) {
		builder.CreateCluster()
	}

	qc := builder.Create()
	qr := builder.ReadId()

	if err := createResource(conn, d, qc, qr, "role"); err != nil {
		return diag.FromErr(err)
	}
	return roleRead(ctx, d, meta)
}

func roleUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*sqlx.DB)
	roleName := d.Get("name").(string)

	if d.HasChange("create_role") {
		_, nr := d.GetChange("create_role")

		var qr string
		if nr.(bool) {
			qr = materialize.NewRoleBuilder(roleName).Alter("CREATEROLE")
		} else {
			qr = materialize.NewRoleBuilder(roleName).Alter("NOCREATEROLE")
		}

		if err := execResource(conn, qr); err != nil {
			log.Printf("[ERROR] could not update 'create role' permission for role: %s", qr)
			return diag.FromErr(err)
		}
	}

	if d.HasChange("create_db") {
		_, nd := d.GetChange("create_db")

		var qd string
		if nd.(bool) {
			qd = materialize.NewRoleBuilder(roleName).Alter("CREATEDB")
		} else {
			qd = materialize.NewRoleBuilder(roleName).Alter("NOCREATEDB")
		}

		if err := execResource(conn, qd); err != nil {
			log.Printf("[ERROR] could not update 'create db' permission for role: %s", qd)
			return diag.FromErr(err)
		}
	}

	if d.HasChange("create_cluster") {
		_, nc := d.GetChange("create_cluster")

		var qc string
		if nc.(bool) {
			qc = materialize.NewRoleBuilder(roleName).Alter("CREATECLUSTER")
		} else {
			qc = materialize.NewRoleBuilder(roleName).Alter("NOCREATECLUSTER")
		}

		if err := execResource(conn, qc); err != nil {
			log.Printf("[ERROR] could not update 'create cluster' permission for role: %s", qc)
			return diag.FromErr(err)
		}
	}

	return roleRead(ctx, d, meta)
}

func roleDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*sqlx.DB)
	roleName := d.Get("name").(string)

	q := materialize.NewRoleBuilder(roleName).Drop()

	if err := dropResource(conn, d, q, "role"); err != nil {
		return diag.FromErr(err)
	}
	return nil
}
