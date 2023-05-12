package resources

import (
	"context"

	"github.com/MaterializeInc/terraform-provider-materialize/pkg/materialize"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/jmoiron/sqlx"
)

var connectionConfluentSchemaRegistrySchema = map[string]*schema.Schema{
	"name":               NameSchema("connection", true, false),
	"schema_name":        SchemaNameSchema("connection", false),
	"database_name":      DatabaseNameSchema("connection", false),
	"qualified_sql_name": QualifiedNameSchema("connection"),
	"url": {
		Description: "The URL of the Confluent Schema Registry.",
		Type:        schema.TypeString,
		Required:    true,
	},
	"ssl_certificate_authority": ValueSecretSchema("ssl_certificate_authority", "The CA certificate for the Confluent Schema Registry.", false, true),
	"ssl_certificate":           ValueSecretSchema("ssl_certificate", "The client certificate for the Confluent Schema Registry.", false, true),
	"ssl_key":                   IdentifierSchema("ssl_key", "The client key for the Confluent Schema Registry.", false),
	"password":                  IdentifierSchema("password", "The password for the Confluent Schema Registry.", false),
	"username":                  ValueSecretSchema("username", "The username for the Confluent Schema Registry.", false, true),
	"ssh_tunnel":                IdentifierSchema("ssh_tunnel", "The SSH tunnel configuration for the Confluent Schema Registry.", false),
	"aws_privatelink":           IdentifierSchema("aws_privatelink", "The AWS PrivateLink configuration for the Confluent Schema Registry.", false),
}

func ConnectionConfluentSchemaRegistry() *schema.Resource {
	return &schema.Resource{
		Description: "The connection resource allows you to manage connections in Materialize.",

		CreateContext: connectionConfluentSchemaRegistryCreate,
		ReadContext:   connectionRead,
		UpdateContext: connectionUpdate,
		DeleteContext: connectionDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: connectionConfluentSchemaRegistrySchema,
	}
}

func connectionConfluentSchemaRegistryCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	connectionName := d.Get("name").(string)
	schemaName := d.Get("schema_name").(string)
	databaseName := d.Get("database_name").(string)

	builder := materialize.NewConnectionConfluentSchemaRegistryBuilder(meta.(*sqlx.DB), connectionName, schemaName, databaseName)

	if v, ok := d.GetOk("url"); ok {
		builder.ConfluentSchemaRegistryUrl(v.(string))
	}

	if v, ok := d.GetOk("ssl_certificate_authority"); ok {
		ssl_ca := materialize.GetValueSecretStruct(databaseName, schemaName, v)
		builder.ConfluentSchemaRegistrySSLCa(ssl_ca)
	}

	if v, ok := d.GetOk("ssl_certificate"); ok {
		ssl_cert := materialize.GetValueSecretStruct(databaseName, schemaName, v)
		builder.ConfluentSchemaRegistrySSLCert(ssl_cert)
	}

	if v, ok := d.GetOk("ssl_key"); ok {
		key := materialize.GetIdentifierSchemaStruct(databaseName, schemaName, v)
		builder.ConfluentSchemaRegistrySSLKey(key)
	}

	if v, ok := d.GetOk("username"); ok {
		user := materialize.GetValueSecretStruct(databaseName, schemaName, v)
		builder.ConfluentSchemaRegistryUsername(user)
	}

	if v, ok := d.GetOk("password"); ok {
		pass := materialize.GetIdentifierSchemaStruct(databaseName, schemaName, v)
		builder.ConfluentSchemaRegistryPassword(pass)
	}

	if v, ok := d.GetOk("ssh_tunnel"); ok {
		conn := materialize.GetIdentifierSchemaStruct(databaseName, schemaName, v)
		builder.ConfluentSchemaRegistrySSHTunnel(conn)
	}

	if v, ok := d.GetOk("aws_privatelink"); ok {
		conn := materialize.GetIdentifierSchemaStruct(databaseName, schemaName, v)
		builder.ConfluentSchemaRegistryAWSPrivateLink(conn)
	}

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
	return connectionRead(ctx, d, meta)
}
