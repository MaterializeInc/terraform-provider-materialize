package resources

import (
	"context"

	"github.com/MaterializeInc/terraform-provider-materialize/pkg/materialize"
	"github.com/MaterializeInc/terraform-provider-materialize/pkg/utils"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

var grantDatabaseDefaultPrivilegeSchema = map[string]*schema.Schema{
	"grantee_name":     GranteeNameSchema(),
	"target_role_name": TargetRoleNameSchema(),
	"privilege":        PrivilegeSchema("DATABASE"),
	"region":           RegionSchema(),
}

func GrantDatabaseDefaultPrivilege() *schema.Resource {
	return &schema.Resource{
		Description: DefaultPrivilegeDefinition,

		CreateContext: grantDatabaseDefaultPrivilegeCreate,
		ReadContext:   grantDefaultPrivilegeRead,
		DeleteContext: grantDatabaseDefaultPrivilegeDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: grantDatabaseDefaultPrivilegeSchema,
	}
}

func grantDatabaseDefaultPrivilegeCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	granteeName := d.Get("grantee_name").(string)
	targetName := d.Get("target_role_name").(string)
	privilege := d.Get("privilege").(string)

	metaDb, region, err := utils.GetDBClientFromMeta(meta, d)
	if err != nil {
		return diag.FromErr(err)
	}

	b := materialize.NewDefaultPrivilegeBuilder(metaDb, "DATABASE", granteeName, targetName, privilege)

	// create resource
	if err := b.Grant(); err != nil {
		return diag.FromErr(err)
	}

	// Query ids
	gId, err := materialize.RoleId(metaDb, granteeName)
	if err != nil {
		return diag.FromErr(err)
	}

	tId, err := materialize.RoleId(metaDb, targetName)
	if err != nil {
		return diag.FromErr(err)
	}

	key := b.GrantKey(string(region), "DATABASE", gId, tId, "", "", privilege)
	d.SetId(key)

	return grantDefaultPrivilegeRead(ctx, d, meta)
}

func grantDatabaseDefaultPrivilegeDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	granteenName := d.Get("grantee_name").(string)
	targetName := d.Get("target_role_name").(string)
	privilege := d.Get("privilege").(string)

	metaDb, _, err := utils.GetDBClientFromMeta(meta, d)
	if err != nil {
		return diag.FromErr(err)
	}

	b := materialize.NewDefaultPrivilegeBuilder(metaDb, "DATABASE", granteenName, targetName, privilege)

	if err := b.Revoke(); err != nil {
		return diag.FromErr(err)
	}
	return nil
}
