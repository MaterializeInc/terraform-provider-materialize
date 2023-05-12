package resources

import (
	"context"

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
	},
	"ssl_certificate_authority": ValueSecretSchema("ssl_certificate_authority", "The CA certificate for the Kafka broker.", false, true),
	"ssl_certificate":           ValueSecretSchema("ssl_certificate", "The client certificate for the Kafka broker.", false, true),
	"ssl_key":                   IdentifierSchema("ssl_key", "The client key for the Kafka broker.", false),
	"sasl_mechanisms": {
		Description:  "The SASL mechanism for the Kafka broker.",
		Type:         schema.TypeString,
		Optional:     true,
		ValidateFunc: validation.StringInSlice(saslMechanisms, true),
		RequiredWith: []string{"sasl_username", "sasl_password"},
	},
	"sasl_username": ValueSecretSchema("sasl_username", "The SASL username for the Kafka broker.", false, true),
	"sasl_password": IdentifierSchema("sasl_password", "The SASL password for the Kafka broker.", false),
	"ssh_tunnel":    IdentifierSchema("ssh_tunnel", "The SSH tunnel configuration for the Kafka broker.", false),
}

func ConnectionKafka() *schema.Resource {
	return &schema.Resource{
		Description: "The connection resource allows you to manage connections in Materialize.",

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

	builder := materialize.NewConnectionKafkaBuilder(meta.(*sqlx.DB), connectionName, schemaName, databaseName)

	if v, ok := d.GetOk("kafka_broker"); ok {
		brokers := materialize.GetKafkaBrokersStruct(databaseName, schemaName, v)
		builder.KafkaBrokers(brokers)
	}

	if v, ok := d.GetOk("progress_topic"); ok {
		builder.KafkaProgressTopic(v.(string))
	}

	if v, ok := d.GetOk("ssl_certificate_authority"); ok {
		ssl_ca := materialize.GetValueSecretStruct(databaseName, schemaName, v)
		builder.KafkaSSLCa(ssl_ca)
	}

	if v, ok := d.GetOk("ssl_certificate"); ok {
		ssl_cert := materialize.GetValueSecretStruct(databaseName, schemaName, v)
		builder.KafkaSSLCert(ssl_cert)
	}

	if v, ok := d.GetOk("ssl_key"); ok {
		key := materialize.GetIdentifierSchemaStruct(databaseName, schemaName, v)
		builder.KafkaSSLKey(key)
	}

	if v, ok := d.GetOk("sasl_mechanisms"); ok {
		builder.KafkaSASLMechanisms(v.(string))
	}

	if v, ok := d.GetOk("sasl_username"); ok {
		sasl_username := materialize.GetValueSecretStruct(databaseName, schemaName, v)
		builder.KafkaSASLUsername(sasl_username)
	}

	if v, ok := d.GetOk("sasl_password"); ok {
		pass := materialize.GetIdentifierSchemaStruct(databaseName, schemaName, v)
		builder.KafkaSASLPassword(pass)
	}

	if v, ok := d.GetOk("ssh_tunnel"); ok {
		conn := materialize.GetIdentifierSchemaStruct(databaseName, schemaName, v)
		builder.KafkaSSHTunnel(conn)
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
