package materialize

import (
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"
)

type KafkaBroker struct {
	Broker                string
	TargetGroupPort       int
	AvailabilityZone      string
	PrivateLinkConnection IdentifierSchemaStruct
	SSHTunnel             IdentifierSchemaStruct
}

func GetKafkaBrokersStruct(v interface{}) []KafkaBroker {
	var brokers []KafkaBroker
	for _, broker := range v.([]interface{}) {
		b := broker.(map[string]interface{})
		privateLinkConn := IdentifierSchemaStruct{}
		if b["privatelink_connection"] != nil && len(b["privatelink_connection"].([]interface{})) > 0 {
			privateLinkConn = GetIdentifierSchemaStruct(b["privatelink_connection"].([]interface{}))
		}
		SshTunnel := IdentifierSchemaStruct{}
		if b["ssh_tunnel"] != nil && len(b["ssh_tunnel"].([]interface{})) > 0 {
			SshTunnel = GetIdentifierSchemaStruct(b["ssh_tunnel"].([]interface{}))
		}
		brokers = append(brokers, KafkaBroker{
			Broker:                b["broker"].(string),
			TargetGroupPort:       b["target_group_port"].(int),
			AvailabilityZone:      b["availability_zone"].(string),
			PrivateLinkConnection: privateLinkConn,
			SSHTunnel:             SshTunnel,
		})
	}
	return brokers
}

type ConnectionKafkaBuilder struct {
	Connection
	kafkaBrokers        []KafkaBroker
	kafkaProgressTopic  string
	kafkaSSLCa          ValueSecretStruct
	kafkaSSLCert        ValueSecretStruct
	kafkaSSLKey         IdentifierSchemaStruct
	kafkaSASLMechanisms string
	kafkaSASLUsername   ValueSecretStruct
	kafkaSASLPassword   IdentifierSchemaStruct
	kafkaSSHTunnel      IdentifierSchemaStruct
	validate            bool
}

func NewConnectionKafkaBuilder(conn *sqlx.DB, obj MaterializeObject) *ConnectionKafkaBuilder {
	b := Builder{conn, BaseConnection}
	return &ConnectionKafkaBuilder{
		Connection: Connection{b, obj.Name, obj.SchemaName, obj.DatabaseName},
	}
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

func (b *ConnectionKafkaBuilder) Validate(validate bool) *ConnectionKafkaBuilder {
	b.validate = validate
	return b
}

func (b *ConnectionKafkaBuilder) Create() error {
	q := strings.Builder{}
	q.WriteString(fmt.Sprintf(`CREATE CONNECTION %s TO KAFKA`, b.QualifiedName()))

	var brokers = []string{}
	for _, broker := range b.kafkaBrokers {
		fb := strings.Builder{}
		fb.WriteString(QuoteString(broker.Broker))

		if broker.SSHTunnel.Name != "" {
			fb.WriteString(fmt.Sprintf(` USING SSH TUNNEL %s`,
				QualifiedName(
					broker.SSHTunnel.DatabaseName,
					broker.SSHTunnel.SchemaName,
					broker.SSHTunnel.Name,
				),
			))
		}

		if broker.PrivateLinkConnection.Name != "" {
			p := strings.Builder{}
			p.WriteString(fmt.Sprintf(` USING AWS PRIVATELINK %s`,
				QualifiedName(
					broker.PrivateLinkConnection.DatabaseName,
					broker.PrivateLinkConnection.SchemaName,
					broker.PrivateLinkConnection.Name,
				),
			))
			fb.WriteString(p.String())

			options := []string{}
			if broker.TargetGroupPort != 0 {
				o := fmt.Sprintf(`PORT %d`, broker.TargetGroupPort)
				options = append(options, o)
			}
			if broker.AvailabilityZone != "" {
				o := fmt.Sprintf(`AVAILABILITY ZONE %s`, QuoteString(broker.AvailabilityZone))
				options = append(options, o)
			}
			if len(options) > 0 {
				fb.WriteString(fmt.Sprintf(` (%s)`, strings.Join(options[:], ", ")))
			}
		}
		brokers = append(brokers, fb.String())
	}
	q.WriteString(fmt.Sprintf(` (BROKERS (%s)`, strings.Join(brokers[:], ", ")))

	if b.kafkaSSHTunnel.Name != "" {
		q.WriteString(fmt.Sprintf(`, SSH TUNNEL %s`,
			QualifiedName(
				b.kafkaSSHTunnel.DatabaseName,
				b.kafkaSSHTunnel.SchemaName,
				b.kafkaSSHTunnel.Name,
			),
		))
	}

	if b.kafkaProgressTopic != "" {
		q.WriteString(fmt.Sprintf(`, PROGRESS TOPIC %s`, QuoteString(b.kafkaProgressTopic)))
	}
	if b.kafkaSSLCa.Text != "" {
		q.WriteString(fmt.Sprintf(`, SSL CERTIFICATE AUTHORITY = %s`, QuoteString(b.kafkaSSLCa.Text)))
	}
	if b.kafkaSSLCa.Secret.Name != "" {
		q.WriteString(fmt.Sprintf(`, SSL CERTIFICATE AUTHORITY = SECRET %s`, b.kafkaSSLCa.Secret.QualifiedName()))
	}
	if b.kafkaSSLCert.Text != "" {
		q.WriteString(fmt.Sprintf(`, SSL CERTIFICATE = %s`, QuoteString(b.kafkaSSLCert.Text)))
	}
	if b.kafkaSSLCert.Secret.Name != "" {
		q.WriteString(fmt.Sprintf(`, SSL CERTIFICATE = SECRET %s`, b.kafkaSSLCert.Secret.QualifiedName()))
	}
	if b.kafkaSSLKey.Name != "" {
		q.WriteString(fmt.Sprintf(`, SSL KEY = SECRET %s`, b.kafkaSSLKey.QualifiedName()))
	}
	if b.kafkaSASLMechanisms != "" {
		q.WriteString(fmt.Sprintf(`, SASL MECHANISMS = %s`, QuoteString(b.kafkaSASLMechanisms)))
	}
	if b.kafkaSASLUsername.Text != "" {
		q.WriteString(fmt.Sprintf(`, SASL USERNAME = %s`, QuoteString(b.kafkaSASLUsername.Text)))
	}
	if b.kafkaSASLUsername.Secret.Name != "" {
		q.WriteString(fmt.Sprintf(`, SASL USERNAME = SECRET %s`, b.kafkaSASLUsername.Secret.QualifiedName()))
	}
	if b.kafkaSASLPassword.Name != "" {
		q.WriteString(fmt.Sprintf(`, SASL PASSWORD = SECRET %s`, b.kafkaSASLPassword.QualifiedName()))
	}

	q.WriteString(`)`)

	if !b.validate {
		q.WriteString(` WITH (VALIDATE = false)`)
	}

	return b.ddl.exec(q.String())
}
