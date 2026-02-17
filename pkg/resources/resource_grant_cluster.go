package resources

import (
	"context"
	"fmt"
	"github.com/MaterializeInc/terraform-provider-materialize/pkg/materialize"

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
	return createGrant(ctx, d, meta, materialize.Cluster, "cluster_name")
}

func grantClusterDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return revokeGrant(d, meta, materialize.Cluster, "cluster_name")
}
