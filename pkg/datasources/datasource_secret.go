package datasources

import (
	"context"

	"github.com/MaterializeInc/terraform-provider-materialize/pkg/materialize"
	"github.com/MaterializeInc/terraform-provider-materialize/pkg/utils"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/jmoiron/sqlx"
)

func Secret() *schema.Resource {
	return &schema.Resource{
		ReadContext: secretRead,
		Schema: map[string]*schema.Schema{
			"database_name": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Limit secrets to a specific database",
			},
			"schema_name": {
				Type:         schema.TypeString,
				Optional:     true,
				Description:  "Limit secrets to a specific schema within a specific database",
				RequiredWith: []string{"database_name"},
			},
			"secrets": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "The secrets in the account",
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
		},
	}
}

func secretRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	schemaName := d.Get("schema_name").(string)
	databaseName := d.Get("database_name").(string)

	var diags diag.Diagnostics

	dataSource, err := materialize.ListSecrets(meta.(*sqlx.DB), schemaName, databaseName)
	if err != nil {
		return diag.FromErr(err)
	}

	secretFormats := []map[string]interface{}{}
	for _, p := range dataSource {
		secretMap := map[string]interface{}{}

		secretMap["id"] = utils.TransformIdWithRegion(p.SecretId.String)
		secretMap["name"] = p.SecretName.String
		secretMap["schema_name"] = p.SchemaName.String
		secretMap["database_name"] = p.DatabaseName.String

		secretFormats = append(secretFormats, secretMap)
	}

	if err := d.Set("secrets", secretFormats); err != nil {
		return diag.FromErr(err)
	}

	SetId("secrets", databaseName, schemaName, d)
	return diags
}
