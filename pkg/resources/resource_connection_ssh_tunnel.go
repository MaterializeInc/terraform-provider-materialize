package resources

import (
	"context"
	"database/sql"
	"log"

	"github.com/MaterializeInc/terraform-provider-materialize/pkg/materialize"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/jmoiron/sqlx"
)

var connectionSshTunnelSchema = map[string]*schema.Schema{
	"name":               ObjectNameSchema("connection", true, false),
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
	"ownership_role": OwnershipRoleSchema(),
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

	if err := d.Set("ownership_role", s.OwnerName.String); err != nil {
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

	o := materialize.ObjectSchemaStruct{Name: connectionName, SchemaName: schemaName, DatabaseName: databaseName}
	b := materialize.NewConnectionSshTunnelBuilder(meta.(*sqlx.DB), o)

	b.SSHHost(d.Get("host").(string))
	b.SSHUser(d.Get("user").(string))
	b.SSHPort(d.Get("port").(int))

	// create resource
	if err := b.Create(); err != nil {
		return diag.FromErr(err)
	}

	// ownership
	if v, ok := d.GetOk("ownership_role"); ok {
		ownership := materialize.NewOwnershipBuilder(meta.(*sqlx.DB), "CONNECTION", o)

		if err := ownership.Alter(v.(string)); err != nil {
			log.Printf("[DEBUG] resource failed ownership, dropping object: %s", o.Name)
			b.Drop()
			return diag.FromErr(err)
		}
	}

	// set id
	i, err := materialize.ConnectionId(meta.(*sqlx.DB), o)
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

	if d.HasChange("name") {
		oldName, newName := d.GetChange("name")

		o := materialize.ObjectSchemaStruct{Name: oldName.(string), SchemaName: schemaName, DatabaseName: databaseName}
		b := materialize.NewConnectionSshTunnelBuilder(meta.(*sqlx.DB), o)

		if err := b.Rename(newName.(string)); err != nil {
			return diag.FromErr(err)
		}
	}

	if d.HasChange("ownership_role") {
		_, newRole := d.GetChange("ownership_role")

		o := materialize.ObjectSchemaStruct{Name: connectionName, SchemaName: schemaName, DatabaseName: databaseName}
		b := materialize.NewOwnershipBuilder(meta.(*sqlx.DB), "CONNECTION", o)

		if err := b.Alter(newRole.(string)); err != nil {
			return diag.FromErr(err)
		}
	}

	return connectionSshTunnelRead(ctx, d, meta)
}
