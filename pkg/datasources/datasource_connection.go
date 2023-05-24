package datasources

import (
	"context"

	"github.com/MaterializeInc/terraform-provider-materialize/pkg/materialize"

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
				Description: "The connections in the account",
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
	schemaName := d.Get("schema_name").(string)
	databaseName := d.Get("database_name").(string)

	var diags diag.Diagnostics

	dataSource, err := materialize.ListConnections(meta.(*sqlx.DB), schemaName, databaseName)
	if err != nil {
		return diag.FromErr(err)
	}

	connectionFormats := []map[string]interface{}{}
	for _, p := range dataSource {
		connectionMap := map[string]interface{}{}

		connectionMap["id"] = p.ConnectionId.String
		connectionMap["name"] = p.ConnectionName.String
		connectionMap["schema_name"] = p.SchemaName.String
		connectionMap["database_name"] = p.DatabaseName.String
		connectionMap["type"] = p.ConnectionType.String

		connectionFormats = append(connectionFormats, connectionMap)
	}

	if err := d.Set("connections", connectionFormats); err != nil {
		return diag.FromErr(err)
	}

	SetId("connections", databaseName, schemaName, d)
	return diags
}
