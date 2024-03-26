package resources

import (
	"context"
	"log"

	"github.com/MaterializeInc/terraform-provider-materialize/pkg/materialize"
	"github.com/MaterializeInc/terraform-provider-materialize/pkg/utils"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

var connectionConfluentSchemaRegistrySchema = map[string]*schema.Schema{
	"name":               ObjectNameSchema("connection", true, false),
	"schema_name":        SchemaNameSchema("connection", false),
	"database_name":      DatabaseNameSchema("connection", false),
	"qualified_sql_name": QualifiedNameSchema("connection"),
	"comment":            CommentSchema(false),
	"url": {
		Description: "The URL of the Confluent Schema Registry.",
		Type:        schema.TypeString,
		Required:    true,
	},
	"ssl_certificate_authority": ValueSecretSchema("ssl_certificate_authority", "The CA certificate for the Confluent Schema Registry.", false, true),
	"ssl_certificate":           ValueSecretSchema("ssl_certificate", "The client certificate for the Confluent Schema Registry.", false, true),
	"ssl_key": IdentifierSchema(IdentifierSchemaParams{
		Elem:        "ssl_key",
		Description: "The client key for the Confluent Schema Registry.",
		Required:    false,
		ForceNew:    true,
	}),
	"password": IdentifierSchema(IdentifierSchemaParams{
		Elem:        "password",
		Description: "The password for the Confluent Schema Registry.",
		Required:    false,
		ForceNew:    true,
	}),
	"username": ValueSecretSchema("username", "The username for the Confluent Schema Registry.", false, true),
	"ssh_tunnel": IdentifierSchema(IdentifierSchemaParams{
		Elem:        "ssh_tunnel",
		Description: "The SSH tunnel configuration for the Confluent Schema Registry.",
		Required:    false,
		ForceNew:    true,
	}),
	"aws_privatelink": IdentifierSchema(IdentifierSchemaParams{
		Elem:        "aws_privatelink",
		Description: "The AWS PrivateLink configuration for the Confluent Schema Registry.",
		Required:    false,
		ForceNew:    true,
	}),
	"validate":       ValidateConnectionSchema(),
	"ownership_role": OwnershipRoleSchema(),
	"region":         RegionSchema(),
}

func ConnectionConfluentSchemaRegistry() *schema.Resource {
	return &schema.Resource{
		Description: "A Confluent Schema Registry connection establishes a link to a Confluent Schema Registry server.",

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

	metaDb, region, err := utils.GetDBClientFromMeta(meta, d)
	if err != nil {
		return diag.FromErr(err)
	}
	o := materialize.MaterializeObject{ObjectType: "CONNECTION", Name: connectionName, SchemaName: schemaName, DatabaseName: databaseName}
	b := materialize.NewConnectionConfluentSchemaRegistryBuilder(metaDb, o)

	if v, ok := d.GetOk("url"); ok {
		b.ConfluentSchemaRegistryUrl(v.(string))
	}

	if v, ok := d.GetOk("ssl_certificate_authority"); ok {
		ssl_ca := materialize.GetValueSecretStruct(v)
		b.ConfluentSchemaRegistrySSLCa(ssl_ca)
	}

	if v, ok := d.GetOk("ssl_certificate"); ok {
		ssl_cert := materialize.GetValueSecretStruct(v)
		b.ConfluentSchemaRegistrySSLCert(ssl_cert)
	}

	if v, ok := d.GetOk("ssl_key"); ok {
		key := materialize.GetIdentifierSchemaStruct(v)
		b.ConfluentSchemaRegistrySSLKey(key)
	}

	if v, ok := d.GetOk("username"); ok {
		user := materialize.GetValueSecretStruct(v)
		b.ConfluentSchemaRegistryUsername(user)
	}

	if v, ok := d.GetOk("password"); ok {
		pass := materialize.GetIdentifierSchemaStruct(v)
		b.ConfluentSchemaRegistryPassword(pass)
	}

	if v, ok := d.GetOk("ssh_tunnel"); ok {
		conn := materialize.GetIdentifierSchemaStruct(v)
		b.ConfluentSchemaRegistrySSHTunnel(conn)
	}

	if v, ok := d.GetOk("aws_privatelink"); ok {
		conn := materialize.GetIdentifierSchemaStruct(v)
		b.ConfluentSchemaRegistryAWSPrivateLink(conn)
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
		ownership := materialize.NewOwnershipBuilder(metaDb, o)

		if err := ownership.Alter(v.(string)); err != nil {
			log.Printf("[DEBUG] resource failed ownership, dropping object: %s", o.Name)
			b.Drop()
			return diag.FromErr(err)
		}
	}

	// object comment
	if v, ok := d.GetOk("comment"); ok {
		comment := materialize.NewCommentBuilder(metaDb, o)

		if err := comment.Object(v.(string)); err != nil {
			log.Printf("[DEBUG] resource failed comment, dropping object: %s", o.Name)
			b.Drop()
			return diag.FromErr(err)
		}
	}

	// set id
	i, err := materialize.ConnectionId(metaDb, o)
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(utils.TransformIdWithRegion(string(region), i))

	return connectionRead(ctx, d, meta)
}
