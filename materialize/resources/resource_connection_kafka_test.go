package resources

import (
	"context"
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/require"
)

func TestResourceKafkaCreate(t *testing.T) {
	r := require.New(t)

	in := map[string]interface{}{
		"name":                      "conn",
		"schema_name":               "schema",
		"database_name":             "database",
		"service_name":              "service",
		"kafka_broker":              []interface{}{map[string]interface{}{"broker": "b-1.hostname-1:9096", "target_group_port": 9001, "availability_zone": "use1-az1", "privatelink_conn": "privatelink_conn"}},
		"progress_topic":            "topic",
		"ssl_certificate_authority": "key",
		"ssl_certificate":           "cert",
		"ssl_key":                   "key",
		"sasl_mechanisms":           "PLAIN",
		"sasl_username":             "username",
		"sasl_password":             "password",
		"ssh_tunnel":                "tunnel",
	}
	d := schema.TestResourceDataRaw(t, ConnectionKafka().Schema, in)
	r.NotNil(d)

	WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		// Create
		mock.ExpectExec(
			`CREATE CONNECTION database.schema.conn TO KAFKA \(BROKERS \('b-1.hostname-1:9096' USING SSH TUNNEL tunnel\), PROGRESS TOPIC 'topic', SSL CERTIFICATE AUTHORITY = SECRET key, SSL CERTIFICATE = SECRET cert, SSL KEY = SECRET key, SASL MECHANISMS = 'PLAIN', SASL USERNAME = 'username', SASL PASSWORD = SECRET password\);`,
		).WillReturnResult(sqlmock.NewResult(1, 1))

		// Query Id
		ir := mock.NewRows([]string{"id"}).AddRow("u1")
		mock.ExpectQuery(`
			SELECT mz_connections.id
			FROM mz_connections
			JOIN mz_schemas
				ON mz_connections.schema_id = mz_schemas.id
			JOIN mz_databases
				ON mz_schemas.database_id = mz_databases.id
			WHERE mz_connections.name = 'conn'
			AND mz_schemas.name = 'schema'
			AND mz_databases.name = 'database';`).WillReturnRows(ir)

		// Query Params
		ip := sqlmock.NewRows([]string{"name", "schema", "database", "connection_type"}).
			AddRow("conn", "schema", "database", "connection_type")
		mock.ExpectQuery(`
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
			WHERE mz_connections.id = 'u1';`).WillReturnRows(ip)

		if err := connectionKafkaCreate(context.TODO(), d, db); err != nil {
			t.Fatal(err)
		}
	})

}

func TestResourceKafkaDelete(t *testing.T) {
	r := require.New(t)

	in := map[string]interface{}{
		"name":          "conn",
		"schema_name":   "schema",
		"database_name": "database",
	}
	d := schema.TestResourceDataRaw(t, ConnectionKafka().Schema, in)
	r.NotNil(d)

	WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		mock.ExpectExec(`DROP CONNECTION database.schema.conn;`).WillReturnResult(sqlmock.NewResult(1, 1))

		if err := connectionKafkaDelete(context.TODO(), d, db); err != nil {
			t.Fatal(err)
		}
	})

}

func TestConnectionKafkaReadIdQuery(t *testing.T) {
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

func TestConnectionKafkaRenameQuery(t *testing.T) {
	r := require.New(t)
	b := newConnectionKafkaBuilder("connection", "schema", "database")
	r.Equal(`ALTER CONNECTION database.schema.connection RENAME TO database.schema.new_connection;`, b.Rename("new_connection"))
}

func TestConnectionKafkaDropQuery(t *testing.T) {
	r := require.New(t)
	b := newConnectionKafkaBuilder("connection", "schema", "database")
	r.Equal(`DROP CONNECTION database.schema.connection;`, b.Drop())
}

func TestConnectionKafkaReadParamsQuery(t *testing.T) {
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

func TestConnectionCreateKafkaQuery(t *testing.T) {
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

func TestConnectionCreateKafkaMultipleBrokersQuery(t *testing.T) {
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

func TestConnectionCreateKafkaSshQuery(t *testing.T) {
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

func TestConnectionCreateKafkaBrokersQuery(t *testing.T) {
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

func TestConnectionCreateKafkaBrokersSshQuery(t *testing.T) {
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

func TestConnectionCreateKafkaSslQuery(t *testing.T) {
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

func TestConnectionKafkaAwsPrivatelinkQuery(t *testing.T) {
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
