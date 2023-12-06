package datasources

import (
	"context"

	"github.com/MaterializeInc/terraform-provider-materialize/pkg/materialize"
	"github.com/MaterializeInc/terraform-provider-materialize/pkg/utils"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/jmoiron/sqlx"
)

func Type() *schema.Resource {
	return &schema.Resource{
		ReadContext: typeRead,
		Schema: map[string]*schema.Schema{
			"database_name": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Limit types to a specific database",
			},
			"schema_name": {
				Type:         schema.TypeString,
				Optional:     true,
				Description:  "Limit types to a specific schema within a specific database",
				RequiredWith: []string{"database_name"},
			},
			"types": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "The types in the account",
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
						"category": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func typeRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	schemaName := d.Get("schema_name").(string)
	databaseName := d.Get("database_name").(string)

	var diags diag.Diagnostics

	dataSource, err := materialize.ListTypes(meta.(*sqlx.DB), schemaName, databaseName)
	if err != nil {
		return diag.FromErr(err)
	}

	typeFormats := []map[string]interface{}{}
	for _, p := range dataSource {
		typeMap := map[string]interface{}{}

		typeMap["id"] = utils.TransformIdWithRegion(p.TypeId.String)
		typeMap["name"] = p.TypeName.String
		typeMap["schema_name"] = p.SchemaName.String
		typeMap["database_name"] = p.DatabaseName.String
		typeMap["category"] = p.Category.String

		typeFormats = append(typeFormats, typeMap)
	}

	if err := d.Set("types", typeFormats); err != nil {
		return diag.FromErr(err)
	}

	SetId("types", databaseName, schemaName, d)

	return diags
}
