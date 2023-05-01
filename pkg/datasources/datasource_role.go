package datasources

import (
	"context"
	"database/sql"
	"errors"
	"log"

	"github.com/MaterializeInc/terraform-provider-materialize/pkg/materialize"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/jmoiron/sqlx"
)

func Role() *schema.Resource {
	return &schema.Resource{
		ReadContext: roleRead,
		Schema: map[string]*schema.Schema{
			"roles": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "The roles in the account",
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

func roleRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*sqlx.DB)

	q := materialize.ReadRoleDatasource()

	rows, err := conn.Query(q)
	if errors.Is(err, sql.ErrNoRows) {
		log.Printf("[DEBUG] no roles found in account")
		d.SetId("")
		return diag.FromErr(err)
	} else if err != nil {
		log.Println("[DEBUG] failed to list roles")
		d.SetId("")
		return diag.FromErr(err)
	}

	roleFormats := []map[string]interface{}{}
	for rows.Next() {
		var id, name string
		rows.Scan(&id, &name)

		roleMap := map[string]interface{}{}

		roleMap["id"] = id
		roleMap["name"] = name

		roleFormats = append(roleFormats, roleMap)
	}

	if err := d.Set("roles", roleFormats); err != nil {
		return diag.FromErr(err)
	}

	d.SetId("roles")
	return diags
}
