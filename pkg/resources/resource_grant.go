package resources

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/MaterializeInc/terraform-provider-materialize/pkg/materialize"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/jmoiron/sqlx"
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

	key, err := parsePrivilegeKey(i)
	if err != nil {
		return diag.FromErr(err)
	}

	privileges, err := materialize.ScanPrivileges(meta.(*sqlx.DB), key.objectType, key.objectId)
	if err != nil {
		return diag.FromErr(err)
	}

	priviledgeMap := materialize.ParsePrivileges(privileges)
	privilege := d.Get("privilege").(string)

	if !materialize.HasPrivilege(priviledgeMap[key.roleId], privilege) {
		log.Printf("[DEBUG] %s: object does not contain privilege: %s", i, privilege)
		// Remove id from state
		d.SetId("")
	}

	return nil
}
