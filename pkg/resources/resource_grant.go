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

type GrantPrivilegeKey struct {
	objectType string
	objectId   string
	roleId     string
}

func parsePrivilegeKey(id string) (GrantPrivilegeKey, error) {
	ie := strings.Split(id, "|")

	if len(ie) != 5 {
		return GrantPrivilegeKey{}, fmt.Errorf("%s cannot be parsed correctly", id)
	}

	return GrantPrivilegeKey{
		objectType: ie[1],
		objectId:   ie[2],
		roleId:     ie[3],
	}, nil
}

func grantRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	i := d.Id()

	metaDb, region, err := utils.GetDBClientFromMeta(meta, d)
	if err != nil {
		return diag.FromErr(err)
	}

	key, err := parsePrivilegeKey(i)
	if err != nil {
		log.Printf("[WARN] malformed privilege (%s), removing from state file", d.Id())
		d.SetId("")
		return nil
	}

	p, err := materialize.ScanPrivileges(metaDb, key.objectType, key.objectId)
	if err == sql.ErrNoRows {
		log.Printf("[WARN] grant (%s) not found, removing from state file", d.Id())
		d.SetId("")
		return nil
	} else if err != nil {
		return diag.FromErr(err)
	}

	privilegeMap, err := materialize.MapGrantPrivileges(p)
	if err != nil {
		return diag.FromErr(err)
	}
	privilege := d.Get("privilege").(string)
	if !slices.Contains(privilegeMap[key.roleId], privilege) {
		log.Printf("[DEBUG] %s object does not contain privilege %s", i, privilege)
		// Remove id from state
		d.SetId("")
	}

	d.SetId(utils.TransformIdWithRegion(string(region), i))
	return nil
}
