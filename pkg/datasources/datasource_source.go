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

func Source() *schema.Resource {
	return &schema.Resource{
		ReadContext: sourceRead,
		Schema: map[string]*schema.Schema{
			"database_name": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Limit sources to a specific database",
			},
			"schema_name": {
				Type:         schema.TypeString,
				Optional:     true,
				Description:  "Limit sources to a specific schema within a specific database",
				RequiredWith: []string{"database_name"},
			},
			"sources": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "The sources in the account",
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
						"type": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"size": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"envelope_type": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"connection_name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"cluster_name": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func sourceRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*sqlx.DB)

	schemaName := d.Get("schema_name").(string)
	databaseName := d.Get("database_name").(string)

	q := materialize.ReadSourceDatasource(databaseName, schemaName)

	rows, err := conn.Query(q)
	if errors.Is(err, sql.ErrNoRows) {
		log.Printf("[DEBUG] no sources found in account")
		d.SetId("")
		return diag.FromErr(err)
	} else if err != nil {
		log.Println("[DEBUG] failed to list sources")
		d.SetId("")
		return diag.FromErr(err)
	}

	sourceFormats := []map[string]interface{}{}
	for rows.Next() {
		var id, name, schema_name, database_name, source_type, size, envelope_type, connection_name, cluster_name string
		rows.Scan(&id, &name, &schema_name, &database_name, &source_type, &size, &envelope_type, &connection_name, &cluster_name)

		sourceMap := map[string]interface{}{}

		sourceMap["id"] = id
		sourceMap["name"] = name
		sourceMap["schema_name"] = schema_name
		sourceMap["database_name"] = database_name
		sourceMap["type"] = source_type
		sourceMap["size"] = size
		sourceMap["envelope_type"] = envelope_type
		sourceMap["connection_name"] = connection_name
		sourceMap["cluster_name"] = cluster_name

		sourceFormats = append(sourceFormats, sourceMap)
	}

	if err := d.Set("sources", sourceFormats); err != nil {
		return diag.FromErr(err)
	}

	SetId("sources", databaseName, schemaName, d)

	return diags
}
