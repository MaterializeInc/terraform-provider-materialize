package datasources

import (
	"context"
	"database/sql"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func CurrentCluster() *schema.Resource {
	return &schema.Resource{
		ReadContext: currentClusterRead,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func currentClusterRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*sql.DB)
	var name string
	conn.QueryRow("SHOW CLUSTER;").Scan(&name)

	d.Set("name", name)
	d.SetId("current_cluster")

	return diags
}
