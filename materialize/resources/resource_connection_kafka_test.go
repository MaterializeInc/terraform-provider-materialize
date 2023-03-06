package resources

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestResourceConnectionKafkaReadId(t *testing.T) {
	r := require.New(t)
	b := newConnectionKafkaBuilder("connection", "schema", "database")
	r.Equal(`
		SELECT mz_connections.id
		FROM mz_connections
		JOIN mz_schemas
			ON mz_connections.schema_id = mz_schemas.id
		JOIN mz_databases
			ON mz_schemas.database_id = mz_databases.id
		WHERE mz_connections.name = 'connection'
		AND mz_schemas.name = 'schema'
		AND mz_databases.name = 'database';
	`, b.ReadId())
}

func TestResourceConnectionKafkaRename(t *testing.T) {
	r := require.New(t)
	b := newConnectionKafkaBuilder("connection", "schema", "database")
	r.Equal(`ALTER CONNECTION database.schema.connection RENAME TO database.schema.new_connection;`, b.Rename("new_connection"))
}

func TestResourceConnectionKafkaDrop(t *testing.T) {
	r := require.New(t)
	b := newConnectionKafkaBuilder("connection", "schema", "database")
	r.Equal(`DROP CONNECTION database.schema.connection;`, b.Drop())
}

func TestResourceConnectionKafkaReadParams(t *testing.T) {
	r := require.New(t)
	b := readConnectionParams("u1")
	r.Equal(`
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
		WHERE mz_connections.id = 'u1';`, b)
}

func TestResourceConnectionCreateKafka(t *testing.T) {
	r := require.New(t)
	b := newConnectionKafkaBuilder("kafka_conn", "schema", "database")
	b.KafkaBrokers([]KafkaBroker{
		{
			Broker: "localhost:9092",
		},
	})
	b.KafkaProgressTopic("topic")
	b.KafkaSASLMechanisms("PLAIN")
	b.KafkaSASLUsername("user")
	b.KafkaSASLPassword("password")
	r.Equal(`CREATE CONNECTION database.schema.kafka_conn TO KAFKA (BROKERS ('localhost:9092'), PROGRESS TOPIC 'topic', SASL MECHANISMS = 'PLAIN', SASL USERNAME = 'user', SASL PASSWORD = SECRET password);`, b.Create())
}

func TestResourceConnectionCreateKafkaMultipleBrokers(t *testing.T) {
	r := require.New(t)
	b := newConnectionKafkaBuilder("kafka_conn", "schema", "database")
	b.KafkaBrokers([]KafkaBroker{
		{
			Broker: "localhost:9092",
		},
		{
			Broker: "localhost:9093",
		},
	})
	b.KafkaProgressTopic("topic")
	b.KafkaSASLMechanisms("PLAIN")
	b.KafkaSASLUsername("user")
	b.KafkaSASLPassword("password")
	r.Equal(`CREATE CONNECTION database.schema.kafka_conn TO KAFKA (BROKERS ('localhost:9092', 'localhost:9093'), PROGRESS TOPIC 'topic', SASL MECHANISMS = 'PLAIN', SASL USERNAME = 'user', SASL PASSWORD = SECRET password);`, b.Create())
}

func TestResourceConnectionCreateKafkaSsh(t *testing.T) {
	r := require.New(t)
	b := newConnectionKafkaBuilder("kafka_conn", "schema", "database")
	b.KafkaBrokers([]KafkaBroker{
		{
			Broker: "localhost:9092",
		},
	})
	b.KafkaProgressTopic("topic")
	b.KafkaSASLMechanisms("PLAIN")
	b.KafkaSASLUsername("user")
	b.KafkaSASLPassword("password")
	b.KafkaSSHTunnel("ssh_conn")
	r.Equal(`CREATE CONNECTION database.schema.kafka_conn TO KAFKA (BROKERS ('localhost:9092' USING SSH TUNNEL ssh_conn), PROGRESS TOPIC 'topic', SASL MECHANISMS = 'PLAIN', SASL USERNAME = 'user', SASL PASSWORD = SECRET password);`, b.Create())
}

