package resources

import (
	"context"
	"fmt"

	"github.com/MaterializeInc/terraform-provider-materialize/pkg/materialize"
	"github.com/MaterializeInc/terraform-provider-materialize/pkg/utils"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

var grantClusterSchema = map[string]*schema.Schema{
	"role_name": RoleNameSchema(),
	"privilege": PrivilegeSchema("CLUSTER"),
	"cluster_name": {
		Description: "The cluster that is being granted on.",
		Type:        schema.TypeString,
		Required:    true,
		ForceNew:    true,
	},
	"region": RegionSchema(),
}

func GrantCluster() *schema.Resource {
	return &schema.Resource{
		Description: fmt.Sprintf(GrantDefinition, "cluster"),

		CreateContext: grantClusterCreate,
		ReadContext:   grantRead,
		DeleteContext: grantClusterDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: grantClusterSchema,
	}
}

func grantClusterCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	roleName := d.Get("role_name").(string)
	privilege := d.Get("privilege").(string)
	clusterName := d.Get("cluster_name").(string)

	obj := materialize.MaterializeObject{
		ObjectType: "CLUSTER",
		Name:       clusterName,
	}

	metaDb, region, err := utils.GetDBClientFromMeta(meta, d)
	if err != nil {
		return diag.FromErr(err)
	}

	b := materialize.NewPrivilegeBuilder(metaDb, roleName, privilege, obj)

	// grant resource
	if err := b.Grant(); err != nil {
		return diag.FromErr(err)
	}

	// set grant id
	roleId, err := materialize.RoleId(metaDb, roleName)
	if err != nil {
		return diag.FromErr(err)
	}

	i, err := materialize.ObjectId(metaDb, obj)
	if err != nil {
		return diag.FromErr(err)
	}

	key := b.GrantKey(string(region), i, roleId, privilege)
	d.SetId(key)

	return grantRead(ctx, d, meta)
}

func grantClusterDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	roleName := d.Get("role_name").(string)
	privilege := d.Get("privilege").(string)
	clusterName := d.Get("cluster_name").(string)

	metaDb, _, err := utils.GetDBClientFromMeta(meta, d)
	if err != nil {
		return diag.FromErr(err)
	}

	b := materialize.NewPrivilegeBuilder(
		metaDb,
		roleName,
		privilege,
		materialize.MaterializeObject{
			ObjectType: "CLUSTER",
			Name:       clusterName,
		},
	)

	if err := b.Revoke(); err != nil {
		return diag.FromErr(err)
	}

	return nil
}
