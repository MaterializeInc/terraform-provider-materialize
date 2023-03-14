package datasources

import (
	"context"
	"database/sql"
	"errors"
	"log"

	"terraform-materialize/materialize/materialize"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/jmoiron/sqlx"
)

func Secret() *schema.Resource {
	return &schema.Resource{
		ReadContext: SecretRead,
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

func SecretRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*sqlx.DB)

	schemaName := d.Get("schema_name").(string)
	databaseName := d.Get("database_name").(string)
	q := materialize.ReadSecretDatasource(databaseName, schemaName)

	rows, err := conn.Query(q)
	if errors.Is(err, sql.ErrNoRows) {
		log.Printf("[DEBUG] no secrets found in account")
		d.SetId("")
		return diag.FromErr(err)
	} else if err != nil {
		log.Println("[DEBUG] failed to list secrets")
		d.SetId("")
		return diag.FromErr(err)
	}

	secretFormats := []map[string]interface{}{}
	for rows.Next() {
		var id, name, schema_name, database_name string
		rows.Scan(&id, &name, &schema_name, &database_name)

		secretMap := map[string]interface{}{}

		secretMap["id"] = id
		secretMap["name"] = name
		secretMap["schema_name"] = schema_name
		secretMap["database_name"] = database_name

		secretFormats = append(secretFormats, secretMap)
	}

	if err := d.Set("secrets", secretFormats); err != nil {
		return diag.FromErr(err)
	}

	SetId("secrets", databaseName, schemaName, d)
	return diags
}
