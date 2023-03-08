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
				"privatelink_connection": {
					Description: "The AWS PrivateLink connection name in Materialize.",
					Type:        schema.TypeString,
					Optional:    true,
				},
			},
		},
	},
	"progress_topic": {
		Description: "The name of a topic that Kafka sinks can use to track internal consistency metadata.",
		Type:        schema.TypeString,
		Optional:    true,
	},
	"ssl_certificate_authority": {
		Description: "The CA certificate for the Kafka broker.",
		Type:        schema.TypeString,
		Optional:    true,
	},
	"ssl_certificate": {
		Description: "The client certificate for the Kafka broker.",
		Type:        schema.TypeString,
		Optional:    true,
	},
	"ssl_key": {
		Description: "The client key for the Kafka broker.",
		Type:        schema.TypeString,
		Optional:    true,
	},
	"sasl_mechanisms": {
		Description:  "The SASL mechanism for the Kafka broker.",
		Type:         schema.TypeString,
		Optional:     true,
		ValidateFunc: validation.StringInSlice(saslMechanisms, true),
		RequiredWith: []string{"sasl_username", "sasl_password"},
	},
	"sasl_username": {
		Description:  "The SASL username for the Kafka broker.",
		Type:         schema.TypeString,
		Optional:     true,
		RequiredWith: []string{"sasl_password", "sasl_mechanisms"},
	},
	"sasl_password": {
		Description:  "The SASL password for the Kafka broker.",
		Type:         schema.TypeString,
		Optional:     true,
		RequiredWith: []string{"sasl_username", "sasl_mechanisms"},
	},
	"ssh_tunnel": {
		Description: "The SSH tunnel configuration for the Kafka broker.",
		Type:        schema.TypeString,
		Optional:    true,
	},
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
	PrivateLinkConnection string
}

