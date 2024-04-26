package datasources

import (
	"context"
	"fmt"

	"github.com/MaterializeInc/terraform-provider-materialize/pkg/materialize"
	"github.com/MaterializeInc/terraform-provider-materialize/pkg/utils"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
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
			"connection_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Limit connections to a specific connection ID",
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
			"region": RegionSchema(),
		},
	}
}

func formatConnectionData(connectionParams materialize.ConnectionParams) map[string]interface{} {
	return map[string]interface{}{
		"id":            connectionParams.ConnectionId.String,
		"name":          connectionParams.ConnectionName.String,
		"schema_name":   connectionParams.SchemaName.String,
		"database_name": connectionParams.DatabaseName.String,
		"type":          connectionParams.ConnectionType.String,
	}
}

func connectionRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	connectionId := d.Get("connection_id").(string)
	schemaName := d.Get("schema_name").(string)
	databaseName := d.Get("database_name").(string)

	metaDb, region, err := utils.GetDBClientFromMeta(meta, d)
	if err != nil {
		return diag.FromErr(err)
	}

	var connectionData materialize.ConnectionParams
	var connections []materialize.ConnectionParams
	var tfId string

	if connectionId != "" {
		connectionData, err = materialize.ScanConnection(metaDb, connectionId)
		if err != nil {
			return diag.FromErr(err)
		}
		connections = []materialize.ConnectionParams{connectionData}
		tfId = fmt.Sprintf("%s|%s", connectionData.ConnectionId.String, "connections")
	} else {
		connections, err = materialize.ListConnections(metaDb, schemaName, databaseName)
		tfId = "connections"
		if err != nil {
			return diag.FromErr(err)
		}
	}

	connectionFormats := make([]map[string]interface{}, len(connections))
	for i, conn := range connections {
		connectionFormats[i] = formatConnectionData(conn)
	}

	if err := d.Set("connections", connectionFormats); err != nil {
		return diag.FromErr(err)
	}

	SetId(string(region), tfId, databaseName, schemaName, d)
	return nil
}
