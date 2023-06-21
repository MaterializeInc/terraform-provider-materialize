package resources

import (
	"context"
	"strings"

	"github.com/MaterializeInc/terraform-provider-materialize/pkg/materialize"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/jmoiron/sqlx"
)

func grantRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	i := d.Id()

	ie := strings.Split(i, "|")
	objType := ie[1]
	objId := ie[2]
	roleId := ie[3]

	s, err := materialize.ScanPrivileges(meta.(*sqlx.DB), objType, objId)
	if err != nil {
		return diag.FromErr(err)
	}

	priviledgeMap := materialize.ParsePrivileges(s)
	privilege := d.Get("privilege").(string)

	if !materialize.HasPrivilege(priviledgeMap[roleId], privilege) {
		return diag.Errorf("%s: object does not contain privilege: %s", i, privilege)
	}

	return nil
}