type ConnectionKafkaBuilder struct {
	connectionName      string
	schemaName          string
	databaseName        string
	kafkaBrokers        []KafkaBroker
	kafkaProgressTopic  string
	kafkaSSLCa          string
	kafkaSSLCert        string
	kafkaSSLKey         string
	kafkaSASLMechanisms string
	kafkaSASLUsername   string
	kafkaSASLPassword   string
	kafkaSSHTunnel      string
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

func (b *ConnectionKafkaBuilder) KafkaSSLCa(kafkaSSLCa string) *ConnectionKafkaBuilder {
	b.kafkaSSLCa = kafkaSSLCa
	return b
}

func (b *ConnectionKafkaBuilder) KafkaSSLCert(kafkaSSLCert string) *ConnectionKafkaBuilder {
	b.kafkaSSLCert = kafkaSSLCert
	return b
}

func (b *ConnectionKafkaBuilder) KafkaSSLKey(kafkaSSLKey string) *ConnectionKafkaBuilder {
	b.kafkaSSLKey = kafkaSSLKey
	return b
}

func (b *ConnectionKafkaBuilder) KafkaSASLMechanisms(kafkaSASLMechanisms string) *ConnectionKafkaBuilder {
	b.kafkaSASLMechanisms = kafkaSASLMechanisms
	return b
}

func (b *ConnectionKafkaBuilder) KafkaSASLUsername(kafkaSASLUsername string) *ConnectionKafkaBuilder {
	b.kafkaSASLUsername = kafkaSASLUsername
	return b
}

func (b *ConnectionKafkaBuilder) KafkaSASLPassword(kafkaSASLPassword string) *ConnectionKafkaBuilder {
	b.kafkaSASLPassword = kafkaSASLPassword
	return b
}

func (b *ConnectionKafkaBuilder) KafkaSSHTunnel(kafkaSSHTunnel string) *ConnectionKafkaBuilder {
	b.kafkaSSHTunnel = kafkaSSHTunnel
	return b
}

func (b *ConnectionKafkaBuilder) Create() string {
	q := strings.Builder{}
	q.WriteString(fmt.Sprintf(`CREATE CONNECTION %s TO KAFKA (`, b.qualifiedName()))

	if b.kafkaSSHTunnel != "" {
		q.WriteString(`BROKERS (`)
		for i, broker := range b.kafkaBrokers {
			q.WriteString(fmt.Sprintf(`'%s' USING SSH TUNNEL %s`, broker.Broker, b.kafkaSSHTunnel))
			if i < len(b.kafkaBrokers)-1 {
				q.WriteString(`,`)
			}
		}
		q.WriteString(`)`)
	} else {
		q.WriteString(`BROKERS (`)
		for i, broker := range b.kafkaBrokers {
			if broker.TargetGroupPort != 0 && broker.AvailabilityZone != "" && broker.PrivateLinkConnection != "" {
				q.WriteString(fmt.Sprintf(`'%s' USING AWS PRIVATELINK %s (PORT %d, AVAILABILITY ZONE '%s')`, broker.Broker, broker.PrivateLinkConnection, broker.TargetGroupPort, broker.AvailabilityZone))
				if i < len(b.kafkaBrokers)-1 {
					q.WriteString(`, `)
				}
			} else {
				q.WriteString(fmt.Sprintf(`'%s'`, broker.Broker))
				if i < len(b.kafkaBrokers)-1 {
					q.WriteString(`, `)
				}
			}
		}
		q.WriteString(`)`)
	}

	if b.kafkaProgressTopic != "" {
		q.WriteString(fmt.Sprintf(`, PROGRESS TOPIC '%s'`, b.kafkaProgressTopic))
	}
	if b.kafkaSSLCa != "" {
		q.WriteString(fmt.Sprintf(`, SSL CERTIFICATE AUTHORITY = SECRET %s`, b.kafkaSSLCa))
	}
	if b.kafkaSSLCert != "" {
		q.WriteString(fmt.Sprintf(`, SSL CERTIFICATE = SECRET %s`, b.kafkaSSLCert))
	}
	if b.kafkaSSLKey != "" {
		q.WriteString(fmt.Sprintf(`, SSL KEY = SECRET %s`, b.kafkaSSLKey))
	}
	if b.kafkaSASLMechanisms != "" {
		q.WriteString(fmt.Sprintf(`, SASL MECHANISMS = '%s'`, b.kafkaSASLMechanisms))
	}
	if b.kafkaSASLUsername != "" {
		q.WriteString(fmt.Sprintf(`, SASL USERNAME = '%s'`, b.kafkaSASLUsername))
	}
	if b.kafkaSASLPassword != "" {
		q.WriteString(fmt.Sprintf(`, SASL PASSWORD = SECRET %s`, b.kafkaSASLPassword))
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
			brokers = append(brokers, KafkaBroker{
				Broker:                b["broker"].(string),
				TargetGroupPort:       b["target_group_port"].(int),
				AvailabilityZone:      b["availability_zone"].(string),
				PrivateLinkConnection: b["privatelink_connection"].(string),
			})
		}
		builder.KafkaBrokers(brokers)
	}

	if v, ok := d.GetOk("progress_topic"); ok {
		builder.KafkaProgressTopic(v.(string))
	}

	if v, ok := d.GetOk("ssl_certificate_authority"); ok {
		builder.KafkaSSLCa(v.(string))
	}

	if v, ok := d.GetOk("ssl_certificate"); ok {
		builder.KafkaSSLCert(v.(string))
	}

	if v, ok := d.GetOk("ssl_key"); ok {
		builder.KafkaSSLKey(v.(string))
	}

	if v, ok := d.GetOk("sasl_mechanisms"); ok {
		builder.KafkaSASLMechanisms(v.(string))
	}

	if v, ok := d.GetOk("sasl_username"); ok {
		builder.KafkaSASLUsername(v.(string))
	}

	if v, ok := d.GetOk("sasl_password"); ok {
		builder.KafkaSASLPassword(v.(string))
	}

	if v, ok := d.GetOk("ssh_tunnel"); ok {
		builder.KafkaSSHTunnel(v.(string))
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
