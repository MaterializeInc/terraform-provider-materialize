package resources

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/jmoiron/sqlx"
)

var connectionKafkaSchema = map[string]*schema.Schema{
	"name": {
		Description: "The name of the connection.",
		Type:        schema.TypeString,
		Required:    true,
	},
	"schema_name": {
		Description: "The identifier for the connection schema.",
		Type:        schema.TypeString,
		Optional:    true,
		Default:     "public",
	},
	"database_name": {
		Description: "The identifier for the connection database.",
		Type:        schema.TypeString,
		Optional:    true,
		Default:     "materialize",
	},
	"qualified_name": {
		Description: "The fully qualified name of the connection.",
		Type:        schema.TypeString,
		Computed:    true,
	},
	"connection_type": {
		Description: "The type of connection.",
		Type:        schema.TypeString,
		Computed:    true,
	},
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
				"privatelink_connection": IdentifierSchema("privatelink_connection", "The AWS PrivateLink connection name in Materialize.", false, true),
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
	"ssl_key":                   IdentifierSchema("ssl_key", "The client key for the Kafka broker.", false, true),
	"sasl_mechanisms": {
		Description:  "The SASL mechanism for the Kafka broker.",
		Type:         schema.TypeString,
		Optional:     true,
		ValidateFunc: validation.StringInSlice(saslMechanisms, true),
		RequiredWith: []string{"sasl_username", "sasl_password"},
	},
	"sasl_username": ValueSecretSchema("sasl_username", "The SASL username for the Kafka broker.", false, true),
	"sasl_password": IdentifierSchema("sasl_password", "The SASL password for the Kafka broker.", false, true),
	"ssh_tunnel":    IdentifierSchema("ssh_tunnel", "The SSH tunnel configuration for the Kafka broker.", false, true),
}

func ConnectionKafka() *schema.Resource {
	return &schema.Resource{
		Description: "The connection resource allows you to manage connections in Materialize.",

		CreateContext: connectionKafkaCreate,
		ReadContext:   ConnectionRead,
		UpdateContext: connectionKafkaUpdate,
		DeleteContext: connectionKafkaDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: connectionKafkaSchema,
	}
}

type KafkaBroker struct {
	Broker                string
	TargetGroupPort       int
	AvailabilityZone      string
	PrivateLinkConnection IdentifierSchemaStruct
}

type ConnectionKafkaBuilder struct {
	connectionName      string
	schemaName          string
	databaseName        string
	kafkaBrokers        []KafkaBroker
	kafkaProgressTopic  string
	kafkaSSLCa          ValueSecretStruct
	kafkaSSLCert        ValueSecretStruct
	kafkaSSLKey         IdentifierSchemaStruct
	kafkaSASLMechanisms string
	kafkaSASLUsername   ValueSecretStruct
	kafkaSASLPassword   IdentifierSchemaStruct
	kafkaSSHTunnel      IdentifierSchemaStruct
}

func newConnectionKafkaBuilder(connectionName, schemaName, databaseName string) *ConnectionKafkaBuilder {
	return &ConnectionKafkaBuilder{
		connectionName: connectionName,
		schemaName:     schemaName,
		databaseName:   databaseName,
	}
}

func (b *ConnectionKafkaBuilder) qualifiedName() string {
	return QualifiedName(b.databaseName, b.schemaName, b.connectionName)
}

func (b *ConnectionKafkaBuilder) KafkaBrokers(kafkaBrokers []KafkaBroker) *ConnectionKafkaBuilder {
	b.kafkaBrokers = kafkaBrokers
	return b
}

func (b *ConnectionKafkaBuilder) KafkaProgressTopic(kafkaProgressTopic string) *ConnectionKafkaBuilder {
	b.kafkaProgressTopic = kafkaProgressTopic
	return b
}

func (b *ConnectionKafkaBuilder) KafkaSSLCa(kafkaSSLCa ValueSecretStruct) *ConnectionKafkaBuilder {
	b.kafkaSSLCa = kafkaSSLCa
	return b
}

func (b *ConnectionKafkaBuilder) KafkaSSLCert(kafkaSSLCert ValueSecretStruct) *ConnectionKafkaBuilder {
	b.kafkaSSLCert = kafkaSSLCert
	return b
}

func (b *ConnectionKafkaBuilder) KafkaSSLKey(kafkaSSLKey IdentifierSchemaStruct) *ConnectionKafkaBuilder {
	b.kafkaSSLKey = kafkaSSLKey
	return b
}

func (b *ConnectionKafkaBuilder) KafkaSASLMechanisms(kafkaSASLMechanisms string) *ConnectionKafkaBuilder {
	b.kafkaSASLMechanisms = kafkaSASLMechanisms
	return b
}

func (b *ConnectionKafkaBuilder) KafkaSASLUsername(kafkaSASLUsername ValueSecretStruct) *ConnectionKafkaBuilder {
	b.kafkaSASLUsername = kafkaSASLUsername
	return b
}

func (b *ConnectionKafkaBuilder) KafkaSASLPassword(kafkaSASLPassword IdentifierSchemaStruct) *ConnectionKafkaBuilder {
	b.kafkaSASLPassword = kafkaSASLPassword
	return b
}

