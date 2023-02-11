package datasources

import (
	"context"
	"database/sql"
	"errors"
	"log"

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
					},
				},
			},
		},
	}
}

func clusterReplicaRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*sqlx.DB)
	rows, err := conn.Query(`
		SELECT
			mz_cluster_replicas.id,
			mz_cluster_replicas.name,
			mz_clusters.name,
			mz_cluster_replicas.size,
			mz_cluster_replicas.availability_zone
		FROM mz_cluster_replicas
		JOIN mz_clusters
			ON mz_cluster_replicas.cluster_id = mz_clusters.id;
	`)
	if errors.Is(err, sql.ErrNoRows) {
		log.Printf("[DEBUG] no cluster replicas found in account")
		d.SetId("")
		return diag.FromErr(err)
	} else if err != nil {
		log.Println("[DEBUG] failed to list cluster replicas")
		d.SetId("")
		return diag.FromErr(err)
	}

	clusterReplicaFormats := []map[string]interface{}{}
	for rows.Next() {
		var id, name, cluster, size, availability_zone string
		rows.Scan(&id, &name, &cluster, &size, &availability_zone)

		clusterReplicaMap := map[string]interface{}{}

		clusterReplicaMap["id"] = id
		clusterReplicaMap["name"] = name
		clusterReplicaMap["cluster"] = cluster
		clusterReplicaMap["size"] = name
		clusterReplicaMap["availability_zone"] = name

		clusterReplicaFormats = append(clusterReplicaFormats, clusterReplicaMap)
	}

	if err := d.Set("cluster_replicas", clusterReplicaFormats); err != nil {
		return diag.FromErr(err)
	}

	d.SetId("cluster_replicas")
	return diags
}
