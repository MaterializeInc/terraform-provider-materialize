package resources

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestResourceConnectionCreateSsh(t *testing.T) {
	r := require.New(t)

	b := newConnectionBuilder("ssh_conn", "schema")
	b.ConnectionType("SSH TUNNEL")
	b.SSHHost("localhost")
	b.SSHPort(123)
	b.SSHUser("user")
	r.Equal(`CREATE CONNECTION schema.ssh_conn TO SSH TUNNEL (HOST 'localhost', USER 'user', PORT 123);`, b.Create())

}

func TestResourceConnectionCreateAwsPrivateLink(t *testing.T) {
	r := require.New(t)

	b := newConnectionBuilder("privatelink_conn", "schema")
	b.ConnectionType("AWS PRIVATELINK")
	b.PrivateLinkServiceName("com.amazonaws.us-east-1.materialize.example")
	b.PrivateLinkAvailabilityZones([]string{"us-east-1a", "us-east-1b"})
	r.Equal(`CREATE CONNECTION schema.privatelink_conn TO AWS PRIVATELINK (SERVICE NAME 'com.amazonaws.us-east-1.materialize.example',AVAILABILITY ZONES ('us-east-1a', 'us-east-1b'));`, b.Create())

}

func TestResourceConnectionCreatePostgres(t *testing.T) {
	r := require.New(t)
	b := newSourceBuilder("source", "schema")
	b.Size("xsmall")
	b.ConnectionType("POSTGRES")
	b.PostgresConnection("pg_connection")
	b.Publication("mz_source")
	r.Equal(`CREATE SOURCE schema.source FROM POSTGRES CONNECTION pg_connection (PUBLICATION 'mz_source') FOR ALL TABLES WITH (SIZE = 'xsmall');`, b.Create())
}

func TestResourceConnectionCreatePostgresTables(t *testing.T) {
	r := require.New(t)
	b := newSourceBuilder("source", "schema")
	b.Size("xsmall")
	b.ConnectionType("POSTGRES")
	b.PostgresConnection("pg_connection")
	b.Publication("mz_source")
	b.Tables(map[string]string{
		"schema1.table_1": "s1_table_1",
		"schema2_table_1": "s2_table_1",
	})
	r.Equal(`CREATE SOURCE schema.source FROM POSTGRES CONNECTION pg_connection (PUBLICATION 'mz_source') FOR TABLES (schema1.table_1 AS s1_table_1, schema2_table_1 AS s2_table_1) WITH (SIZE = 'xsmall');`, b.Create())
}

func TestResourceConnectionCreateKafka(t *testing.T) {
	r := require.New(t)
	b := newConnectionBuilder("kafka_conn", "schema")
	b.ConnectionType("KAFKA")
	b.KafkaBroker("localhost:9092")
	b.KafkaProgressTopic("topic")
	b.KafkaSASLMechanisms("PLAIN")
	b.KafkaSASLUsername("user")
	b.KafkaSASLPassword("password")
	r.Equal(`CREATE CONNECTION schema.kafka_conn TO KAFKA (BROKER 'localhost:9092', PROGRESS TOPIC 'topic', SASL MECHANISMS 'PLAIN', SASL USERNAME 'user', SASL PASSWORD SECRET password);`, b.Create())
}

func TestResourceConnectionCreateKafkaSsh(t *testing.T) {
	r := require.New(t)
	b := newConnectionBuilder("kafka_conn", "schema")
	b.ConnectionType("KAFKA")
	b.KafkaBroker("localhost:9092")
	b.KafkaProgressTopic("topic")
	b.KafkaSASLMechanisms("PLAIN")
	b.KafkaSASLUsername("user")
	b.KafkaSASLPassword("password")
	b.KafkaSSHTunnel("ssh_conn")
	r.Equal(`CREATE CONNECTION schema.kafka_conn TO KAFKA (BROKER 'localhost:9092' USING SSH TUNNEL ssh_conn, PROGRESS TOPIC 'topic', SASL MECHANISMS 'PLAIN', SASL USERNAME 'user', SASL PASSWORD SECRET password);`, b.Create())
}

func TestResourceConnectionCreateKafkaBrokers(t *testing.T) {
	r := require.New(t)
	b := newConnectionBuilder("kafka_conn", "schema")
	b.ConnectionType("KAFKA")
	b.KafkaBrokers([]string{"localhost:9092", "localhost:9093"})
	b.KafkaProgressTopic("topic")
	b.KafkaSASLMechanisms("PLAIN")
	b.KafkaSASLUsername("user")
	b.KafkaSASLPassword("password")
	r.Equal(`CREATE CONNECTION schema.kafka_conn TO KAFKA (BROKERS ('localhost:9092', 'localhost:9093'), PROGRESS TOPIC 'topic', SASL MECHANISMS 'PLAIN', SASL USERNAME 'user', SASL PASSWORD SECRET password);`, b.Create())
}

func TestResourceConnectionCreateKafkaBrokersSsh(t *testing.T) {
	r := require.New(t)
	b := newConnectionBuilder("kafka_conn", "schema")
	b.ConnectionType("KAFKA")
	b.KafkaBrokers([]string{"localhost:9092", "localhost:9093"})
	b.KafkaProgressTopic("topic")
	b.KafkaSASLMechanisms("PLAIN")
	b.KafkaSASLUsername("user")
	b.KafkaSASLPassword("password")
	b.KafkaSSHTunnel("ssh_conn")
	r.Equal(`CREATE CONNECTION schema.kafka_conn TO KAFKA (BROKERS ('localhost:9092' USING SSH TUNNEL ssh_conn,'localhost:9093' USING SSH TUNNEL ssh_conn), PROGRESS TOPIC 'topic', SASL MECHANISMS 'PLAIN', SASL USERNAME 'user', SASL PASSWORD SECRET password);`, b.Create())
}

func TestResourceConnectionCreateKafkaSsl(t *testing.T) {
	r := require.New(t)
	b := newConnectionBuilder("kafka_conn", "schema")
	b.ConnectionType("KAFKA")
	b.KafkaBroker("localhost:9092")
	b.KafkaProgressTopic("topic")
	b.KafkaSSLKey("key")
	b.KafkaSSLCert("cert")
	b.KafkaSSLCa("ca")
	r.Equal(`CREATE CONNECTION schema.kafka_conn TO KAFKA (BROKER 'localhost:9092', PROGRESS TOPIC 'topic', SSL CERTIFICATE AUTHORITY SECRET ca, SSL CERTIFICATE SECRET cert, SSL KEY SECRET key);`, b.Create())
}

func TestResourceConnectionCreateConfluentSchemaRegistry(t *testing.T) {
	r := require.New(t)
	b := newConnectionBuilder("csr_conn", "schema")
	b.ConnectionType("CONFLUENT SCHEMA REGISTRY")
	b.ConfluentSchemaRegistryUrl("http://localhost:8081")
	b.ConfluentSchemaRegistryUsername("user")
	b.ConfluentSchemaRegistryPassword("password")
	r.Equal(`CREATE CONNECTION schema.csr_conn TO CONFLUENT SCHEMA REGISTRY (URL 'http://localhost:8081', USERNAME = 'user', PASSWORD = SECRET password);`, b.Create())

}
