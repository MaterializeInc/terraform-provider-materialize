package datasources

import (
	"context"

	"github.com/MaterializeInc/terraform-provider-materialize/pkg/materialize"
	"github.com/MaterializeInc/terraform-provider-materialize/pkg/utils"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/jmoiron/sqlx"
)

func ClusterReplica() *schema.Resource {
	return &schema.Resource{
		ReadContext: clusterReplicaRead,
		Schema: map[string]*schema.Schema{
			"cluster_replicas": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "The cluster replicas in the account",
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
						"cluster": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"size": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"availability_zone": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"disk": {
							Type:     schema.TypeBool,
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func clusterReplicaRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	dataSource, err := materialize.ListClusterReplicas(meta.(*sqlx.DB))
	if err != nil {
		return diag.FromErr(err)
	}

	clusterReplicaFormats := []map[string]interface{}{}
	for _, p := range dataSource {
		clusterReplicaMap := map[string]interface{}{}

		clusterReplicaMap["id"] = p.ReplicaId.String
		clusterReplicaMap["name"] = p.ReplicaName.String
		clusterReplicaMap["cluster"] = p.ClusterName.String
		clusterReplicaMap["size"] = p.Size.String
		clusterReplicaMap["availability_zone"] = p.AvailabilityZone.String
		clusterReplicaMap["disk"] = p.Disk.Bool

		clusterReplicaFormats = append(clusterReplicaFormats, clusterReplicaMap)
	}

	if err := d.Set("cluster_replicas", clusterReplicaFormats); err != nil {
		return diag.FromErr(err)
	}

	d.SetId(utils.TransformIdWithRegion("cluster_replicas"))
	return diags
}
