package datasources

import (
	"context"
	"database/sql"
	"errors"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func Schema() *schema.Resource {
	return &schema.Resource{
		ReadContext: schemaRead,
		Schema: map[string]*schema.Schema{
			"schemas": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "The schemas in the account",
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

func schemaRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*sql.DB)
	rows, err := conn.Query(`
		SELECT
			mz_schemas.id,
			mz_schemas.name,
			mz_databases.name
		FROM mz_schemas JOIN mz_databases
			ON mz_schemas.database_id = mz_databases.id;
	`)
	if errors.Is(err, sql.ErrNoRows) {
		log.Printf("[DEBUG] no schemas found in account")
		d.SetId("")
		return diag.FromErr(err)
	} else if err != nil {
		log.Println("[DEBUG] failed to list schemas")
		d.SetId("")
		return diag.FromErr(err)
	}

	schemasFormats := []map[string]interface{}{}
	for rows.Next() {
		var id, name, database_name string
		rows.Scan(&id, &name, &database_name)

		schemaMap := map[string]interface{}{}

		schemaMap["id"] = id
		schemaMap["name"] = name
		schemaMap["database_name"] = database_name

		schemasFormats = append(schemasFormats, schemaMap)
	}

	if err := d.Set("schemas", schemasFormats); err != nil {
		return diag.FromErr(err)
	}

	d.SetId("schemas")
	return diags
}
