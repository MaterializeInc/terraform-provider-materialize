package resources

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/jmoiron/sqlx"
)

var connectionSchema = map[string]*schema.Schema{
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
		Description:  "The type of the connection.",
		Type:         schema.TypeString,
		Required:     true,
		ValidateFunc: validation.StringInSlice(connectionTypes, true),
	},
	"ssh_host": {
		Description:  "The host of the SSH tunnel.",
		Type:         schema.TypeString,
		Optional:     true,
		RequiredWith: []string{"ssh_user", "ssh_port"},
	},
	"ssh_user": {
		Description:  "The user of the SSH tunnel.",
		Type:         schema.TypeString,
		Optional:     true,
		RequiredWith: []string{"ssh_host", "ssh_port"},
	},
	"ssh_port": {
		Description:  "The port of the SSH tunnel.",
		Type:         schema.TypeInt,
		Optional:     true,
		RequiredWith: []string{"ssh_host", "ssh_user"},
	},
	"aws_privatelink_service_name": {
		Description:   "The name of the AWS PrivateLink service.",
		Type:          schema.TypeString,
		Optional:      true,
		ForceNew:      true,
		ConflictsWith: []string{"ssh_host", "ssh_user", "ssh_port"},
		RequiredWith:  []string{"aws_privatelink_availability_zones"},
	},
	"aws_privatelink_availability_zones": {
		Description:   "The availability zones of the AWS PrivateLink service.",
		Type:          schema.TypeList,
		Elem:          &schema.Schema{Type: schema.TypeString},
		Optional:      true,
		ForceNew:      true,
		ConflictsWith: []string{"ssh_host", "ssh_user", "ssh_port"},
		RequiredWith:  []string{"aws_privatelink_service_name"},
	},
	"postgres_database": {
		Description: "The target Postgres database.",
		Type:        schema.TypeString,
		Optional:    true,
		RequiredWith: []string{
			"postgres_host",
			"postgres_port",
			"postgres_user",
			"postgres_password",
		},
	},
	"postgres_host": {
		Description: "The Postgres database hostname.",
		Type:        schema.TypeString,
		Optional:    true,
		RequiredWith: []string{
			"postgres_database",
			"postgres_port",
			"postgres_user",
			"postgres_password",
		},
	},
	"postgres_port": {
		Description: "The Postgres database port.",
		Type:        schema.TypeInt,
		Optional:    true,
		RequiredWith: []string{
			"postgres_database",
			"postgres_host",
			"postgres_user",
			"postgres_password",
		},
	},
	"postgres_user": {
		Description: "The Postgres database username.",
		Type:        schema.TypeString,
		Optional:    true,
		RequiredWith: []string{
			"postgres_database",
			"postgres_host",
			"postgres_port",
			"postgres_password",
		},
	},
	"postgres_password": {
		Description: "The Postgres database password.",
		Type:        schema.TypeString,
		Optional:    true,
		RequiredWith: []string{
			"postgres_database",
			"postgres_host",
			"postgres_port",
			"postgres_user",
		},
	},
	"postgres_ssh_tunnel": {
		Description: "The SSH tunnel configuration for the Postgres database.",
		Type:        schema.TypeString,
		Optional:    true,
	},
	"postgres_ssl_ca": {
		Description: "The CA certificate for the Postgres database.",
		Type:        schema.TypeString,
		Optional:    true,
	},
	"postgres_ssl_cert": {
		Description: "The client certificate for the Postgres database.",
		Type:        schema.TypeString,
		Optional:    true,
	},
	"postgres_ssl_key": {
		Description: "The client key for the Postgres database.",
		Type:        schema.TypeString,
		Optional:    true,
	},
	"postgres_ssl_mode": {
		Description: "The SSL mode for the Postgres database.",
		Type:        schema.TypeString,
		Optional:    true,
	},
	"postgres_aws_privatelink": {
		Description: "The AWS PrivateLink configuration for the Postgres database.",
		Type:        schema.TypeString,
		Optional:    true,
	},
	"kafka_brokers": {
		Description: "The Kafka brokers configuration.",
		Type:        schema.TypeList,
		Optional:    true,
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
	"kafka_progress_topic": {
		Description: "The name of a topic that Kafka sinks can use to track internal consistency metadata.",
		Type:        schema.TypeString,
		Optional:    true,
	},
	"kafka_ssl_ca": {
		Description: "The CA certificate for the Kafka broker.",
		Type:        schema.TypeString,
		Optional:    true,
	},
	"kafka_ssl_cert": {
		Description: "The client certificate for the Kafka broker.",
		Type:        schema.TypeString,
		Optional:    true,
	},
	"kafka_ssl_key": {
		Description: "The client key for the Kafka broker.",
		Type:        schema.TypeString,
		Optional:    true,
	},
	"kafka_sasl_mechanisms": {
		Description:  "The SASL mechanism for the Kafka broker.",
		Type:         schema.TypeString,
		Optional:     true,
		ValidateFunc: validation.StringInSlice(saslMechanisms, true),
		RequiredWith: []string{"kafka_sasl_username", "kafka_sasl_password"},
	},
	"kafka_sasl_username": {
		Description:  "The SASL username for the Kafka broker.",
		Type:         schema.TypeString,
		Optional:     true,
		RequiredWith: []string{"kafka_sasl_password", "kafka_sasl_mechanisms"},
	},
	"kafka_sasl_password": {
		Description:  "The SASL password for the Kafka broker.",
		Type:         schema.TypeString,
		Optional:     true,
		RequiredWith: []string{"kafka_sasl_username", "kafka_sasl_mechanisms"},
	},
	"kafka_ssh_tunnel": {
		Description: "The SSH tunnel configuration for the Kafka broker.",
		Type:        schema.TypeString,
		Optional:    true,
	},
	"confluent_schema_registry_url": {
		Description: "The URL of the Confluent Schema Registry.",
		Type:        schema.TypeString,
		Optional:    true,
	},
	"confluent_schema_registry_ssl_ca": {
		Description: "The CA certificate for the Confluent Schema Registry.",
		Type:        schema.TypeString,
		Optional:    true,
	},
	"confluent_schema_registry_ssl_cert": {
		Description: "The client certificate for the Confluent Schema Registry.",
		Type:        schema.TypeString,
		Optional:    true,
	},
	"confluent_schema_registry_ssl_key": {
		Description: "The client key for the Confluent Schema Registry.",
		Type:        schema.TypeString,
		Optional:    true,
	},
	"confluent_schema_registry_password": {
		Description:  "The password for the Confluent Schema Registry.",
		Type:         schema.TypeString,
		Optional:     true,
		RequiredWith: []string{"confluent_schema_registry_username"},
	},
	"confluent_schema_registry_username": {
		Description:  "The username for the Confluent Schema Registry.",
		Type:         schema.TypeString,
		Optional:     true,
		RequiredWith: []string{"confluent_schema_registry_password"},
	},
	"confluent_schema_registry_ssh_tunnel": {
		Description: "The SSH tunnel configuration for the Confluent Schema Registry.",
		Type:        schema.TypeString,
		Optional:    true,
	},
	"confluent_schema_registry_aws_privatelink": {
		Description: "The AWS PrivateLink configuration for the Confluent Schema Registry.",
		Type:        schema.TypeString,
		Optional:    true,
	},
}

