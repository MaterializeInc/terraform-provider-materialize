package datasources

import (
	"context"
	"database/sql"
	"errors"
	"log"

	"github.com/MaterializeInc/terraform-materialize-provider/pkg/materialize"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/jmoiron/sqlx"
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
					},
				},
			},
		},
	}
}

func clusterRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*sqlx.DB)

	q := materialize.ReadClusterDatasource()

	rows, err := conn.Query(q)
	if errors.Is(err, sql.ErrNoRows) {
		log.Printf("[DEBUG] no clusters found in account")
		d.SetId("")
		return diag.FromErr(err)
	} else if err != nil {
		log.Println("[DEBUG] failed to list clusters")
		d.SetId("")
		return diag.FromErr(err)
	}

	clusterFormats := []map[string]interface{}{}
	for rows.Next() {
		var id, name string
		rows.Scan(&id, &name)

		clusterMap := map[string]interface{}{}

		clusterMap["id"] = id
		clusterMap["name"] = name

		clusterFormats = append(clusterFormats, clusterMap)
	}

	if err := d.Set("clusters", clusterFormats); err != nil {
		return diag.FromErr(err)
	}

	d.SetId("clusters")
	return diags
}
