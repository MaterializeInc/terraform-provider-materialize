package resources

import (
	"context"

	"github.com/MaterializeInc/terraform-provider-materialize/pkg/materialize"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/jmoiron/sqlx"
)

var grantClusterDefaultPrivilegeSchema = map[string]*schema.Schema{
	"grantee_name": {
		Description: "The role name that will gain the default privilege. Use the `PUBLIC` pseudo-role to grant privileges to all roles.",
		Type:        schema.TypeString,
		Required:    true,
		ForceNew:    true,
	},
	"target_role_name": {
		Description: "The default privilege will apply to objects created by this role. If this is left blank, then the current role is assumed. Use the `PUBLIC` pseudo-role to target objects created by all roles. If using `ALL` will apply to objects created by all roles",
		Type:        schema.TypeString,
		Optional:    true,
		ForceNew:    true,
		Computed:    true,
	},
	"privilege": {
		Description:  "The privilege to grant to the object.",
		Type:         schema.TypeString,
		Required:     true,
		ForceNew:     true,
		ValidateFunc: validPrivileges("CLUSTER"),
	},
}

func GrantClusterDefaultPrivilege() *schema.Resource {
	return &schema.Resource{
		Description: "Defines default privileges that will be applied to objects created in the future. It does not affect any existing objects.",

		CreateContext: grantClusterDefaultPrivilegeCreate,
		ReadContext:   grantDefaultPrivilegeRead,
		DeleteContext: grantClusterDefaultPrivilegeDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: grantClusterDefaultPrivilegeSchema,
	}
}

func grantClusterDefaultPrivilegeCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	granteenName := d.Get("grantee_name").(string)
	privilege := d.Get("privilege").(string)

	b := materialize.NewDefaultPrivilegeBuilder(meta.(*sqlx.DB), "CLUSTER", granteenName, privilege)

	var targetRole string
	if v, ok := d.GetOk("target_role_name"); ok && v.(string) != "" {
		targetRole = v.(string)
		b.TargetRole(targetRole)
	}

	// create resource
	if err := b.Grant(); err != nil {
		return diag.FromErr(err)
	}

	// set id
	i, err := materialize.DefaultPrivilegeId(meta.(*sqlx.DB), "CLUSTER", granteenName, targetRole, "", "", privilege)
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(i)

	return grantDefaultPrivilegeRead(ctx, d, meta)
}

func grantClusterDefaultPrivilegeDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	granteenName := d.Get("grantee_name").(string)
	privilege := d.Get("privilege").(string)

	b := materialize.NewDefaultPrivilegeBuilder(meta.(*sqlx.DB), "CLUSTER", granteenName, privilege)

	if v, ok := d.GetOk("target_role_name"); ok && v.(string) != "" {
		b.TargetRole(v.(string))
	}

	if err := b.Revoke(); err != nil {
		return diag.FromErr(err)
	}
	return nil
}
