package resources

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/MaterializeInc/terraform-provider-materialize/pkg/materialize"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/jmoiron/sqlx"
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

type RolePrivilege struct {
	roleId   string
	memberId string
}

func parseRolePrivilegeId(id string) (RolePrivilege, error) {
	ie := strings.Split(id, "|")

	if len(ie) != 3 {
		return RolePrivilege{}, fmt.Errorf("%s cannot be parsed correctly", id)
	}

	return RolePrivilege{roleId: ie[1], memberId: ie[2]}, nil
}

func grantRoleRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	i := d.Id()

	sp, err := parseRolePrivilegeId(i)
	if err != nil {
		return diag.FromErr(err)
	}

	_, err = materialize.ScanRolePrivilege(meta.(*sqlx.DB), sp.roleId, sp.memberId)
	if err == sql.ErrNoRows {
		d.SetId("")
		return nil
	} else if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(i)

	return nil
}

func grantRoleCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	roleName := d.Get("role_name").(string)
	memberName := d.Get("member_name").(string)

	b := materialize.NewRolePrivilegeBuilder(meta.(*sqlx.DB), roleName, memberName)

	if err := b.Grant(); err != nil {
		return diag.FromErr(err)
	}

	i, err := materialize.RolePrivilegeId(meta.(*sqlx.DB), roleName, memberName)
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(i)

	return grantRoleRead(ctx, d, meta)
}

func grantRoleDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	roleName := d.Get("role_name").(string)
	memberName := d.Get("member_name").(string)

	b := materialize.NewRolePrivilegeBuilder(meta.(*sqlx.DB), roleName, memberName)

	if err := b.Revoke(); err != nil {
		return diag.FromErr(err)
	}

	return nil
}
