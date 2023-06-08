package resources

import (
	"context"
	"database/sql"

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
		ForceNew:    true,
	},
	"user": {
		Description: "The user of the SSH tunnel.",
		Type:        schema.TypeString,
		Required:    true,
		ForceNew:    true,
	},
	"port": {
		Description: "The port of the SSH tunnel.",
		Type:        schema.TypeInt,
		Required:    true,
		ForceNew:    true,
	},
	"public_key_1": {
		Description: "The first public key associated with the SSH tunnel.",
		Type:        schema.TypeString,
		Computed:    true,
		ForceNew:    true,
	},
	"public_key_2": {
		Description: "The second public key associated with the SSH tunnel.",
		Type:        schema.TypeString,
		Computed:    true,
		ForceNew:    true,
	},
}

func ConnectionSshTunnel() *schema.Resource {
	return &schema.Resource{
		Description: "An SSH tunnel connection establishes a link to an SSH bastion server.",

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
	i := d.Id()

	s, err := materialize.ScanConnectionSshTunnel(meta.(*sqlx.DB), i)
	if err == sql.ErrNoRows {
		d.SetId("")
		return nil
	} else if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(i)

	if err := d.Set("name", s.ConnectionName.String); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("schema_name", s.SchemaName.String); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("database_name", s.DatabaseName.String); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("public_key_1", s.PublicKey1.String); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("public_key_2", s.PublicKey2.String); err != nil {
		return diag.FromErr(err)
	}

	b := materialize.Connection{ConnectionName: s.ConnectionName.String, SchemaName: s.SchemaName.String, DatabaseName: s.DatabaseName.String}
	if err := d.Set("qualified_sql_name", b.QualifiedName()); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func connectionSshTunnelCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	connectionName := d.Get("name").(string)
	schemaName := d.Get("schema_name").(string)
	databaseName := d.Get("database_name").(string)

	b := materialize.NewConnectionSshTunnelBuilder(meta.(*sqlx.DB), connectionName, schemaName, databaseName)

	b.SSHHost(d.Get("host").(string))
	b.SSHUser(d.Get("user").(string))
	b.SSHPort(d.Get("port").(int))

	// create resource
	if err := b.Create(); err != nil {
		return diag.FromErr(err)
	}

	// set id
	i, err := materialize.ConnectionId(meta.(*sqlx.DB), connectionName, schemaName, databaseName)
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

	b := materialize.NewConnectionSshTunnelBuilder(meta.(*sqlx.DB), connectionName, schemaName, databaseName)

	if d.HasChange("name") {
		_, newConnectionName := d.GetChange("name")
		b.Rename(newConnectionName.(string))
	}

	return connectionSshTunnelRead(ctx, d, meta)
}
