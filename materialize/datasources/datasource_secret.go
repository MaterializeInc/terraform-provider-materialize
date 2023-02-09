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
)

func Secret() *schema.Resource {
	return &schema.Resource{
		ReadContext: secretRead,
		Schema: map[string]*schema.Schema{
			"database_name": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Limit secrets to a specific database",
			},
			"schema_name": {
				Type:         schema.TypeString,
				Optional:     true,
				Description:  "Limit secrets to a specific schema within a specific database",
				RequiredWith: []string{"database_name"},
			},
			"secrets": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "The secrets in the account",
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

func secretQuery(databaseName, schemaName string) string {
	q := strings.Builder{}
	q.WriteString(`
		SELECT
			mz_secrets.id,
			mz_secrets.name,
			mz_schemas.name,
			mz_databases.name
		FROM mz_secrets
		JOIN mz_schemas
			ON mz_secrets.schema_id = mz_schemas.id
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

func secretRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*sql.DB)

	schemaName := d.Get("schema_name").(string)
	databaseName := d.Get("database_name").(string)
	q := secretQuery(databaseName, schemaName)

	rows, err := conn.Query(q)
	if errors.Is(err, sql.ErrNoRows) {
		log.Printf("[DEBUG] no secrets found in account")
		d.SetId("")
		return diag.FromErr(err)
	} else if err != nil {
		log.Println("[DEBUG] failed to list secrets")
		d.SetId("")
		return diag.FromErr(err)
	}

	secretFormats := []map[string]interface{}{}
	for rows.Next() {
		var id, name, schema_name, database_name string
		rows.Scan(&id, &name, &schema_name, &database_name)

		secretMap := map[string]interface{}{}

		secretMap["id"] = id
		secretMap["name"] = name
		secretMap["schema_name"] = schema_name
		secretMap["database_name"] = database_name

		secretFormats = append(secretFormats, secretMap)
	}

	if err := d.Set("secrets", secretFormats); err != nil {
		return diag.FromErr(err)
	}

	if databaseName != "" && schemaName != "" {
		id := fmt.Sprintf("%s|%s|secrets", databaseName, schemaName)
		d.SetId(id)
	} else if databaseName != "" {
		id := fmt.Sprintf("%s|secrets", databaseName)
		d.SetId(id)
	} else {
		d.SetId("secrets")
	}

	return diags
}
