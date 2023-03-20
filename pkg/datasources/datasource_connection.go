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

func Connection() *schema.Resource {
	return &schema.Resource{
		ReadContext: connectionRead,
		Schema: map[string]*schema.Schema{
			"database_name": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Limit connections to a specific database",
			},
			"schema_name": {
				Type:         schema.TypeString,
				Optional:     true,
				Description:  "Limit connections to a specific schema within a specific database",
				RequiredWith: []string{"database_name"},
			},
			"connections": {
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
					},
				},
			},
		},
	}
}

func connectionRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*sqlx.DB)

	schemaName := d.Get("schema_name").(string)
	databaseName := d.Get("database_name").(string)
	q := materialize.ReadConnectionDatasource(databaseName, schemaName)

	rows, err := conn.Query(q)
	if errors.Is(err, sql.ErrNoRows) {
		log.Printf("[DEBUG] no connections found in account")
		d.SetId("")
		return diag.FromErr(err)
	} else if err != nil {
		log.Println("[DEBUG] failed to list connections")
		d.SetId("")
		return diag.FromErr(err)
	}

	connectionFormats := []map[string]interface{}{}
	for rows.Next() {
		var id, name, schema_name, database_name, connection_type string
		rows.Scan(&id, &name, &schema_name, &database_name, &connection_type)

		connectionMap := map[string]interface{}{}

		connectionMap["id"] = id
		connectionMap["name"] = name
		connectionMap["schema_name"] = schema_name
		connectionMap["database_name"] = database_name
		connectionMap["type"] = connection_type

		connectionFormats = append(connectionFormats, connectionMap)
	}

	if err := d.Set("connections", connectionFormats); err != nil {
		return diag.FromErr(err)
	}

	SetId("connections", databaseName, schemaName, d)

	return diags
}
