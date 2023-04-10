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

func Index() *schema.Resource {
	return &schema.Resource{
		ReadContext: indexRead,
		Schema: map[string]*schema.Schema{
			"database_name": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Limit indexes to a specific database",
			},
			"schema_name": {
				Type:         schema.TypeString,
				Optional:     true,
				Description:  "Limit indexes to a specific schema within a specific database",
				RequiredWith: []string{"database_name"},
			},
			"indexes": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "The indexes in the account",
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
						"obj_name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"obj_schema": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"obj_database": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func indexRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*sqlx.DB)

	schemaName := d.Get("schema_name").(string)
	databaseName := d.Get("database_name").(string)
	q := materialize.ReadIndexDatasource(databaseName, schemaName)

	rows, err := conn.Query(q)
	if errors.Is(err, sql.ErrNoRows) {
		log.Printf("[DEBUG] no indexes found in account")
		d.SetId("")
		return diag.FromErr(err)
	} else if err != nil {
		log.Println("[DEBUG] failed to list indexes")
		d.SetId("")
		return diag.FromErr(err)
	}

	indexFormats := []map[string]interface{}{}
	for rows.Next() {
		var id, name, obj, schema, database string
		rows.Scan(&id, &name, &obj, &schema, &database)

		indexMap := map[string]interface{}{}

		indexMap["id"] = id
		indexMap["name"] = name
		indexMap["obj_name"] = obj
		indexMap["obj_schema"] = schema
		indexMap["obj_database"] = database

		indexFormats = append(indexFormats, indexMap)
	}

	if err := d.Set("indexes", indexFormats); err != nil {
		return diag.FromErr(err)
	}

	SetId("indexes", databaseName, schemaName, d)

	return diags
}