func Connection() *schema.Resource {
	return &schema.Resource{
		Description: "The connection resource allows you to manage connections in Materialize.",

		CreateContext: connectionCreate,
		ReadContext:   connectionRead,
		UpdateContext: connectionUpdate,
		DeleteContext: connectionDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: connectionSchema,
	}
}

type ConnectionBuilder struct {
	connectionName                        string
	schemaName                            string
	databaseName                          string
	connectionType                        string
	sshHost                               string
	sshUser                               string
	sshPort                               int
	privateLinkServiceName                string
	privateLinkAvailabilityZones          []string
	postgresDatabase                      string
	postgresHost                          string
	postgresPort                          int
	postgresUser                          string
	postgresPassword                      string
	postgresSSHTunnel                     string
	postgresSSLCa                         string
	postgresSSLCert                       string
	postgresSSLKey                        string
	postgresSSLMode                       string
	postgresAWSPrivateLink                string
	kafkaBrokers                          []map[string]interface{}
	kafkaProgressTopic                    string
	kafkaSSLCa                            string
	kafkaSSLCert                          string
	kafkaSSLKey                           string
	kafkaSASLMechanisms                   string
	kafkaSASLUsername                     string
	kafkaSASLPassword                     string
	kafkaSSHTunnel                        string
	confluentSchemaRegistryUrl            string
	confluentSchemaRegistrySSLCa          string
	confluentSchemaRegistrySSLCert        string
	confluentSchemaRegistrySSLKey         string
	confluentSchemaRegistryUsername       string
	confluentSchemaRegistryPassword       string
	confluentSchemaRegistrySSHTunnel      string
	confluentSchemaRegistryAWSPrivateLink string
}

