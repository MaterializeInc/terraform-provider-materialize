package datasources

import (
	"context"
	"database/sql"
	"errors"
	"log"

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

	conn := meta.(*sqlx.DB)
	rows, err := conn.Query(`SELECT id, name FROM mz_databases;`)
	if errors.Is(err, sql.ErrNoRows) {
		log.Printf("[DEBUG] no databases found in account")
		d.SetId("")
		return diag.FromErr(err)
	} else if err != nil {
		log.Println("[DEBUG] failed to list databases")
		d.SetId("")
		return diag.FromErr(err)
	}

	databaseFormats := []map[string]interface{}{}
	for rows.Next() {
		var id, name string
		rows.Scan(&id, &name)

		databaseMap := map[string]interface{}{}

		databaseMap["id"] = id
		databaseMap["name"] = name

		databaseFormats = append(databaseFormats, databaseMap)
	}

	if err := d.Set("databases", databaseFormats); err != nil {
		return diag.FromErr(err)
	}

	d.SetId("databases")
	return diags
}
