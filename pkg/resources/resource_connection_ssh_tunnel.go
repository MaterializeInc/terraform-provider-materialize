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
	"name":               NameSchema("connection", true, false),
	"schema_name":        SchemaNameSchema("connection", false),
	"database_name":      DatabaseNameSchema("connection", false),
	"qualified_sql_name": QualifiedNameSchema("connection"),
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
	"public_key_1": {
		Description: "The first public key associated with the SSH tunnel.",
		Type:        schema.TypeString,
		Computed:    true,
	},
	"public_key_2": {
		Description: "The second public key associated with the SSH tunnel.",
		Type:        schema.TypeString,
		Computed:    true,
	},
}

func ConnectionSshTunnel() *schema.Resource {
	return &schema.Resource{
		Description: "The connection resource allows you to manage connections in Materialize.",

		CreateContext: connectionSshTunnelCreate,
		ReadContext:   connectionSshTunnelRead,
		UpdateContext: connectionSshTunnelUpdate,
		DeleteContext: connectionDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: connectionSshTunnelSchema,
	}
}

func connectionSshTunnelRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	connectionName := d.Get("name").(string)
	schemaName := d.Get("schema_name").(string)
	databaseName := d.Get("database_name").(string)

	builder := materialize.NewConnectionSshTunnelBuilder(meta.(*sqlx.DB), connectionName, schemaName, databaseName)

	i := d.Id()
	params, _ := builder.SshTunnelParams(i)

	if err := d.Set("name", params.ConnectionName); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("schema_name", params.SchemaName); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("database_name", params.DatabaseName); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("public_key_1", params.PublicKey1); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("public_key_2", params.PublicKey2); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("qualified_sql_name", builder.QualifiedName()); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func connectionSshTunnelCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	connectionName := d.Get("name").(string)
	schemaName := d.Get("schema_name").(string)
	databaseName := d.Get("database_name").(string)

	builder := materialize.NewConnectionSshTunnelBuilder(meta.(*sqlx.DB), connectionName, schemaName, databaseName)

	builder.SSHHost(d.Get("host").(string))
	builder.SSHUser(d.Get("user").(string))
	builder.SSHPort(d.Get("port").(int))

	// create resource
	if err := builder.Create(); err != nil {
		return diag.FromErr(err)
	}

	// set id
	i, err := builder.ReadId()
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(i)
	return connectionSshTunnelRead(ctx, d, meta)
}

func connectionSshTunnelUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	connectionName := d.Get("name").(string)
	schemaName := d.Get("schema_name").(string)
	databaseName := d.Get("database_name").(string)

	builder := materialize.NewConnectionSshTunnelBuilder(meta.(*sqlx.DB), connectionName, schemaName, databaseName)

	if d.HasChange("name") {
		_, newName := d.GetChange("name")

		if err := builder.Rename(newName.(string)); err != nil {
			log.Printf("[ERROR] could not rename connection %s", connectionName)
			return diag.FromErr(err)
		}
	}

	return connectionSshTunnelRead(ctx, d, meta)
}
