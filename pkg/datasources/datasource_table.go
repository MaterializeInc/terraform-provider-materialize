package datasources

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"terraform-materialize/pkg/materialize"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/jmoiron/sqlx"
)

func Table() *schema.Resource {
	return &schema.Resource{
		ReadContext: tableRead,
		Schema: map[string]*schema.Schema{
			"database_name": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Limit tables to a specific database",
			},
			"schema_name": {
				Type:         schema.TypeString,
				Optional:     true,
				Description:  "Limit tables to a specific schema within a specific database",
				RequiredWith: []string{"database_name"},
			},
			"tables": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "The tables in the account",
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

func tableRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*sqlx.DB)

	schemaName := d.Get("schema_name").(string)
	databaseName := d.Get("database_name").(string)
	q := materialize.ReadTableDatasource(databaseName, schemaName)

	rows, err := conn.Query(q)
	if errors.Is(err, sql.ErrNoRows) {
		log.Printf("[DEBUG] no tables found in account")
		d.SetId("")
		return diag.FromErr(err)
	} else if err != nil {
		log.Println("[DEBUG] failed to list tables")
		d.SetId("")
		return diag.FromErr(err)
	}

	tableFormats := []map[string]interface{}{}
	for rows.Next() {
		var id, name, schema_name, database_name string
		rows.Scan(&id, &name, &schema_name, &database_name)

		tableMap := map[string]interface{}{}

		tableMap["id"] = id
		tableMap["name"] = name
		tableMap["schema_name"] = schema_name
		tableMap["database_name"] = database_name

		tableFormats = append(tableFormats, tableMap)
	}

	if err := d.Set("tables", tableFormats); err != nil {
		return diag.FromErr(err)
	}

	SetId("tables", databaseName, schemaName, d)

	return diags
}
