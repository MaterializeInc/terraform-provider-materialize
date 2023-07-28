package resources

import (
	"context"
	"log"

	"github.com/MaterializeInc/terraform-provider-materialize/pkg/materialize"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/jmoiron/sqlx"
)

var connectionPostgresSchema = map[string]*schema.Schema{
	"name":               NameSchema("connection", true, false),
	"schema_name":        SchemaNameSchema("connection", false),
	"database_name":      DatabaseNameSchema("connection", false),
	"qualified_sql_name": QualifiedNameSchema("connection"),
	"database": {
		Description: "The target Postgres database.",
		Type:        schema.TypeString,
		Required:    true,
		ForceNew:    true,
	},
	"host": {
		Description: "The Postgres database hostname.",
		Type:        schema.TypeString,
		Required:    true,
		ForceNew:    true,
	},
	"port": {
		Description: "The Postgres database port.",
		Type:        schema.TypeInt,
		Optional:    true,
		Default:     5432,
		ForceNew:    true,
	},
	"user":                      ValueSecretSchema("user", "The Postgres database username.", true),
	"password":                  IdentifierSchema("password", "The Postgres database password.", false),
	"ssh_tunnel":                IdentifierSchema("ssh_tunnel", "The SSH tunnel configuration for the Postgres database.", false),
	"ssl_certificate_authority": ValueSecretSchema("ssl_certificate_authority", "The CA certificate for the Postgres database.", false),
	"ssl_certificate":           ValueSecretSchema("ssl_certificate", "The client certificate for the Postgres database.", false),
	"ssl_key":                   IdentifierSchema("ssl_key", "The client key for the Postgres database.", false),
	"ssl_mode": {
		Description: "The SSL mode for the Postgres database.",
		Type:        schema.TypeString,
		Optional:    true,
		ForceNew:    true,
	},
	"aws_privatelink": IdentifierSchema("aws_privatelink", "The AWS PrivateLink configuration for the Postgres database.", false),
	"validate":        ValidateConnection(),
	"ownership_role":  OwnershipRole(),
}

func ConnectionPostgres() *schema.Resource {
	return &schema.Resource{
		Description: "A Postgres connection establishes a link to a single database of a PostgreSQL server.",

		CreateContext: connectionPostgresCreate,
		ReadContext:   connectionRead,
		UpdateContext: connectionUpdate,
		DeleteContext: connectionDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: connectionPostgresSchema,
	}
}

func connectionPostgresCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	connectionName := d.Get("name").(string)
	schemaName := d.Get("schema_name").(string)
	databaseName := d.Get("database_name").(string)

	o := materialize.ObjectSchemaStruct{Name: connectionName, SchemaName: schemaName, DatabaseName: databaseName}
	b := materialize.NewConnectionPostgresBuilder(meta.(*sqlx.DB), o)

	if v, ok := d.GetOk("connection_type"); ok {
		b.ConnectionType(v.(string))
	}

	if v, ok := d.GetOk("host"); ok {
		b.PostgresHost(v.(string))
	}

	if v, ok := d.GetOk("port"); ok {
		b.PostgresPort(v.(int))
	}

	if v, ok := d.GetOk("user"); ok {
		user := materialize.GetValueSecretStruct(databaseName, schemaName, v)
		b.PostgresUser(user)
	}

	if v, ok := d.GetOk("password"); ok {
		pass := materialize.GetIdentifierSchemaStruct(databaseName, schemaName, v)
		b.PostgresPassword(pass)
	}

	if v, ok := d.GetOk("database"); ok {
		b.PostgresDatabase(v.(string))
	}

	if v, ok := d.GetOk("ssl_mode"); ok {
		b.PostgresSSLMode(v.(string))
	}

	if v, ok := d.GetOk("ssl_certificate_authority"); ok {
		ssl_ca := materialize.GetValueSecretStruct(databaseName, schemaName, v)
		b.PostgresSSLCa(ssl_ca)
	}

	if v, ok := d.GetOk("ssl_certificate"); ok {
		ssl_cert := materialize.GetValueSecretStruct(databaseName, schemaName, v)
		b.PostgresSSLCert(ssl_cert)
	}

	if v, ok := d.GetOk("ssl_key"); ok {
		k := materialize.GetIdentifierSchemaStruct(databaseName, schemaName, v)
		b.PostgresSSLKey(k)
	}

	if v, ok := d.GetOk("aws_privatelink"); ok {
		conn := materialize.GetIdentifierSchemaStruct(databaseName, schemaName, v)
		b.PostgresAWSPrivateLink(conn)
	}

	if v, ok := d.GetOk("ssh_tunnel"); ok {
		conn := materialize.GetIdentifierSchemaStruct(databaseName, schemaName, v)
		b.PostgresSSHTunnel(conn)
	}

	if v, ok := d.GetOk("validate"); ok {
		b.Validate(v.(bool))
	}

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

	return connectionRead(ctx, d, meta)
}
