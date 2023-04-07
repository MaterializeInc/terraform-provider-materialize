package datasources

import (
	"context"
	"database/sql"
	"errors"
	"log"

	"github.com/MaterializeInc/terraform-materialize-provider/pkg/materialize"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/jmoiron/sqlx"
)

func MaterializedView() *schema.Resource {
	return &schema.Resource{
		ReadContext: materializedViewRead,
		Schema: map[string]*schema.Schema{
			"database_name": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Limit materialized views to a specific database",
			},
			"schema_name": {
				Type:         schema.TypeString,
				Optional:     true,
				Description:  "Limit materialized views to a specific schema within a specific database",
				RequiredWith: []string{"database_name"},
			},
			"materialized_views": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "The materialized views in the account",
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

func materializedViewRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*sqlx.DB)

	schemaName := d.Get("schema_name").(string)
	databaseName := d.Get("database_name").(string)
	q := materialize.ReadMaterializedViewDatasource(databaseName, schemaName)

	rows, err := conn.Query(q)
	if errors.Is(err, sql.ErrNoRows) {
		log.Printf("[DEBUG] no materialized views found in account")
		d.SetId("")
		return diag.FromErr(err)
	} else if err != nil {
		log.Println("[DEBUG] failed to list materialized views")
		d.SetId("")
		return diag.FromErr(err)
	}

	viewFormats := []map[string]interface{}{}
	for rows.Next() {
		var id, name, schema_name, database_name string
		rows.Scan(&id, &name, &schema_name, &database_name)

		tableMap := map[string]interface{}{}

		tableMap["id"] = id
		tableMap["name"] = name
		tableMap["schema_name"] = schema_name
		tableMap["database_name"] = database_name

		viewFormats = append(viewFormats, tableMap)
	}

	if err := d.Set("materialized_views", viewFormats); err != nil {
		return diag.FromErr(err)
	}

	SetId("materialized_views", databaseName, schemaName, d)

	return diags
}