func TestResourceConnectionCreateKafkaBrokers(t *testing.T) {
	r := require.New(t)
	b := newConnectionKafkaBuilder("kafka_conn", "schema", "database")
	b.KafkaBrokers([]KafkaBroker{
		{
			Broker: "localhost:9092",
		},
		{
			Broker: "localhost:9093",
		},
	})
	b.KafkaProgressTopic("topic")
	b.KafkaSASLMechanisms("PLAIN")
	b.KafkaSASLUsername("user")
	b.KafkaSASLPassword("password")
	r.Equal(`CREATE CONNECTION database.schema.kafka_conn TO KAFKA (BROKERS ('localhost:9092', 'localhost:9093'), PROGRESS TOPIC 'topic', SASL MECHANISMS = 'PLAIN', SASL USERNAME = 'user', SASL PASSWORD = SECRET password);`, b.Create())
}

func TestResourceConnectionCreateKafkaBrokersSsh(t *testing.T) {
	r := require.New(t)
	b := newConnectionKafkaBuilder("kafka_conn", "schema", "database")
	b.KafkaBrokers([]KafkaBroker{
		{
			Broker: "localhost:9092",
		},
		{
			Broker: "localhost:9093",
		},
	})
	b.KafkaProgressTopic("topic")
	b.KafkaSASLMechanisms("PLAIN")
	b.KafkaSASLUsername("user")
	b.KafkaSASLPassword("password")
	b.KafkaSSHTunnel("ssh_conn")
	r.Equal(`CREATE CONNECTION database.schema.kafka_conn TO KAFKA (BROKERS ('localhost:9092' USING SSH TUNNEL ssh_conn,'localhost:9093' USING SSH TUNNEL ssh_conn), PROGRESS TOPIC 'topic', SASL MECHANISMS = 'PLAIN', SASL USERNAME = 'user', SASL PASSWORD = SECRET password);`, b.Create())
}

func TestResourceConnectionCreateKafkaSsl(t *testing.T) {
	r := require.New(t)
	b := newConnectionKafkaBuilder("kafka_conn", "schema", "database")
	b.KafkaBrokers([]KafkaBroker{
		{
			Broker: "localhost:9092",
		},
	})
	b.KafkaProgressTopic("topic")
	b.KafkaSSLKey("key")
	b.KafkaSSLCert("cert")
	b.KafkaSSLCa("ca")
	r.Equal(`CREATE CONNECTION database.schema.kafka_conn TO KAFKA (BROKERS ('localhost:9092'), PROGRESS TOPIC 'topic', SSL CERTIFICATE AUTHORITY = SECRET ca, SSL CERTIFICATE = SECRET cert, SSL KEY = SECRET key);`, b.Create())
}

func TestResourceConnectionKafkaAwsPrivatelink(t *testing.T) {
	r := require.New(t)
	b := newConnectionKafkaBuilder("kafka_conn", "schema", "database")
	b.KafkaBrokers([]KafkaBroker{
		{
			Broker:                "b-1.hostname-1:9096",
			TargetGroupPort:       9001,
			AvailabilityZone:      "use1-az1",
			PrivateLinkConnection: "privatelink_conn",
		},
		{
			Broker:                "b-1.hostname-1:9097",
			TargetGroupPort:       9002,
			AvailabilityZone:      "use1-az2",
			PrivateLinkConnection: "privatelink_conn",
		},
	})
	b.KafkaSASLMechanisms("PLAIN")
	b.KafkaSASLUsername("user")
	b.KafkaSASLPassword("password")
	r.Equal(`CREATE CONNECTION database.schema.kafka_conn TO KAFKA (BROKERS ('b-1.hostname-1:9096' USING AWS PRIVATELINK privatelink_conn (PORT 9001, AVAILABILITY ZONE 'use1-az1'), 'b-1.hostname-1:9097' USING AWS PRIVATELINK privatelink_conn (PORT 9002, AVAILABILITY ZONE 'use1-az2')), SASL MECHANISMS = 'PLAIN', SASL USERNAME = 'user', SASL PASSWORD = SECRET password);`, b.Create())
}
