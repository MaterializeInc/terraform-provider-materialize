package datasources

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"strings"

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

func connectionQuery(databaseName, schemaName string) string {
	q := strings.Builder{}
	q.WriteString(`
		SELECT
			mz_connections.id,
			mz_connections.name,
			mz_schemas.name AS schema_name,
			mz_databases.name AS database_name,
			mz_connections.type
		FROM mz_connections
		JOIN mz_schemas
			ON mz_connections.schema_id = mz_schemas.id
		JOIN mz_databases
			ON mz_schemas.database_id = mz_databases.id`)

	if databaseName != "" {
		q.WriteString(fmt.Sprintf(`
		WHERE mz_databases.name = '%s'`, databaseName))

		if schemaName != "" {
			q.WriteString(fmt.Sprintf(` AND mz_schemas.name = '%s'`, schemaName))
		}
	}

	q.WriteString(`;`)
	return q.String()
}

func connectionRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*sqlx.DB)

	schemaName := d.Get("schema_name").(string)
	databaseName := d.Get("database_name").(string)
	q := connectionQuery(databaseName, schemaName)

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

	if databaseName != "" && schemaName != "" {
		id := fmt.Sprintf("%s|%s|connections", databaseName, schemaName)
		d.SetId(id)
	} else if databaseName != "" {
		id := fmt.Sprintf("%s|connections", databaseName)
		d.SetId(id)
	} else {
		d.SetId("connections")
	}

	return diags
}