func (b *ConnectionKafkaBuilder) KafkaSSHTunnel(kafkaSSHTunnel IdentifierSchemaStruct) *ConnectionKafkaBuilder {
	b.kafkaSSHTunnel = kafkaSSHTunnel
	return b
}

func (b *ConnectionKafkaBuilder) Create() string {
	q := strings.Builder{}
	q.WriteString(fmt.Sprintf(`CREATE CONNECTION %s TO KAFKA (`, b.qualifiedName()))

	if b.kafkaSSHTunnel.Name != "" {
		q.WriteString(`BROKERS (`)
		for i, broker := range b.kafkaBrokers {
			q.WriteString(fmt.Sprintf(`%s USING SSH TUNNEL %s`, QuoteString(broker.Broker), QualifiedName(b.kafkaSSHTunnel.DatabaseName, b.kafkaSSHTunnel.SchemaName, b.kafkaSSHTunnel.Name)))
			if i < len(b.kafkaBrokers)-1 {
				q.WriteString(`,`)
			}
		}
		q.WriteString(`)`)
	} else {
		q.WriteString(`BROKERS (`)
		for i, broker := range b.kafkaBrokers {
			if broker.TargetGroupPort != 0 && broker.AvailabilityZone != "" && broker.PrivateLinkConnection.Name != "" {
				q.WriteString(fmt.Sprintf(`%s USING AWS PRIVATELINK %s (PORT %d, AVAILABILITY ZONE %s)`, QuoteString(broker.Broker),
					QualifiedName(broker.PrivateLinkConnection.DatabaseName, broker.PrivateLinkConnection.SchemaName, broker.PrivateLinkConnection.Name), broker.TargetGroupPort, QuoteString(broker.AvailabilityZone)))
				if i < len(b.kafkaBrokers)-1 {
					q.WriteString(`, `)
				}
			} else {
				q.WriteString(fmt.Sprintf(`%s`, QuoteString(broker.Broker)))
				if i < len(b.kafkaBrokers)-1 {
					q.WriteString(`, `)
				}
			}
		}
		q.WriteString(`)`)
	}

	if b.kafkaProgressTopic != "" {
		q.WriteString(fmt.Sprintf(`, PROGRESS TOPIC %s`, QuoteString(b.kafkaProgressTopic)))
	}
	if b.kafkaSSLCa.Text != "" {
		q.WriteString(fmt.Sprintf(`, SSL CERTIFICATE AUTHORITY = %s`, QuoteString(b.kafkaSSLCa.Text)))
	}
	if b.kafkaSSLCa.Secret.Name != "" {
		q.WriteString(fmt.Sprintf(`, SSL CERTIFICATE AUTHORITY = SECRET %s`, QualifiedName(b.kafkaSSLCa.Secret.DatabaseName, b.kafkaSSLCa.Secret.SchemaName, b.kafkaSSLCa.Secret.Name)))
	}
	if b.kafkaSSLCert.Text != "" {
		q.WriteString(fmt.Sprintf(`, SSL CERTIFICATE = %s`, QuoteString(b.kafkaSSLCert.Text)))
	}
	if b.kafkaSSLCert.Secret.Name != "" {
		q.WriteString(fmt.Sprintf(`, SSL CERTIFICATE = SECRET %s`, QualifiedName(b.kafkaSSLCert.Secret.DatabaseName, b.kafkaSSLCert.Secret.SchemaName, b.kafkaSSLCert.Secret.Name)))
	}
	if b.kafkaSSLKey.Name != "" {
		q.WriteString(fmt.Sprintf(`, SSL KEY = SECRET %s`, QualifiedName(b.kafkaSSLKey.DatabaseName, b.kafkaSSLKey.SchemaName, b.kafkaSSLKey.Name)))
	}
	if b.kafkaSASLMechanisms != "" {
		q.WriteString(fmt.Sprintf(`, SASL MECHANISMS = %s`, QuoteString(b.kafkaSASLMechanisms)))
	}
	if b.kafkaSASLUsername.Text != "" {
		q.WriteString(fmt.Sprintf(`, SASL USERNAME = %s`, QuoteString(b.kafkaSASLUsername.Text)))
	}
	if b.kafkaSASLUsername.Secret.Name != "" {
		q.WriteString(fmt.Sprintf(`, SASL USERNAME = SECRET %s`, QualifiedName(b.kafkaSASLUsername.Secret.DatabaseName, b.kafkaSASLUsername.Secret.SchemaName, b.kafkaSASLUsername.Secret.Name)))
	}
	if b.kafkaSASLPassword.Name != "" {
		q.WriteString(fmt.Sprintf(`, SASL PASSWORD = SECRET %s`, QualifiedName(b.kafkaSASLPassword.DatabaseName, b.kafkaSASLPassword.SchemaName, b.kafkaSASLPassword.Name)))
	}

	q.WriteString(`);`)
	return q.String()
}

func (b *ConnectionKafkaBuilder) Rename(newConnectionName string) string {
	n := QualifiedName(b.databaseName, b.schemaName, newConnectionName)
	return fmt.Sprintf(`ALTER CONNECTION %s RENAME TO %s;`, b.qualifiedName(), n)
}