func newConnectionBuilder(connectionName, schemaName, databaseName string) *ConnectionBuilder {
	return &ConnectionBuilder{
		connectionName: connectionName,
		schemaName:     schemaName,
		databaseName:   databaseName,
	}
}

func (b *ConnectionBuilder) ConnectionName(connectionName string) *ConnectionBuilder {
	b.connectionName = connectionName
	return b
}

func (b *ConnectionBuilder) SchemaName(schemaName string) *ConnectionBuilder {
	b.schemaName = schemaName
	return b
}

func (b *ConnectionBuilder) ConnectionType(connectionType string) *ConnectionBuilder {
	b.connectionType = connectionType
	return b
}

func (b *ConnectionBuilder) SSHHost(sshHost string) *ConnectionBuilder {
	b.sshHost = sshHost
	return b
}

func (b *ConnectionBuilder) SSHUser(sshUser string) *ConnectionBuilder {
	b.sshUser = sshUser
	return b
}

func (b *ConnectionBuilder) SSHPort(sshPort int) *ConnectionBuilder {
	b.sshPort = sshPort
	return b
}

func (b *ConnectionBuilder) PrivateLinkServiceName(privateLinkServiceName string) *ConnectionBuilder {
	b.privateLinkServiceName = privateLinkServiceName
	return b
}

func (b *ConnectionBuilder) PrivateLinkAvailabilityZones(privateLinkAvailabilityZones []string) *ConnectionBuilder {
	b.privateLinkAvailabilityZones = privateLinkAvailabilityZones
	return b
}

func (b *ConnectionBuilder) PostgresDatabase(postgresDatabase string) *ConnectionBuilder {
	b.postgresDatabase = postgresDatabase
	return b
}

func (b *ConnectionBuilder) PostgresHost(postgresHost string) *ConnectionBuilder {
	b.postgresHost = postgresHost
	return b
}

func (b *ConnectionBuilder) PostgresPort(postgresPort int) *ConnectionBuilder {
	b.postgresPort = postgresPort
	return b
}

func (b *ConnectionBuilder) PostgresUser(postgresUser string) *ConnectionBuilder {
	b.postgresUser = postgresUser
	return b
}

func (b *ConnectionBuilder) PostgresPassword(postgresPassword string) *ConnectionBuilder {
	b.postgresPassword = postgresPassword
	return b
}

func (b *ConnectionBuilder) PostgresSSHTunnel(postgresSSHTunnel string) *ConnectionBuilder {
	b.postgresSSHTunnel = postgresSSHTunnel
	return b
}

func (b *ConnectionBuilder) PostgresSSLCa(postgresSSLCa string) *ConnectionBuilder {
	b.postgresSSLCa = postgresSSLCa
	return b
}

func (b *ConnectionBuilder) PostgresSSLCert(postgresSSLCert string) *ConnectionBuilder {
	b.postgresSSLCert = postgresSSLCert
	return b
}

func (b *ConnectionBuilder) PostgresSSLKey(postgresSSLKey string) *ConnectionBuilder {
	b.postgresSSLKey = postgresSSLKey
	return b
}

func (b *ConnectionBuilder) PostgresSSLMode(postgresSSLMode string) *ConnectionBuilder {
	b.postgresSSLMode = postgresSSLMode
	return b
}

func (b *ConnectionBuilder) PostgresAWSPrivateLink(postgresAWSPrivateLink string) *ConnectionBuilder {
	b.postgresAWSPrivateLink = postgresAWSPrivateLink
	return b
}

func (b *ConnectionBuilder) KafkaBrokers(kafkaBrokers []map[string]interface{}) *ConnectionBuilder {
	b.kafkaBrokers = kafkaBrokers
	return b
}

func (b *ConnectionBuilder) KafkaProgressTopic(kafkaProgressTopic string) *ConnectionBuilder {
	b.kafkaProgressTopic = kafkaProgressTopic
	return b
}

func (b *ConnectionBuilder) KafkaSSLCa(kafkaSSLCa string) *ConnectionBuilder {
	b.kafkaSSLCa = kafkaSSLCa
	return b
}

