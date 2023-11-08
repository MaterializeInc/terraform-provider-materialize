package resources

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"strings"

	"github.com/MaterializeInc/terraform-provider-materialize/pkg/materialize"
	"github.com/MaterializeInc/terraform-provider-materialize/pkg/utils"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"golang.org/x/exp/slices"
)

var grantSystemPrivilegeSchema = map[string]*schema.Schema{
	"role_name": RoleNameSchema(),
	"privilege": {
		Description:  "The system privilege to grant.",
		Type:         schema.TypeString,
		Required:     true,
		ForceNew:     true,
		ValidateFunc: validPrivileges("SYSTEM"),
	},
	"region": RegionSchema(),
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

	metaDb, err := utils.GetDBClientFromMeta(meta, d)
	if err != nil {
		return diag.FromErr(err)
	}

	key, err := parseSystemPrivilegeKey(i)
	if err != nil {
		log.Printf("[WARN] malformed privilege (%s), removing from state file", d.Id())
		d.SetId("")
		return nil
	}

	p, err := materialize.ScanSystemPrivileges(metaDb)
	if err == sql.ErrNoRows {
		log.Printf("[WARN] grant (%s) not found, removing from state file", d.Id())
		d.SetId("")
		return nil
	} else if err != nil {
		return diag.FromErr(err)
	}

	// Check if system role contains privilege
	var privileges = []string{}
	for _, pr := range p {
		privileges = append(privileges, pr.Privileges)
	}
	privilegeMap, err := materialize.MapGrantPrivileges(privileges)
	if err != nil {
		return diag.FromErr(err)
	}

	if !slices.Contains(privilegeMap[key.roleId], key.privilege) {
		log.Printf("[DEBUG] %s object does not contain privilege %s", i, key.privilege)
		// Remove id from state
		d.SetId("")
	}

	d.SetId(utils.TransformIdWithRegion(i))
	return nil
}

func grantSystemPrivilegeCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	roleName := d.Get("role_name").(string)
	privilege := d.Get("privilege").(string)

	metaDb, err := utils.GetDBClientFromMeta(meta, d)
	if err != nil {
		return diag.FromErr(err)
	}

	b := materialize.NewSystemPrivilegeBuilder(metaDb, roleName, privilege)

	if err := b.Grant(); err != nil {
		return diag.FromErr(err)
	}

	rId, err := materialize.RoleId(metaDb, roleName)
	if err != nil {
		return diag.FromErr(err)
	}

	key := b.GrantKey(utils.Region, rId, privilege)
	d.SetId(key)

	return grantSystemPrivilegeRead(ctx, d, meta)
}

func grantSystemPrivilegeDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	roleName := d.Get("role_name").(string)
	privilege := d.Get("privilege").(string)

	metaDb, err := utils.GetDBClientFromMeta(meta, d)
	if err != nil {
		return diag.FromErr(err)
	}

	b := materialize.NewSystemPrivilegeBuilder(metaDb, roleName, privilege)

	if err := b.Revoke(); err != nil {
		return diag.FromErr(err)
	}

	return nil
}
