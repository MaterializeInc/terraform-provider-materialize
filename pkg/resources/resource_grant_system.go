package resources

import (
	"context"
	"fmt"
	"strings"

	"github.com/MaterializeInc/terraform-provider-materialize/pkg/materialize"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/jmoiron/sqlx"
)

var grantSystemSchema = map[string]*schema.Schema{
	"role_name": {
		Description: "The name of the role to grant privilege to.",
		Type:        schema.TypeString,
		Required:    true,
		ForceNew:    true,
	},
	"privilege": {
		Description:  "The system privilege to grant.",
		Type:         schema.TypeString,
		Required:     true,
		ForceNew:     true,
		ValidateFunc: validPrivileges("SYSTEM"),
	},
}

func GrantSystem() *schema.Resource {
	return &schema.Resource{
		Description: "Manages the system privileges for roles.",

		CreateContext: grantSystemCreate,
		ReadContext:   grantSystemRead,
		DeleteContext: grantSystemDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: grantSystemSchema,
	}
}

type SystemPrivilege struct {
	roleId string
}

func parseSystemPrivilegeId(id string) (SystemPrivilege, error) {
	ie := strings.Split(id, "|")

	if len(ie) != 3 {
		return SystemPrivilege{}, fmt.Errorf("%s cannot be parsed correctly", id)
	}

	return SystemPrivilege{roleId: ie[1]}, nil
}

func grantSystemRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	i := d.Id()

	sp, err := parseSystemPrivilegeId(i)
	if err != nil {
		return diag.FromErr(err)
	}

	privileges, err := materialize.ScanSystemPrivileges(meta.(*sqlx.DB))
	if err != nil {
		return diag.FromErr(err)
	}

	privilegeMap := materialize.ParseSystemPrivileges(privileges)
	privilege := d.Get("privilege").(string)

	if !materialize.HasPrivilege(privilegeMap[sp.roleId], privilege) {
		return diag.Errorf("%s: default privilege privilege: %s not set", i, privilege)
	}

	return nil
}

func grantSystemCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	roleName := d.Get("role_name").(string)
	privilege := d.Get("privilege").(string)

	b := materialize.NewSystemPrivilegeBuilder(meta.(*sqlx.DB), roleName, privilege)

	if err := b.Grant(); err != nil {
		return diag.FromErr(err)
	}

	i, err := materialize.SystemPrivilegeId(meta.(*sqlx.DB), roleName, privilege)
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(i)

	return grantSystemRead(ctx, d, meta)
}

func grantSystemDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	roleName := d.Get("role_name").(string)
	privilege := d.Get("privilege").(string)

	b := materialize.NewSystemPrivilegeBuilder(meta.(*sqlx.DB), roleName, privilege)

	if err := b.Revoke(); err != nil {
		return diag.FromErr(err)
	}

	return nil
}
