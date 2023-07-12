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
	"golang.org/x/exp/slices"
)

var grantSystemPrivilegeSchema = map[string]*schema.Schema{
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

func GrantSystemPrivilege() *schema.Resource {
	return &schema.Resource{
		Description: "Manages the system privileges for roles.",

		CreateContext: grantSystemPrivilegeCreate,
		ReadContext:   grantSystemPrivilegeRead,
		DeleteContext: grantSystemPrivilegeDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: grantSystemPrivilegeSchema,
	}
}

type SystemPrivilegeKey struct {
	roleId    string
	privilege string
}

func parseSystemPrivilegeKey(id string) (SystemPrivilegeKey, error) {
	ie := strings.Split(id, "|")

	if len(ie) != 3 {
		return SystemPrivilegeKey{}, fmt.Errorf("%s cannot be parsed correctly", id)
	}

	return SystemPrivilegeKey{roleId: ie[1], privilege: ie[2]}, nil
}

func grantSystemPrivilegeRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	i := d.Id()

	key, err := parseSystemPrivilegeKey(i)
	if err != nil {
		return diag.FromErr(err)
	}

	privileges, err := materialize.ScanSystemPrivileges(meta.(*sqlx.DB))
	if err == sql.ErrNoRows {
		d.SetId("")
		return nil
	} else if err != nil {
		return diag.FromErr(err)
	}

	// Check if system role contains privilege
	mapping, _ := materialize.ParseSystemPrivileges(privileges)

	if !slices.Contains(mapping[key.roleId], key.privilege) {
		d.SetId("")
		return diag.Errorf("system role does contain privilege %s", key.privilege)
	}

	d.SetId(i)

	return nil
}

func grantSystemPrivilegeCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	roleName := d.Get("role_name").(string)
	privilege := d.Get("privilege").(string)

	b := materialize.NewSystemPrivilegeBuilder(meta.(*sqlx.DB), roleName, privilege)

	if err := b.Grant(); err != nil {
		return diag.FromErr(err)
	}

	rId, err := materialize.RoleId(meta.(*sqlx.DB), roleName)
	if err != nil {
		return diag.FromErr(err)
	}

	key := b.GrantKey(rId, privilege)
	d.SetId(key)

	return grantSystemPrivilegeRead(ctx, d, meta)
}

func grantSystemPrivilegeDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	roleName := d.Get("role_name").(string)
	privilege := d.Get("privilege").(string)

	b := materialize.NewSystemPrivilegeBuilder(meta.(*sqlx.DB), roleName, privilege)

	if err := b.Revoke(); err != nil {
		return diag.FromErr(err)
	}

	return nil
}
