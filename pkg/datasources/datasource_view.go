package datasources

import (
	"context"

	"github.com/MaterializeInc/terraform-provider-materialize/pkg/materialize"
	"github.com/MaterializeInc/terraform-provider-materialize/pkg/utils"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func View() *schema.Resource {
	return &schema.Resource{
		ReadContext: viewRead,
		Schema: map[string]*schema.Schema{
			"database_name": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Limit views to a specific database",
			},
			"schema_name": {
				Type:         schema.TypeString,
				Optional:     true,
				Description:  "Limit views to a specific schema within a specific database",
				RequiredWith: []string{"database_name"},
			},
			"views": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "The views in the account",
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
						"schema_name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"database_name": {
							Type:     schema.TypeString,
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

func viewRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	schemaName := d.Get("schema_name").(string)
	databaseName := d.Get("database_name").(string)

	var diags diag.Diagnostics

	metaDb, err := utils.GetDBClientFromMeta(meta, d)
	if err != nil {
		return diag.FromErr(err)
	}
	dataSource, err := materialize.ListViews(metaDb, schemaName, databaseName)
	if err != nil {
		return diag.FromErr(err)
	}

	viewFormats := []map[string]interface{}{}
	for _, p := range dataSource {
		viewMap := map[string]interface{}{}

		viewMap["id"] = p.ViewId.String
		viewMap["name"] = p.ViewName.String
		viewMap["schema_name"] = p.SchemaName.String
		viewMap["database_name"] = p.DatabaseName.String

		viewFormats = append(viewFormats, viewMap)
	}

	if err := d.Set("views", viewFormats); err != nil {
		return diag.FromErr(err)
	}

	SetId("views", databaseName, schemaName, d)

	return diags
}
