package datasources

import (
	"context"
	"database/sql"
	"errors"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func DatasourceCluster() *schema.Resource {
	return &schema.Resource{
		ReadContext: datasourceClusterReplicaRead,
		Schema: map[string]*schema.Schema{
			"coffees": {
				Type:     schema.TypeList,
				Computed: true,
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
					},
				},
			},
		},
	}
}

func datasourceClusterReplicaRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*sql.DB)

	rows, err := conn.Query(`SELECT * FROM mz_clusters;`)
	if errors.Is(err, sql.ErrNoRows) {
		// If not found, mark resource to be removed from statefile during apply or refresh
		log.Printf("[DEBUG] clusters not found")
		d.SetId("")
		return diags
	} else if err != nil {
		log.Printf("[DEBUG] unable to parse clusters")
		d.SetId("")
	}

	clusterFormats := []map[string]interface{}{}

	for rows.Next() {
		var id, name string
		if err := rows.Scan(&id, &name); err != nil {
			return nil
		}

		clusterMap := map[string]interface{}{}

		clusterMap["id"] = id
		clusterMap["name"] = name

		clusterFormats = append(clusterFormats, clusterMap)

	}

	d.SetId("clusters_read")
	d.Set("clusters", clusterFormats)

	return diags
}
