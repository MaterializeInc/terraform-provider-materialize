package datasources

import (
	"context"

	"github.com/MaterializeInc/terraform-provider-materialize/pkg/materialize"
	"github.com/MaterializeInc/terraform-provider-materialize/pkg/utils"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func Cluster() *schema.Resource {
	return &schema.Resource{
		ReadContext: clusterRead,
		Schema: map[string]*schema.Schema{
			"clusters": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "The clusters in the account",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"managed": {
							Type:     schema.TypeBool,
							Computed: true,
						},
						"size": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"replication_factor": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"disk": {
							Type:     schema.TypeBool,
							Computed: true,
						},
					},
				},
			},
			"region": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func clusterRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	metaDb, err := utils.GetDBClientFromMeta(meta, d)
	if err != nil {
		return diag.FromErr(err)
	}
	dataSource, err := materialize.ListClusters(metaDb)
	if err != nil {
		return diag.FromErr(err)
	}

	clusterFormats := []map[string]interface{}{}
	for _, p := range dataSource {
		clusterMap := map[string]interface{}{}

		clusterMap["id"] = p.ClusterId.String
		clusterMap["name"] = p.ClusterName.String
		clusterMap["managed"] = p.Managed.Bool
		clusterMap["size"] = p.Size.String
		clusterMap["replication_factor"] = p.ReplicationFactor.Int64
		clusterMap["disk"] = p.Disk.Bool

		clusterFormats = append(clusterFormats, clusterMap)
	}

	if err := d.Set("clusters", clusterFormats); err != nil {
		return diag.FromErr(err)
	}

	d.SetId(utils.TransformIdWithRegion("clusters"))
	return diags
}
