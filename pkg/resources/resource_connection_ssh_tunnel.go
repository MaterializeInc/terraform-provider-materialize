package resources

import (
	"context"
	"log"

	"github.com/MaterializeInc/terraform-provider-materialize/pkg/materialize"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/jmoiron/sqlx"
)

var connectionSshTunnelSchema = map[string]*schema.Schema{
	"name":               SchemaResourceName("connection", true, false),
	"schema_name":        SchemaResourceSchemaName("connection", false),
	"database_name":      SchemaResourceDatabaseName("connection", false),
	"qualified_sql_name": SchemaResourceQualifiedName("connection"),
	"connection_type":    SchemaResourceConnectionName(),
	"host": {
		Description: "The host of the SSH tunnel.",
		Type:        schema.TypeString,
		Required:    true,
	},
	"user": {
		Description: "The user of the SSH tunnel.",
		Type:        schema.TypeString,
		Required:    true,
	},
	"port": {
		Description: "The port of the SSH tunnel.",
		Type:        schema.TypeInt,
		Required:    true,
	},
}

func ConnectionSshTunnel() *schema.Resource {
	return &schema.Resource{
		Description: "The connection resource allows you to manage connections in Materialize.",

		CreateContext: connectionSshTunnelCreate,
		ReadContext:   connectionRead,
		UpdateContext: connectionSshTunnelUpdate,
		DeleteContext: connectionSshTunnelDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: connectionSshTunnelSchema,
	}
}

func connectionSshTunnelCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*sqlx.DB)

	connectionName := d.Get("name").(string)
	schemaName := d.Get("schema_name").(string)
	databaseName := d.Get("database_name").(string)

	builder := materialize.NewConnectionSshTunnelBuilder(connectionName, schemaName, databaseName)

	builder.SSHHost(d.Get("host").(string))
	builder.SSHUser(d.Get("user").(string))
	builder.SSHPort(d.Get("port").(int))

	qc := builder.Create()
	qr := builder.ReadId()

	if err := createResource(conn, d, qc, qr, "connection"); err != nil {
		return diag.FromErr(err)
	}
	return connectionRead(ctx, d, meta)
}

func connectionSshTunnelUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*sqlx.DB)
	connectionName := d.Get("name").(string)
	schemaName := d.Get("schema_name").(string)
	databaseName := d.Get("database_name").(string)

	if d.HasChange("name") {
		newConnectionName := d.Get("name").(string)
		q := materialize.NewConnectionSshTunnelBuilder(connectionName, schemaName, databaseName).Rename(newConnectionName)
		if err := ExecResource(conn, q); err != nil {
			log.Printf("[ERROR] could not execute query: %s", q)
			return diag.FromErr(err)
		}
	}

	return connectionRead(ctx, d, meta)
}

func connectionSshTunnelDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*sqlx.DB)
	connectionName := d.Get("name").(string)
	schemaName := d.Get("schema_name").(string)
	databaseName := d.Get("database_name").(string)

	q := materialize.NewConnectionSshTunnelBuilder(connectionName, schemaName, databaseName).Drop()

	if err := dropResource(conn, d, q, "connection"); err != nil {
		return diag.FromErr(err)
	}
	return nil
}