func (b *ConnectionKafkaBuilder) Drop() string {
	return fmt.Sprintf(`DROP CONNECTION %s;`, b.qualifiedName())
}

func (b *ConnectionKafkaBuilder) ReadId() string {
	return readConnectionId(b.connectionName, b.schemaName, b.databaseName)
}

func connectionKafkaCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*sqlx.DB)

	connectionName := d.Get("name").(string)
	schemaName := d.Get("schema_name").(string)
	databaseName := d.Get("database_name").(string)

	builder := newConnectionKafkaBuilder(connectionName, schemaName, databaseName)

	if v, ok := d.GetOk("kafka_broker"); ok {
		var brokers []KafkaBroker
		for _, broker := range v.([]interface{}) {
			b := broker.(map[string]interface{})
			privateLinkConn := IdentifierSchemaStruct{}
			if b["private_link_connection"] != nil {
				privateLinkConn = GetIdentifierSchemaStruct(databaseName, schemaName, b["private_link_connection"].([]interface{}))
			}
			brokers = append(brokers, KafkaBroker{
				Broker:                b["broker"].(string),
				TargetGroupPort:       b["target_group_port"].(int),
				AvailabilityZone:      b["availability_zone"].(string),
				PrivateLinkConnection: privateLinkConn,
			})
		}
		builder.KafkaBrokers(brokers)
	}

	if v, ok := d.GetOk("progress_topic"); ok {
		builder.KafkaProgressTopic(v.(string))
	}

	if v, ok := d.GetOk("ssl_certificate_authority"); ok {
		var ssl_ca ValueSecretStruct
		u := v.([]interface{})[0].(map[string]interface{})
		if v, ok := u["text"]; ok {
			ssl_ca.Text = v.(string)
		}
		if v, ok := u["secret"]; ok && len(v.([]interface{})) > 0 {
			ssl_ca.Secret = GetIdentifierSchemaStruct(databaseName, schemaName, v)
		}
		builder.KafkaSSLCa(ssl_ca)
	}

	if v, ok := d.GetOk("ssl_certificate"); ok {
		var ssl_cert ValueSecretStruct
		u := v.([]interface{})[0].(map[string]interface{})
		if v, ok := u["text"]; ok {
			ssl_cert.Text = v.(string)
		}
		if v, ok := u["secret"]; ok && len(v.([]interface{})) > 0 {
			ssl_cert.Secret = GetIdentifierSchemaStruct(databaseName, schemaName, v)
		}
		builder.KafkaSSLCert(ssl_cert)
	}

	if v, ok := d.GetOk("ssl_key"); ok {
		key := GetIdentifierSchemaStruct(databaseName, schemaName, v)
		builder.KafkaSSLKey(key)
	}

	if v, ok := d.GetOk("sasl_mechanisms"); ok {
		builder.KafkaSASLMechanisms(v.(string))
	}

	if v, ok := d.GetOk("sasl_username"); ok {
		var sasl_username ValueSecretStruct
		u := v.([]interface{})[0].(map[string]interface{})
		if v, ok := u["text"]; ok {
			sasl_username.Text = v.(string)
		}
		if v, ok := u["secret"]; ok && len(v.([]interface{})) > 0 {
			sasl_username.Secret = GetIdentifierSchemaStruct(databaseName, schemaName, v)
		}
		builder.KafkaSASLUsername(sasl_username)
	}

	if v, ok := d.GetOk("sasl_password"); ok {
		pass := GetIdentifierSchemaStruct(databaseName, schemaName, v)
		builder.KafkaSASLPassword(pass)
	}

	if v, ok := d.GetOk("ssh_tunnel"); ok {
		conn := GetIdentifierSchemaStruct(databaseName, schemaName, v)
		builder.KafkaSSHTunnel(conn)
	}

	qc := builder.Create()
	qr := builder.ReadId()

	if err := createResource(conn, d, qc, qr, "connection"); err != nil {
		return diag.FromErr(err)
	}
	return ConnectionRead(ctx, d, meta)
}

func connectionKafkaUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*sqlx.DB)
	connectionName := d.Get("name").(string)
	schemaName := d.Get("schema_name").(string)
	databaseName := d.Get("database_name").(string)

	if d.HasChange("name") {
		newConnectionName := d.Get("name").(string)
		q := newConnectionKafkaBuilder(connectionName, schemaName, databaseName).Rename(newConnectionName)
		if err := ExecResource(conn, q); err != nil {
			log.Printf("[ERROR] could not execute query: %s", q)
			return diag.FromErr(err)
		}
	}

	return ConnectionRead(ctx, d, meta)
}

func connectionKafkaDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*sqlx.DB)
	connectionName := d.Get("name").(string)
	schemaName := d.Get("schema_name").(string)
	databaseName := d.Get("database_name").(string)

	builder := newConnectionKafkaBuilder(connectionName, schemaName, databaseName)
	q := builder.Drop()

	if err := dropResource(conn, d, q, "connection"); err != nil {
		return diag.FromErr(err)
	}
	return nil
}
