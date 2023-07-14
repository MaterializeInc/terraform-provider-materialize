package resources

import (
	"context"
	"log"

	"github.com/MaterializeInc/terraform-provider-materialize/pkg/materialize"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/jmoiron/sqlx"
)

var connectionKafkaSchema = map[string]*schema.Schema{
	"name":               NameSchema("connection", true, false),
	"schema_name":        SchemaNameSchema("connection", false),
	"database_name":      DatabaseNameSchema("connection", false),
	"qualified_sql_name": QualifiedNameSchema("connection"),
	"kafka_broker": {
		Description: "The Kafka brokers configuration.",
		Type:        schema.TypeList,
		Required:    true,
		MinItems:    1,
		ForceNew:    true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"broker": {
					Description: "The Kafka broker, in the form of `host:port`.",
					Type:        schema.TypeString,
					Required:    true,
				},
				"target_group_port": {
					Description: "The port of the target group associated with the Kafka broker.",
					Type:        schema.TypeInt,
					Optional:    true,
				},
				"availability_zone": {
					Description: "The availability zone of the Kafka broker.",
					Type:        schema.TypeString,
					Optional:    true,
				},
				"privatelink_connection": IdentifierSchema("privatelink_connection", "The AWS PrivateLink connection name in Materialize.", false),
			},
		},
	},
	"progress_topic": {
		Description: "The name of a topic that Kafka sinks can use to track internal consistency metadata.",
		Type:        schema.TypeString,
		Optional:    true,
		ForceNew:    true,
	},
	"ssl_certificate_authority": ValueSecretSchema("ssl_certificate_authority", "The CA certificate for the Kafka broker.", false),
	"ssl_certificate":           ValueSecretSchema("ssl_certificate", "The client certificate for the Kafka broker.", false),
	"ssl_key":                   IdentifierSchema("ssl_key", "The client key for the Kafka broker.", false),
	"sasl_mechanisms": {
		Description:  "The SASL mechanism for the Kafka broker.",
		Type:         schema.TypeString,
		Optional:     true,
		ValidateFunc: validation.StringInSlice(saslMechanisms, true),
		RequiredWith: []string{"sasl_username", "sasl_password"},
		ForceNew:     true,
	},
	"sasl_username":  ValueSecretSchema("sasl_username", "The SASL username for the Kafka broker.", false),
	"sasl_password":  IdentifierSchema("sasl_password", "The SASL password for the Kafka broker.", false),
	"ssh_tunnel":     IdentifierSchema("ssh_tunnel", "The SSH tunnel configuration for the Kafka broker.", false),
	"validate":       ValidateConnection(),
	"ownership_role": OwnershipRole(),
}

func ConnectionKafka() *schema.Resource {
	return &schema.Resource{
		Description: "A Kafka connection establishes a link to a Kafka cluster.",

		CreateContext: connectionKafkaCreate,
		ReadContext:   connectionRead,
		UpdateContext: connectionUpdate,
		DeleteContext: connectionDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: connectionKafkaSchema,
	}
}

func connectionKafkaCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	connectionName := d.Get("name").(string)
	schemaName := d.Get("schema_name").(string)
	databaseName := d.Get("database_name").(string)

	o := materialize.ObjectSchemaStruct{Name: connectionName, SchemaName: schemaName, DatabaseName: databaseName}
	b := materialize.NewConnectionKafkaBuilder(meta.(*sqlx.DB), o)

	if v, ok := d.GetOk("kafka_broker"); ok {
		brokers := materialize.GetKafkaBrokersStruct(databaseName, schemaName, v)
		b.KafkaBrokers(brokers)
	}

	if v, ok := d.GetOk("progress_topic"); ok {
		b.KafkaProgressTopic(v.(string))
	}

	if v, ok := d.GetOk("ssl_certificate_authority"); ok {
		ssl_ca := materialize.GetValueSecretStruct(databaseName, schemaName, v)
		b.KafkaSSLCa(ssl_ca)
	}

	if v, ok := d.GetOk("ssl_certificate"); ok {
		ssl_cert := materialize.GetValueSecretStruct(databaseName, schemaName, v)
		b.KafkaSSLCert(ssl_cert)
	}

	if v, ok := d.GetOk("ssl_key"); ok {
		key := materialize.GetIdentifierSchemaStruct(databaseName, schemaName, v)
		b.KafkaSSLKey(key)
	}

	if v, ok := d.GetOk("sasl_mechanisms"); ok {
		b.KafkaSASLMechanisms(v.(string))
	}

	if v, ok := d.GetOk("sasl_username"); ok {
		sasl_username := materialize.GetValueSecretStruct(databaseName, schemaName, v)
		b.KafkaSASLUsername(sasl_username)
	}

	if v, ok := d.GetOk("sasl_password"); ok {
		pass := materialize.GetIdentifierSchemaStruct(databaseName, schemaName, v)
		b.KafkaSASLPassword(pass)
	}

	if v, ok := d.GetOk("ssh_tunnel"); ok {
		conn := materialize.GetIdentifierSchemaStruct(databaseName, schemaName, v)
		b.KafkaSSHTunnel(conn)
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
