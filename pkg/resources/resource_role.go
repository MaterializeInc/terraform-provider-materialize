package resources

import (
	"context"
	"database/sql"
	"log"

	"github.com/MaterializeInc/terraform-provider-materialize/pkg/materialize"
	"github.com/MaterializeInc/terraform-provider-materialize/pkg/utils"

	"github.com/hashicorp/go-cty/cty"
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
	"password": {
		Description: "Password for the role. Only available in self-hosted Materialize environments with password authentication enabled. Required for password-based authentication. Use password_wo for write-only ephemeral values that won't be stored in state.",
		Type:        schema.TypeString,
		Optional:    true,
		Sensitive:   true,
	},
	"password_wo": {
		Description:  "Write-only password for the role that supports ephemeral values and won't be stored in Terraform state or plan. Only available in self-hosted Materialize environments with password authentication enabled. Required for password-based authentication. Requires Terraform 1.11+. Must be used with password_wo_version.",
		Type:         schema.TypeString,
		Optional:     true,
		Sensitive:    true,
		WriteOnly:    true,
		RequiredWith: []string{"password_wo_version"},
	},
	"password_wo_version": {
		Description:  "Version number for the write-only password. Increment this to trigger an update of the password value when using password_wo. Must be used with password_wo.",
		Type:         schema.TypeInt,
		Optional:     true,
		RequiredWith: []string{"password_wo"},
	},
	"superuser": {
		Description: "Whether the role is a superuser. Only available in self-hosted Materialize environments with password authentication enabled.",
		Type:        schema.TypeBool,
		Optional:    true,
		Computed:    true,
	},
	"login": {
		Description: "Whether the role can log in. Only available in self-hosted Materialize environments with password authentication enabled.",
		Type:        schema.TypeBool,
		Optional:    true,
		Computed:    true,
	},
	"create_if_not_exists": {
		Description: "If `true`, adopt a pre-existing role with the same name instead of failing when it already exists. Materialize has no `CREATE ROLE IF NOT EXISTS`, so this is useful when roles may be auto-provisioned by an external system (for example, an SSO/OIDC user whose role is created on first login). When the role already exists, Terraform takes over managing it and applies the configured attributes. Note that Terraform will then drop the role on destroy.",
		Type:        schema.TypeBool,
		Optional:    true,
		Default:     false,
	},
	"region": RegionSchema(),
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

		Schema: roleSchema,
	}
}

func roleRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	i := d.Id()

	metaDb, region, err := utils.GetDBClientFromMeta(meta, d)
	if err != nil {
		return diag.FromErr(err)
	}

	s, err := materialize.ScanRole(metaDb, utils.ExtractId(i))
	if err == sql.ErrNoRows {
		d.SetId("")
		return nil
	} else if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(utils.TransformIdWithRegion(string(region), i))

	if err := d.Set("name", s.RoleName.String); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("inherit", s.Inherit.Bool); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("superuser", s.Superuser.Bool); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("login", s.Login.Bool); err != nil {
		return diag.FromErr(err)
	}

	qn := materialize.QualifiedName(s.RoleName.String)
	if err := d.Set("qualified_sql_name", qn); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("comment", s.Comment.String); err != nil {
		return diag.FromErr(err)
	}

	// create_if_not_exists is a create-time behavior with no representation in
	// the catalog. Preserve the configured value (defaulting to false on
	// import, where there is no prior state) so plans and imports round-trip
	// cleanly instead of showing spurious drift.
	if err := d.Set("create_if_not_exists", d.Get("create_if_not_exists")); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func roleCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	roleName := d.Get("name").(string)

	metaDb, region, err := utils.GetDBClientFromMeta(meta, d)
	if err != nil {
		return diag.FromErr(err)
	}

	o := materialize.MaterializeObject{ObjectType: materialize.Role, Name: roleName}
	b := materialize.NewRoleBuilder(metaDb, o)

	// When create_if_not_exists is set, adopt a role that already exists (for
	// example, one auto-provisioned by SSO/OIDC on first login) instead of
	// failing with "role already exists". Materialize has no
	// CREATE ROLE IF NOT EXISTS, so we probe the catalog and, if the role is
	// present, reconcile it into state by applying the configured attributes.
	if d.Get("create_if_not_exists").(bool) {
		existingID, err := materialize.RoleId(metaDb, roleName)
		if err != nil && err != sql.ErrNoRows {
			return diag.FromErr(err)
		}
		if err == nil && existingID != "" {
			log.Printf("[DEBUG] role %q already exists (id %s), adopting into state", roleName, existingID)
			if diags := roleAdopt(d, metaDb, b, o); diags != nil {
				return diags
			}
			d.SetId(utils.TransformIdWithRegion(string(region), existingID))
			return roleRead(ctx, d, meta)
		}
	}

	if v, ok := d.GetOk("inherit"); ok && v.(bool) {
		b.Inherit()
	}

	if v, ok := d.GetOk("password"); ok && v.(string) != "" {
		b.Password(v.(string))
	} else if valueWo, ok := d.GetOk("password_wo"); ok && valueWo.(string) != "" {
		b.Password(valueWo.(string))
	}

	if v, ok := d.GetOk("superuser"); ok {
		b.Superuser(v.(bool))
	}

	if v, ok := d.GetOk("login"); ok {
		b.Login(v.(bool))
	}

	// create resource
	if err := b.Create(); err != nil {
		return diag.FromErr(err)
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
	i, err := materialize.RoleId(metaDb, roleName)
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(utils.TransformIdWithRegion(string(region), i))

	return roleRead(ctx, d, meta)
}

