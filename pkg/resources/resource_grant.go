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

type GrantPrivilege struct {
	objectType string
	objectId   string
	roleId     string
}

func parsePrivilegeId(id string) (GrantPrivilege, error) {
	ie := strings.Split(id, "|")

	if len(ie) != 5 {
		return GrantPrivilege{}, fmt.Errorf("%s cannot be parsed correctly", id)
	}

	return GrantPrivilege{
		objectType: ie[1],
		objectId:   ie[2],
		roleId:     ie[3],
	}, nil
}

func grantRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	i := d.Id()

	dp, err := parsePrivilegeId(i)
	if err != nil {
		return diag.FromErr(err)
	}

	s, err := materialize.ScanPrivileges(meta.(*sqlx.DB), dp.objectType, dp.objectId)
	if err != nil {
		return diag.FromErr(err)
	}

	priviledgeMap := materialize.ParsePrivileges(s)
	privilege := d.Get("privilege").(string)

	if !materialize.HasPrivilege(priviledgeMap[dp.roleId], privilege) {
		return diag.Errorf("%s: object does not contain privilege: %s", i, privilege)
	}

	return nil
}
