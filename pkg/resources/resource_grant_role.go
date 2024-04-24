package resources

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/MaterializeInc/terraform-provider-materialize/pkg/materialize"
	"github.com/MaterializeInc/terraform-provider-materialize/pkg/utils"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"golang.org/x/exp/slices"
)

var grantRoleSchema = map[string]*schema.Schema{
	"role_name": {
		Description: "The role name to add member_name as a member.",
		Type:        schema.TypeString,
		Required:    true,
		ForceNew:    true,
	},
	"member_name": {
		Description: "The role name to add to role_name as a member.",
		Type:        schema.TypeString,
		Required:    true,
		ForceNew:    true,
	},
	"region": RegionSchema(),
}

func GrantRole() *schema.Resource {
	return &schema.Resource{
		Description: "Manages the system privileges for roles.",

		CreateContext: grantRoleCreate,
		ReadContext:   grantRoleRead,
		DeleteContext: grantRoleDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: grantRoleSchema,
	}
}

type RolePrivilegeKey struct {
	roleId   string
	memberId string
}

func parseRolePrivilegeKey(id string) (RolePrivilegeKey, error) {
	ie := strings.Split(id, "|")

	if len(ie) != 3 {
		return RolePrivilegeKey{}, fmt.Errorf("%s cannot be parsed correctly", id)
	}

	return RolePrivilegeKey{roleId: ie[1], memberId: ie[2]}, nil
}

func grantRoleRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	i := d.Id()

	key, err := parseRolePrivilegeKey(i)
	if err != nil {
		return diag.FromErr(err)
	}

	metaDb, region, err := utils.GetDBClientFromMeta(meta, d)
	if err != nil {
		return diag.FromErr(err)
	}

	// Scan role members
	roles, err := materialize.ScanRolePrivilege(metaDb, key.roleId, key.memberId)
	if err == sql.ErrNoRows {
		d.SetId("")
		return nil
	} else if err != nil {
		return diag.FromErr(err)
	}

	// Check if role contains member
	mapping, _ := materialize.ParseRolePrivileges(roles)

	// Check if role contains member and if not, set id to empty string so that it can be recreated
	if !slices.Contains(mapping[key.roleId], key.memberId) {
		d.SetId("")
		return nil
	}

	d.SetId(utils.TransformIdWithRegion(string(region), i))
	return nil
}

func grantRoleCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	roleName := d.Get("role_name").(string)
	memberName := d.Get("member_name").(string)

	metaDb, region, err := utils.GetDBClientFromMeta(meta, d)
	if err != nil {
		return diag.FromErr(err)
	}

	b := materialize.NewRolePrivilegeBuilder(metaDb, roleName, memberName)

	if err := b.Grant(); err != nil {
		return diag.FromErr(err)
	}

	rId, err := materialize.RoleId(metaDb, roleName)
	if err != nil {
		return diag.FromErr(err)
	}

	mId, err := materialize.RoleId(metaDb, memberName)
	if err != nil {
		return diag.FromErr(err)
	}

	key := b.GrantKey(string(region), rId, mId)
	d.SetId(key)

	return grantRoleRead(ctx, d, meta)
}

func grantRoleDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	roleName := d.Get("role_name").(string)
	memberName := d.Get("member_name").(string)

	metaDb, _, err := utils.GetDBClientFromMeta(meta, d)
	if err != nil {
		return diag.FromErr(err)
	}

	b := materialize.NewRolePrivilegeBuilder(metaDb, roleName, memberName)

	if err := b.Revoke(); err != nil {
		return diag.FromErr(err)
	}

	return nil
}
