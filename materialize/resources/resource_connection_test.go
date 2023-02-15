package resources

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestResourceConnectoinReadId(t *testing.T) {
	r := require.New(t)
	b := newConnectionBuilder("connection", "schema", "database")
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

func TestResourceConnectionCreateSsh(t *testing.T) {
	r := require.New(t)

	b := newConnectionBuilder("ssh_conn", "schema", "database")
	b.ConnectionType("SSH TUNNEL")
	b.SSHHost("localhost")
	b.SSHPort(123)
	b.SSHUser("user")
	r.Equal(`CREATE CONNECTION database.schema.ssh_conn TO SSH TUNNEL (HOST 'localhost', USER 'user', PORT 123);`, b.Create())

}

func TestResourceConnectionCreateAwsPrivateLink(t *testing.T) {
	r := require.New(t)

	b := newConnectionBuilder("privatelink_conn", "schema", "database")
	b.ConnectionType("AWS PRIVATELINK")
	b.PrivateLinkServiceName("com.amazonaws.us-east-1.materialize.example")
	b.PrivateLinkAvailabilityZones([]string{"us-east-1a", "us-east-1b"})
	r.Equal(`CREATE CONNECTION database.schema.privatelink_conn TO AWS PRIVATELINK (SERVICE NAME 'com.amazonaws.us-east-1.materialize.example',AVAILABILITY ZONES ('us-east-1a', 'us-east-1b'));`, b.Create())
}

func TestResourceConnectionCreateKafka(t *testing.T) {
	r := require.New(t)
	b := newConnectionBuilder("kafka_conn", "schema", "database")
	b.ConnectionType("KAFKA")
	b.KafkaBroker("localhost:9092")
	b.KafkaProgressTopic("topic")
	b.KafkaSASLMechanisms("PLAIN")
	b.KafkaSASLUsername("user")
	b.KafkaSASLPassword("password")
	r.Equal(`CREATE CONNECTION database.schema.kafka_conn TO KAFKA (BROKER 'localhost:9092', PROGRESS TOPIC 'topic', SASL MECHANISMS 'PLAIN', SASL USERNAME 'user', SASL PASSWORD SECRET password);`, b.Create())
}

func TestResourceConnectionCreateKafkaSsh(t *testing.T) {
	r := require.New(t)
	b := newConnectionBuilder("kafka_conn", "schema", "database")
	b.ConnectionType("KAFKA")
	b.KafkaBroker("localhost:9092")
	b.KafkaProgressTopic("topic")
	b.KafkaSASLMechanisms("PLAIN")
	b.KafkaSASLUsername("user")
	b.KafkaSASLPassword("password")
	b.KafkaSSHTunnel("ssh_conn")
	r.Equal(`CREATE CONNECTION database.schema.kafka_conn TO KAFKA (BROKER 'localhost:9092' USING SSH TUNNEL ssh_conn, PROGRESS TOPIC 'topic', SASL MECHANISMS 'PLAIN', SASL USERNAME 'user', SASL PASSWORD SECRET password);`, b.Create())
}

func TestResourceConnectionCreateKafkaBrokers(t *testing.T) {
	r := require.New(t)
	b := newConnectionBuilder("kafka_conn", "schema", "database")
	b.ConnectionType("KAFKA")
	b.KafkaBrokers([]string{"localhost:9092", "localhost:9093"})
	b.KafkaProgressTopic("topic")
	b.KafkaSASLMechanisms("PLAIN")
	b.KafkaSASLUsername("user")
	b.KafkaSASLPassword("password")
	r.Equal(`CREATE CONNECTION database.schema.kafka_conn TO KAFKA (BROKERS ('localhost:9092', 'localhost:9093'), PROGRESS TOPIC 'topic', SASL MECHANISMS 'PLAIN', SASL USERNAME 'user', SASL PASSWORD SECRET password);`, b.Create())
}

func TestResourceConnectionCreateKafkaBrokersSsh(t *testing.T) {
	r := require.New(t)
	b := newConnectionBuilder("kafka_conn", "schema", "database")
	b.ConnectionType("KAFKA")
	b.KafkaBrokers([]string{"localhost:9092", "localhost:9093"})
	b.KafkaProgressTopic("topic")
	b.KafkaSASLMechanisms("PLAIN")
	b.KafkaSASLUsername("user")
	b.KafkaSASLPassword("password")
	b.KafkaSSHTunnel("ssh_conn")
	r.Equal(`CREATE CONNECTION database.schema.kafka_conn TO KAFKA (BROKERS ('localhost:9092' USING SSH TUNNEL ssh_conn,'localhost:9093' USING SSH TUNNEL ssh_conn), PROGRESS TOPIC 'topic', SASL MECHANISMS 'PLAIN', SASL USERNAME 'user', SASL PASSWORD SECRET password);`, b.Create())
}

func TestResourceConnectionCreateKafkaSsl(t *testing.T) {
	r := require.New(t)
	b := newConnectionBuilder("kafka_conn", "schema", "database")
	b.ConnectionType("KAFKA")
	b.KafkaBroker("localhost:9092")
	b.KafkaProgressTopic("topic")
	b.KafkaSSLKey("key")
	b.KafkaSSLCert("cert")
	b.KafkaSSLCa("ca")
	r.Equal(`CREATE CONNECTION database.schema.kafka_conn TO KAFKA (BROKER 'localhost:9092', PROGRESS TOPIC 'topic', SSL CERTIFICATE AUTHORITY SECRET ca, SSL CERTIFICATE SECRET cert, SSL KEY SECRET key);`, b.Create())
}

func TestResourceConnectionCreateConfluentSchemaRegistry(t *testing.T) {
	r := require.New(t)
	b := newConnectionBuilder("csr_conn", "schema", "database")
	b.ConnectionType("CONFLUENT SCHEMA REGISTRY")
	b.ConfluentSchemaRegistryUrl("http://localhost:8081")
	b.ConfluentSchemaRegistryUsername("user")
	b.ConfluentSchemaRegistryPassword("password")
	r.Equal(`CREATE CONNECTION database.schema.csr_conn TO CONFLUENT SCHEMA REGISTRY (URL 'http://localhost:8081', USERNAME = 'user', PASSWORD = SECRET password);`, b.Create())

}

func TestResourceConnectionRename(t *testing.T) {
	r := require.New(t)
	b := newConnectionBuilder("connection", "schema", "database")
	r.Equal(`ALTER CONNECTION database.schema.connection RENAME TO database.schema.new_connection;`, b.Rename("new_connection"))
}

func TestResourceConnectionDrop(t *testing.T) {
	r := require.New(t)
	b := newConnectionBuilder("connection", "schema", "database")
	r.Equal(`DROP CONNECTION database.schema.connection;`, b.Drop())
}

func TestResourceConnectionReadParams(t *testing.T) {
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