// roleAdopt reconciles a pre-existing role into Terraform state by applying the
// configured attributes via ALTER ROLE. It is used by roleCreate when
// create_if_not_exists is set and the role already exists in the catalog. Note
// that inherit cannot be altered (Materialize roles are always INHERIT), so it
// is intentionally not reconciled here.
func roleAdopt(d *schema.ResourceData, metaDb *sqlx.DB, b *materialize.RoleBuilder, o materialize.MaterializeObject) diag.Diagnostics {
	var password string
	if v, ok := d.GetOk("password"); ok && v.(string) != "" {
		password = v.(string)
	} else if valueWo, ok := d.GetOk("password_wo"); ok && valueWo.(string) != "" {
		password = valueWo.(string)
	}
	if password != "" {
		if err := b.AlterPassword(password); err != nil {
			return diag.FromErr(err)
		}
	}

	if v, ok := d.GetOk("login"); ok {
		if err := b.AlterLogin(v.(bool)); err != nil {
			return diag.FromErr(err)
		}
	}

	if v, ok := d.GetOk("superuser"); ok {
		if err := b.AlterSuperuser(v.(bool)); err != nil {
			return diag.FromErr(err)
		}
	}

	if v, ok := d.GetOk("comment"); ok {
		comment := materialize.NewCommentBuilder(metaDb, o)
		if err := comment.Object(v.(string)); err != nil {
			return diag.FromErr(err)
		}
	}

	return nil
}

func roleUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	roleName := d.Get("name").(string)

	o := materialize.MaterializeObject{ObjectType: materialize.Role, Name: roleName}

	metaDb, _, err := utils.GetDBClientFromMeta(meta, d)
	if err != nil {
		return diag.FromErr(err)
	}

	b := materialize.NewRoleBuilder(metaDb, o)

	if d.HasChange("password") {
		_, newPassword := d.GetChange("password")
		if newPassword.(string) != "" {
			if err := b.AlterPassword(newPassword.(string)); err != nil {
				return diag.FromErr(err)
			}
		}
	}

	if d.HasChange("superuser") {
		_, newSuperuser := d.GetChange("superuser")
		if err := b.AlterSuperuser(newSuperuser.(bool)); err != nil {
			return diag.FromErr(err)
		}
	}

	if d.HasChange("login") {
		_, newLogin := d.GetChange("login")
		if err := b.AlterLogin(newLogin.(bool)); err != nil {
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

	if d.HasChange("password_wo_version") {
		if passwordWo, _ := d.GetRawConfigAt(cty.GetAttrPath("password_wo")); !passwordWo.IsNull() {
			if !passwordWo.Type().Equals(cty.String) {
				return diag.Errorf("error retrieving write-only argument: password_wo - retrieved config value is not a string")
			}
			if err := b.AlterPassword(passwordWo.AsString()); err != nil {
				return diag.FromErr(err)
			}
		}
	}

	return roleRead(ctx, d, meta)
}

func roleDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	roleName := d.Get("name").(string)

	metaDb, _, err := utils.GetDBClientFromMeta(meta, d)
	if err != nil {
		return diag.FromErr(err)
	}

	o := materialize.MaterializeObject{ObjectType: materialize.Role, Name: roleName}
	b := materialize.NewRoleBuilder(metaDb, o)

	if err := b.Drop(); err != nil {
		return diag.FromErr(err)
	}
	return nil
}
