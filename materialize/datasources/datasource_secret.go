package datasources

import (
	"context"
	"database/sql"
	"errors"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func Secret() *schema.Resource {
	return &schema.Resource{
		ReadContext: secretRead,
		Schema: map[string]*schema.Schema{
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
					},
				},
			},
		},
	}
}

func secretRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*sql.DB)
	rows, err := conn.Query(`
		SELECT mz_secrets.id, mz_secrets.name, mz_schemas.name
		FROM mz_secrets JOIN mz_schemas
			ON mz_secrets.schema_id = mz_schemas.id;
	`)
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
		var id, name, schema string
		rows.Scan(&id, &name)

		secretMap := map[string]interface{}{}

		secretMap["id"] = id
		secretMap["name"] = name
		secretMap["schema_name"] = schema

		secretFormats = append(secretFormats, secretMap)
	}

	if err := d.Set("secrets", secretFormats); err != nil {
		return diag.FromErr(err)
	}

	d.SetId("secrets")
	return diags
}
