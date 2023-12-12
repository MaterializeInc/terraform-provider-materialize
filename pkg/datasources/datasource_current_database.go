package datasources

import (
	"context"

	"github.com/MaterializeInc/terraform-provider-materialize/pkg/utils"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func CurrentDatabase() *schema.Resource {
	return &schema.Resource{
		ReadContext: currentDatabaseRead,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"region": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func currentDatabaseRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	metaDb, region, err := utils.GetDBClientFromMeta(meta, d)
	if err != nil {
		return diag.FromErr(err)
	}
	conn := metaDb
	var name string
	conn.QueryRow("SHOW DATABASE;").Scan(&name)

	d.Set("name", name)
	d.SetId(utils.TransformIdWithRegion(string(region), "current_database"))

	return diags
}
