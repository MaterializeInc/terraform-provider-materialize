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
	},
	"host": {
		Description: "The Postgres database hostname.",
		Type:        schema.TypeString,
		Required:    true,
	},
	"port": {
		Description: "The Postgres database port.",
		Type:        schema.TypeInt,
		Optional:    true,
		Default:     5432,
	},
	"user":                      ValueSecretSchema("user", "The Postgres database username.", true, false),
	"password":                  IdentifierSchema("password", "The Postgres database password.", false),
	"ssh_tunnel":                IdentifierSchema("ssh_tunnel", "The SSH tunnel configuration for the Postgres database.", false),
	"ssl_certificate_authority": ValueSecretSchema("ssl_certificate_authority", "The CA certificate for the Postgres database.", false, true),
	"ssl_certificate":           ValueSecretSchema("ssl_certificate", "The client certificate for the Postgres database.", false, true),
	"ssl_key":                   IdentifierSchema("ssl_key", "The client key for the Postgres database.", false),
	"ssl_mode": {
		Description: "The SSL mode for the Postgres database.",
		Type:        schema.TypeString,
		Optional:    true,
	},
	"aws_privatelink": IdentifierSchema("aws_privatelink", "The AWS PrivateLink configuration for the Postgres database.", false),
}

func ConnectionPostgres() *schema.Resource {
	return &schema.Resource{
		Description: "The connection resource allows you to manage connections in Materialize.",

		CreateContext: connectionPostgresCreate,
		ReadContext:   connectionRead,
		UpdateContext: connectionPostgresUpdate,
		DeleteContext: connectionPostgresDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: connectionPostgresSchema,
	}
}

func connectionPostgresCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*sqlx.DB)

	connectionName := d.Get("name").(string)
	schemaName := d.Get("schema_name").(string)
	databaseName := d.Get("database_name").(string)

	builder := materialize.NewConnectionPostgresBuilder(connectionName, schemaName, databaseName)

	if v, ok := d.GetOk("connection_type"); ok {
		builder.ConnectionType(v.(string))
	}

	if v, ok := d.GetOk("host"); ok {
		builder.PostgresHost(v.(string))
	}

	if v, ok := d.GetOk("port"); ok {
		builder.PostgresPort(v.(int))
	}

	if v, ok := d.GetOk("user"); ok {
		user := materialize.GetValueSecretStruct(databaseName, schemaName, v)
		builder.PostgresUser(user)
	}

	if v, ok := d.GetOk("password"); ok {
		pass := materialize.GetIdentifierSchemaStruct(databaseName, schemaName, v)
		builder.PostgresPassword(pass)
	}

	if v, ok := d.GetOk("database"); ok {
		builder.PostgresDatabase(v.(string))
	}

	if v, ok := d.GetOk("ssl_mode"); ok {
		builder.PostgresSSLMode(v.(string))
	}

	if v, ok := d.GetOk("ssl_certificate_authority"); ok {
		ssl_ca := materialize.GetValueSecretStruct(databaseName, schemaName, v)
		builder.PostgresSSLCa(ssl_ca)
	}

	if v, ok := d.GetOk("ssl_certificate"); ok {
		ssl_cert := materialize.GetValueSecretStruct(databaseName, schemaName, v)
		builder.PostgresSSLCert(ssl_cert)
	}

	if v, ok := d.GetOk("ssl_key"); ok {
		k := materialize.GetIdentifierSchemaStruct(databaseName, schemaName, v)
		builder.PostgresSSLKey(k)
	}

	if v, ok := d.GetOk("aws_privatelink"); ok {
		conn := materialize.GetIdentifierSchemaStruct(databaseName, schemaName, v)
		builder.PostgresAWSPrivateLink(conn)
	}

	if v, ok := d.GetOk("ssh_tunnel"); ok {
		conn := materialize.GetIdentifierSchemaStruct(databaseName, schemaName, v)
		builder.PostgresSSHTunnel(conn)
	}

	qc := builder.Create()
	qr := builder.ReadId()

	if err := createResource(conn, d, qc, qr, "connection"); err != nil {
		return diag.FromErr(err)
	}
	return connectionRead(ctx, d, meta)
}

func connectionPostgresUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*sqlx.DB)
	connectionName := d.Get("name").(string)
	schemaName := d.Get("schema_name").(string)
	databaseName := d.Get("database_name").(string)

	if d.HasChange("name") {
		_, newConnectionName := d.GetChange("name")
		q := materialize.NewConnectionPostgresBuilder(connectionName, schemaName, databaseName).Rename(newConnectionName.(string))
		if err := execResource(conn, q); err != nil {
			log.Printf("[ERROR] could not execute query: %s", q)
			return diag.FromErr(err)
		}
	}

	return connectionRead(ctx, d, meta)
}

func connectionPostgresDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*sqlx.DB)
	connectionName := d.Get("name").(string)
	schemaName := d.Get("schema_name").(string)
	databaseName := d.Get("database_name").(string)

	q := materialize.NewConnectionPostgresBuilder(connectionName, schemaName, databaseName).Drop()

	if err := dropResource(conn, d, q, "connection"); err != nil {
		return diag.FromErr(err)
	}
	return nil
}
