package resources

import (
	"context"
	"database/sql"
	"errors"
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

	p, err := materialize.ScanPrivileges(metaDb, materialize.EntityType(key.objectType), key.objectId)
	if errors.Is(err, sql.ErrNoRows) && materialize.EntityType(key.objectType) == materialize.Cluster {
		// A swapped cluster leaves a dropped id in state, re-resolve it from the configured
		// name. cluster_name is absent when a cluster id was imported into another grant
		// resource, and then there is nothing to re-resolve from.
		clusterName, _ := d.Get("cluster_name").(string)
		if clusterName != "" {
			c, clusterErr := materialize.ScanCluster(metaDb, clusterName, true)
			if clusterErr != nil && !errors.Is(clusterErr, sql.ErrNoRows) {
				return diag.FromErr(clusterErr)
			}
			if clusterErr == nil {
				key.objectId = c.ClusterId.String
				ie := strings.Split(i, "|")
				ie[2] = key.objectId
				i = strings.Join(ie, "|")
				p, err = materialize.ScanPrivileges(metaDb, materialize.Cluster, key.objectId)
			}
		}
	}
	if errors.Is(err, sql.ErrNoRows) {
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
		return nil
	}

	d.SetId(utils.TransformIdWithRegion(string(region), i))
	return nil
}
