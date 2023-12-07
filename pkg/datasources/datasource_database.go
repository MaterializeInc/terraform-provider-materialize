package datasources

import (
	"context"

	"github.com/MaterializeInc/terraform-provider-materialize/pkg/materialize"
	"github.com/MaterializeInc/terraform-provider-materialize/pkg/utils"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/jmoiron/sqlx"
)

func Database() *schema.Resource {
	return &schema.Resource{
		ReadContext: databaseRead,
		Schema: map[string]*schema.Schema{
			"databases": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "The databases in the account",
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

func databaseRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	dataSource, err := materialize.ListDatabases(meta.(*sqlx.DB))
	if err != nil {
		return diag.FromErr(err)
	}

	databaseFormats := []map[string]interface{}{}
	for _, p := range dataSource {
		databaseMap := map[string]interface{}{}

		databaseMap["id"] = p.DatabaseId.String
		databaseMap["name"] = p.DatabaseName.String

		databaseFormats = append(databaseFormats, databaseMap)
	}

	if err := d.Set("databases", databaseFormats); err != nil {
		return diag.FromErr(err)
	}

	d.SetId(utils.TransformIdWithRegion("databases"))
	return diags
}