func (b *ConnectionBuilder) KafkaSSLCert(kafkaSSLCert string) *ConnectionBuilder {
	b.kafkaSSLCert = kafkaSSLCert
	return b
}

func (b *ConnectionBuilder) KafkaSSLKey(kafkaSSLKey string) *ConnectionBuilder {
	b.kafkaSSLKey = kafkaSSLKey
	return b
}

func (b *ConnectionBuilder) KafkaSASLMechanisms(kafkaSASLMechanisms string) *ConnectionBuilder {
	b.kafkaSASLMechanisms = kafkaSASLMechanisms
	return b
}

func (b *ConnectionBuilder) KafkaSASLUsername(kafkaSASLUsername string) *ConnectionBuilder {
	b.kafkaSASLUsername = kafkaSASLUsername
	return b
}

func (b *ConnectionBuilder) KafkaSASLPassword(kafkaSASLPassword string) *ConnectionBuilder {
	b.kafkaSASLPassword = kafkaSASLPassword
	return b
}

func (b *ConnectionBuilder) KafkaSSHTunnel(kafkaSSHTunnel string) *ConnectionBuilder {
	b.kafkaSSHTunnel = kafkaSSHTunnel
	return b
}

func (b *ConnectionBuilder) ConfluentSchemaRegistryUrl(confluentSchemaRegistryUrl string) *ConnectionBuilder {
	b.confluentSchemaRegistryUrl = confluentSchemaRegistryUrl
	return b
}

func (b *ConnectionBuilder) ConfluentSchemaRegistryUsername(confluentSchemaRegistryUsername string) *ConnectionBuilder {
	b.confluentSchemaRegistryUsername = confluentSchemaRegistryUsername
	return b
}

func (b *ConnectionBuilder) ConfluentSchemaRegistryPassword(confluentSchemaRegistryPassword string) *ConnectionBuilder {
	b.confluentSchemaRegistryPassword = confluentSchemaRegistryPassword
	return b
}

func (b *ConnectionBuilder) ConfluentSchemaRegistrySSLCa(confluentSchemaRegistrySSLCa string) *ConnectionBuilder {
	b.confluentSchemaRegistrySSLCa = confluentSchemaRegistrySSLCa
	return b
}

func (b *ConnectionBuilder) ConfluentSchemaRegistrySSLCert(confluentSchemaRegistrySSLCert string) *ConnectionBuilder {
	b.confluentSchemaRegistrySSLCert = confluentSchemaRegistrySSLCert
	return b
}

func (b *ConnectionBuilder) ConfluentSchemaRegistrySSLKey(confluentSchemaRegistrySSLKey string) *ConnectionBuilder {
	b.confluentSchemaRegistrySSLKey = confluentSchemaRegistrySSLKey
	return b
}

func (b *ConnectionBuilder) ConfluentSchemaRegistrySSHTunnel(confluentSchemaRegistrySSHTunnel string) *ConnectionBuilder {
	b.confluentSchemaRegistrySSHTunnel = confluentSchemaRegistrySSHTunnel
	return b
}

func (b *ConnectionBuilder) ConfluentSchemaRegistryAWSPrivateLink(confluentSchemaRegistryAWSPrivateLink string) *ConnectionBuilder {
	b.confluentSchemaRegistryAWSPrivateLink = confluentSchemaRegistryAWSPrivateLink
	return b
}

