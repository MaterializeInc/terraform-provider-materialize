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

func Sink() *schema.Resource {
	return &schema.Resource{
		ReadContext: sinkRead,
		Schema: map[string]*schema.Schema{
			"database_name": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Limit sinks to a specific database",
			},
			"schema_name": {
				Type:         schema.TypeString,
				Optional:     true,
				Description:  "Limit sinks to a specific schema within a specific database",
				RequiredWith: []string{"database_name"},
			},
			"sinks": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "The sinks in the account",
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

func sinkRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*sqlx.DB)

	schemaName := d.Get("schema_name").(string)
	databaseName := d.Get("database_name").(string)

	q := materialize.ReadSinkDatasource(databaseName, schemaName)

	rows, err := conn.Query(q)
	if errors.Is(err, sql.ErrNoRows) {
		log.Printf("[DEBUG] no sinks found in account")
		d.SetId("")
		return diag.FromErr(err)
	} else if err != nil {
		log.Println("[DEBUG] failed to list sinks")
		d.SetId("")
		return diag.FromErr(err)
	}

	sinkFormats := []map[string]interface{}{}
	for rows.Next() {
		var id, name, schema_name, database_name, sink_type, size, envelope_type, connection_name, cluster_name string
		rows.Scan(&id, &name, &schema_name, &database_name, &sink_type, &size, &envelope_type, &connection_name, &cluster_name)

		sinkMap := map[string]interface{}{}

		sinkMap["id"] = id
		sinkMap["name"] = name
		sinkMap["schema_name"] = schema_name
		sinkMap["database_name"] = database_name
		sinkMap["type"] = sink_type
		sinkMap["size"] = size
		sinkMap["envelope_type"] = envelope_type
		sinkMap["connection_name"] = connection_name
		sinkMap["cluster_name"] = cluster_name

		sinkFormats = append(sinkFormats, sinkMap)
	}

	if err := d.Set("sinks", sinkFormats); err != nil {
		return diag.FromErr(err)
	}

	SetId("sinks", databaseName, schemaName, d)

	return diags
}