func (b *ConnectionBuilder) Create() string {
	q := strings.Builder{}
	q.WriteString(fmt.Sprintf(`CREATE CONNECTION %s.%s.%s`, b.databaseName, b.schemaName, b.connectionName))

	q.WriteString(fmt.Sprintf(` TO %s (`, b.connectionType))

	if b.connectionType == "SSH TUNNEL" {
		q.WriteString(fmt.Sprintf(`HOST '%s', `, b.sshHost))
		q.WriteString(fmt.Sprintf(`USER '%s', `, b.sshUser))
		q.WriteString(fmt.Sprintf(`PORT %d`, b.sshPort))
	}

	if b.connectionType == "AWS PRIVATELINK" {
		q.WriteString(fmt.Sprintf(`SERVICE NAME '%s',`, b.privateLinkServiceName))
		q.WriteString(fmt.Sprintf(`AVAILABILITY ZONES ('%s')`, strings.Join(b.privateLinkAvailabilityZones, "', '")))
	}

	if b.connectionType == "POSTGRES" {
		q.WriteString(fmt.Sprintf(`HOST '%s',`, b.postgresHost))
		q.WriteString(fmt.Sprintf(`PORT %d,`, b.postgresPort))
		q.WriteString(fmt.Sprintf(`USER '%s',`, b.postgresUser))
		q.WriteString(fmt.Sprintf(`PASSWORD SECRET %s`, b.postgresPassword))
		if b.postgresSSLMode != "" {
			q.WriteString(fmt.Sprintf(`, SSL MODE '%s'`, b.postgresSSLMode))
		}
		if b.postgresSSHTunnel != "" {
			q.WriteString(fmt.Sprintf(`, SSH TUNNEL '%s'`, b.postgresSSHTunnel))
		}
		if b.postgresSSLCa != "" {
			q.WriteString(fmt.Sprintf(`, SSL CERTIFICATE AUTHORITY SECRET %s`, b.postgresSSLCa))
		}
		if b.postgresSSLCert != "" {
			q.WriteString(fmt.Sprintf(`, SSL CERTIFICATE SECRET %s`, b.postgresSSLCert))
		}
		if b.postgresSSLKey != "" {
			q.WriteString(fmt.Sprintf(`, SSL KEY SECRET %s`, b.postgresSSLKey))
		}
		if b.postgresAWSPrivateLink != "" {
			q.WriteString(fmt.Sprintf(`, AWS PRIVATELINK %s`, b.postgresAWSPrivateLink))
		}

		q.WriteString(fmt.Sprintf(`, DATABASE '%s'`, b.postgresDatabase))
	}

	if b.connectionType == "KAFKA" {
		if len(b.kafkaBrokers) != 0 {
			if b.kafkaSSHTunnel != "" {
				q.WriteString(`BROKERS (`)
				for i, broker := range b.kafkaBrokers {
					q.WriteString(fmt.Sprintf(`'%s' USING SSH TUNNEL %s`, broker["broker"], b.kafkaSSHTunnel))
					if i < len(b.kafkaBrokers)-1 {
						q.WriteString(`,`)
					}
				}
				q.WriteString(`)`)
			} else {
				q.WriteString(fmt.Sprintf(`BROKERS (`))
				for i, broker := range b.kafkaBrokers {
					if broker["target_group_port"] != nil && broker["availability_zone"] != nil && broker["privatelink_connection"] != nil {
						q.WriteString(fmt.Sprintf(`'%s' USING AWS PRIVATELINK %s (PORT %d, AVAILABILITY ZONE '%s')`, broker["broker"], broker["privatelink_connection"], broker["target_group_port"], broker["availability_zone"]))
						if i < len(b.kafkaBrokers)-1 {
							q.WriteString(`, `)
						}
					} else {
						q.WriteString(fmt.Sprintf(`'%s'`, broker["broker"]))
						if i < len(b.kafkaBrokers)-1 {
							q.WriteString(`, `)
						}
					}
				}
				q.WriteString(`)`)
			}
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

	}

	if b.connectionType == "CONFLUENT SCHEMA REGISTRY" {
		q.WriteString(fmt.Sprintf(`URL '%s'`, b.confluentSchemaRegistryUrl))
		if b.confluentSchemaRegistryUsername != "" {
			q.WriteString(fmt.Sprintf(`, USERNAME = '%s'`, b.confluentSchemaRegistryUsername))
		}
		if b.confluentSchemaRegistryPassword != "" {
			q.WriteString(fmt.Sprintf(`, PASSWORD = SECRET %s`, b.confluentSchemaRegistryPassword))
		}
		if b.confluentSchemaRegistrySSLCa != "" {
			q.WriteString(fmt.Sprintf(`, SSL CERTIFICATE AUTHORITY = %s`, b.confluentSchemaRegistrySSLCa))
		}
		if b.confluentSchemaRegistrySSLCert != "" {
			q.WriteString(fmt.Sprintf(`, SSL CERTIFICATE = %s`, b.confluentSchemaRegistrySSLCert))
		}
		if b.confluentSchemaRegistrySSLKey != "" {
			q.WriteString(fmt.Sprintf(`, SSL KEY = %s`, b.confluentSchemaRegistrySSLKey))
		}
		if b.confluentSchemaRegistryAWSPrivateLink != "" {
			q.WriteString(fmt.Sprintf(`, AWS PRIVATELINK %s`, b.confluentSchemaRegistryAWSPrivateLink))
		}
		if b.confluentSchemaRegistrySSHTunnel != "" {
			q.WriteString(fmt.Sprintf(`, SSH TUNNEL %s`, b.confluentSchemaRegistrySSHTunnel))
		}
	}

	q.WriteString(`);`)
	return q.String()
}

func (b *ConnectionBuilder) ReadId() string {
	return fmt.Sprintf(`
		SELECT mz_connections.id
		FROM mz_connections
		JOIN mz_schemas
			ON mz_connections.schema_id = mz_schemas.id
		JOIN mz_databases
			ON mz_schemas.database_id = mz_databases.id
		WHERE mz_connections.name = '%s'
		AND mz_schemas.name = '%s'
		AND mz_databases.name = '%s';
	`, b.connectionName, b.schemaName, b.databaseName)
}

func (b *ConnectionBuilder) Rename(newConnectionName string) string {
	return fmt.Sprintf(`ALTER CONNECTION %s.%s.%s RENAME TO %s.%s.%s;`, b.databaseName, b.schemaName, b.connectionName, b.databaseName, b.schemaName, newConnectionName)
}

func (b *ConnectionBuilder) Drop() string {
	return fmt.Sprintf(`DROP CONNECTION %s.%s.%s;`, b.databaseName, b.schemaName, b.connectionName)
}

func readConnectionParams(id string) string {
	return fmt.Sprintf(`
		SELECT
			mz_connections.name,
			mz_schemas.name,
			mz_databases.name,
			mz_connections.type
		FROM mz_connections
		JOIN mz_schemas
			ON mz_connections.schema_id = mz_schemas.id
		JOIN mz_databases
			ON mz_schemas.database_id = mz_databases.id
		WHERE mz_connections.id = '%s';`, id)
}

//lint:ignore U1000 Ignore unused function temporarily for debugging
type _connection struct {
	name            sql.NullString `db:"name"`
	schema_name     sql.NullString `db:"schema_name"`
	database_name   sql.NullString `db:"database_name"`
	connection_type sql.NullString `db:"connection_type"`
}

func connectionRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*sqlx.DB)
	i := d.Id()
	q := readConnectionParams(i)

	readResource(conn, d, i, q, _connection{}, "connection")
	setQualifiedName(d)
	return nil
}

func connectionCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*sqlx.DB)

	connectionName := d.Get("name").(string)
	schemaName := d.Get("schema_name").(string)
	databaseName := d.Get("database_name").(string)

	builder := newConnectionBuilder(connectionName, schemaName, databaseName)

	if v, ok := d.GetOk("connection_type"); ok {
		builder.ConnectionType(v.(string))
	}

	if v, ok := d.GetOk("ssh_host"); ok {
		builder.SSHHost(v.(string))
	}

	if v, ok := d.GetOk("ssh_user"); ok {
		builder.SSHUser(v.(string))
	}

	if v, ok := d.GetOk("ssh_port"); ok {
		builder.SSHPort(v.(int))
	}

	if v, ok := d.GetOk("aws_privatelink_service_name"); ok {
		builder.PrivateLinkServiceName(v.(string))
	}

	if v, ok := d.GetOk("aws_privatelink_availability_zones"); ok {
		azs := v.([]interface{})
		var azStrings []string
		for _, az := range azs {
			azStrings = append(azStrings, az.(string))
		}
		builder.PrivateLinkAvailabilityZones(azStrings)
	}

	if v, ok := d.GetOk("postgres_host"); ok {
		builder.PostgresHost(v.(string))
	}

	if v, ok := d.GetOk("postgres_port"); ok {
		builder.PostgresPort(v.(int))
	}

	if v, ok := d.GetOk("postgres_user"); ok {
		builder.PostgresUser(v.(string))
	}

	if v, ok := d.GetOk("postgres_password"); ok {
		builder.PostgresPassword(v.(string))
	}

	if v, ok := d.GetOk("postgres_database"); ok {
		builder.PostgresDatabase(v.(string))
	}

	if v, ok := d.GetOk("postgres_ssl_mode"); ok {
		builder.PostgresSSLMode(v.(string))
	}

	if v, ok := d.GetOk("postgres_ssl_ca"); ok {
		builder.PostgresSSLCa(v.(string))
	}

	if v, ok := d.GetOk("postgres_ssl_cert"); ok {
		builder.PostgresSSLCert(v.(string))
	}

	if v, ok := d.GetOk("postgres_ssl_key"); ok {
		builder.PostgresSSLKey(v.(string))
	}

	if v, ok := d.GetOk("postgres_aws_privatelink"); ok {
		builder.PostgresAWSPrivateLink(v.(string))
	}

	if v, ok := d.GetOk("postgres_ssh_tunnel"); ok {
		builder.PostgresSSHTunnel(v.(string))
	}

	if v, ok := d.GetOk("kafka_brokers"); ok {
		brokers := []map[string]interface{}{}
		for _, b := range v.([]interface{}) {
			brokers = append(brokers, b.(map[string]interface{}))
		}
		builder.KafkaBrokers(brokers)
	}

	if v, ok := d.GetOk("kafka_progress_topic"); ok {
		builder.KafkaProgressTopic(v.(string))
	}

	if v, ok := d.GetOk("kafka_ssl_ca"); ok {
		builder.KafkaSSLCa(v.(string))
	}

	if v, ok := d.GetOk("kafka_ssl_cert"); ok {
		builder.KafkaSSLCert(v.(string))
	}

	if v, ok := d.GetOk("kafka_ssl_key"); ok {
		builder.KafkaSSLKey(v.(string))
	}

	if v, ok := d.GetOk("kafka_sasl_mechanisms"); ok {
		builder.KafkaSASLMechanisms(v.(string))
	}

	if v, ok := d.GetOk("kafka_sasl_username"); ok {
		builder.KafkaSASLUsername(v.(string))
	}

	if v, ok := d.GetOk("kafka_sasl_password"); ok {
		builder.KafkaSASLPassword(v.(string))
	}

	if v, ok := d.GetOk("kafka_ssh_tunnel"); ok {
		builder.KafkaSSHTunnel(v.(string))
	}

	if v, ok := d.GetOk("confluent_schema_registry_url"); ok {
		builder.ConfluentSchemaRegistryUrl(v.(string))
	}

	if v, ok := d.GetOk("confluent_schema_registry_ssl_ca"); ok {
		builder.ConfluentSchemaRegistrySSLCa(v.(string))
	}

	if v, ok := d.GetOk("confluent_schema_registry_ssl_cert"); ok {
		builder.ConfluentSchemaRegistrySSLCert(v.(string))
	}

	if v, ok := d.GetOk("confluent_schema_registry_ssl_key"); ok {
		builder.ConfluentSchemaRegistrySSLKey(v.(string))
	}

	if v, ok := d.GetOk("confluent_schema_registry_username"); ok {
		builder.ConfluentSchemaRegistryUsername(v.(string))
	}

	if v, ok := d.GetOk("confluent_schema_registry_password"); ok {
		builder.ConfluentSchemaRegistryPassword(v.(string))
	}

	if v, ok := d.GetOk("confluent_schema_registry_ssh_tunnel"); ok {
		builder.ConfluentSchemaRegistrySSHTunnel(v.(string))
	}

	if v, ok := d.GetOk("confluent_schema_registry_aws_privatelink"); ok {
		builder.ConfluentSchemaRegistryAWSPrivateLink(v.(string))
	}

	qc := builder.Create()
	qr := builder.ReadId()

	createResource(conn, d, qc, qr, "connection")
	return connectionRead(ctx, d, meta)
}

func connectionUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*sqlx.DB)
	connectionName := d.Get("name").(string)
	schemaName := d.Get("schema_name").(string)
	databaseName := d.Get("database_name").(string)

	if d.HasChange("name") {
		newConnectionName := d.Get("name").(string)
		q := newConnectionBuilder(connectionName, schemaName, databaseName).Rename(newConnectionName)
		if err := ExecResource(conn, q); err != nil {
			log.Printf("[ERROR] could not execute query: %s", q)
			return diag.FromErr(err)
		}
	}

	return connectionRead(ctx, d, meta)
}

func connectionDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*sqlx.DB)
	connectionName := d.Get("name").(string)
	schemaName := d.Get("schema_name").(string)
	databaseName := d.Get("database_name").(string)

	builder := newConnectionBuilder(connectionName, schemaName, databaseName)
	q := builder.Drop()

	dropResource(conn, d, q, "connection")
	return nil
}
