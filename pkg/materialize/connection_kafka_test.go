package materialize

import (
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/MaterializeInc/terraform-provider-materialize/pkg/testhelpers"
	"github.com/jmoiron/sqlx"
)

var connKafka = MaterializeObject{Name: "kafka_conn", SchemaName: "schema", DatabaseName: "database"}

func TestConnectionKafkaCreate(t *testing.T) {
	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		mock.ExpectExec(
			`CREATE CONNECTION "database"."schema"."kafka_conn" TO KAFKA \(BROKERS \('localhost:9092'\), SECURITY PROTOCOL = 'PLAIN', PROGRESS TOPIC 'topic', SASL MECHANISMS = 'PLAIN', SASL USERNAME = 'user', SASL PASSWORD = SECRET "database"."schema"."password"\);`,
		).WillReturnResult(sqlmock.NewResult(1, 1))

		b := NewConnectionKafkaBuilder(db, connKafka)
		b.KafkaBrokers([]KafkaBroker{
			{
				Broker: "localhost:9092",
			},
		})
		b.KafkaProgressTopic("topic")
		b.KafkaSecurityProtocol("PLAIN")
		b.KafkaSASLMechanisms("PLAIN")
		b.KafkaSASLUsername(ValueSecretStruct{Text: "user"})
		b.KafkaSASLPassword(IdentifierSchemaStruct{Name: "password", DatabaseName: "database", SchemaName: "schema"})
		b.Validate(true)

		if err := b.Create(); err != nil {
			t.Fatal(err)
		}
	})
}

func TestConnectionKafkaDefaultSshTunnelCreate(t *testing.T) {
	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		mock.ExpectExec(
			`CREATE CONNECTION "database"."schema"."kafka_conn" TO KAFKA \(BROKERS \('localhost:9092'\), SSH TUNNEL "database"."schema"."ssh_conn"\, PROGRESS TOPIC 'topic', SASL MECHANISMS = 'PLAIN', SASL USERNAME = 'user', SASL PASSWORD = SECRET "database"."schema"."password"\);`,
		).WillReturnResult(sqlmock.NewResult(1, 1))

		b := NewConnectionKafkaBuilder(db, connKafka)
		b.KafkaBrokers([]KafkaBroker{
			{
				Broker: "localhost:9092",
			},
		})
		b.KafkaProgressTopic("topic")
		b.KafkaSASLMechanisms("PLAIN")
		b.KafkaSSHTunnel(IdentifierSchemaStruct{Name: "ssh_conn", DatabaseName: "database", SchemaName: "schema"})
		b.KafkaSASLUsername(ValueSecretStruct{Text: "user"})
		b.KafkaSASLPassword(IdentifierSchemaStruct{Name: "password", DatabaseName: "database", SchemaName: "schema"})
		b.Validate(true)

		if err := b.Create(); err != nil {
			t.Fatal(err)
		}
	})
}

func TestConnectionKafkaMultipleBrokersCreate(t *testing.T) {
	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		mock.ExpectExec(
			`CREATE CONNECTION "database"."schema"."kafka_conn" TO KAFKA \(BROKERS \('localhost:9092', 'localhost:9093'\), PROGRESS TOPIC 'topic', SASL MECHANISMS = 'PLAIN', SASL USERNAME = 'user', SASL PASSWORD = SECRET "database"."schema"."password"\);`,
		).WillReturnResult(sqlmock.NewResult(1, 1))

		b := NewConnectionKafkaBuilder(db, connKafka)
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
		b.KafkaSASLUsername(ValueSecretStruct{Text: "user"})
		b.KafkaSASLPassword(IdentifierSchemaStruct{Name: "password", DatabaseName: "database", SchemaName: "schema"})
		b.Validate(true)

		if err := b.Create(); err != nil {
			t.Fatal(err)
		}
	})
}

func TestConnectionKafkaSshCreate(t *testing.T) {
	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		mock.ExpectExec(
			`CREATE CONNECTION "database"."schema"."kafka_conn" TO KAFKA \(BROKERS \('localhost:9092' USING SSH TUNNEL "database"."schema"."ssh_conn"\), PROGRESS TOPIC 'topic', SASL MECHANISMS = 'PLAIN', SASL USERNAME = 'user', SASL PASSWORD = SECRET "database"."schema"."password"\);`,
		).WillReturnResult(sqlmock.NewResult(1, 1))

		b := NewConnectionKafkaBuilder(db, connKafka)
		b.KafkaBrokers([]KafkaBroker{
			{
				Broker:    "localhost:9092",
				SSHTunnel: IdentifierSchemaStruct{Name: "ssh_conn", DatabaseName: "database", SchemaName: "schema"},
			},
		})
		b.KafkaProgressTopic("topic")
		b.KafkaSASLMechanisms("PLAIN")
		b.KafkaSASLUsername(ValueSecretStruct{Text: "user"})
		b.KafkaSASLPassword(IdentifierSchemaStruct{Name: "password", DatabaseName: "database", SchemaName: "schema"})
		b.Validate(true)

		if err := b.Create(); err != nil {
			t.Fatal(err)
		}
	})
}

func TestConnectionKafkaBrokersCreate(t *testing.T) {
	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		mock.ExpectExec(
			`CREATE CONNECTION "database"."schema"."kafka_conn" TO KAFKA \(BROKERS \('localhost:9092', 'localhost:9093'\), PROGRESS TOPIC 'topic', SASL MECHANISMS = 'PLAIN', SASL USERNAME = 'user', SASL PASSWORD = SECRET "database"."schema"."password"\);`,
		).WillReturnResult(sqlmock.NewResult(1, 1))

		b := NewConnectionKafkaBuilder(db, connKafka)
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
		b.KafkaSASLUsername(ValueSecretStruct{Text: "user"})
		b.KafkaSASLPassword(IdentifierSchemaStruct{Name: "password", DatabaseName: "database", SchemaName: "schema"})
		b.Validate(true)

		if err := b.Create(); err != nil {
			t.Fatal(err)
		}
	})
}

func TestConnectionKafkaBrokersSshCreate(t *testing.T) {
	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		mock.ExpectExec(
			`CREATE CONNECTION "database"."schema"."kafka_conn" TO KAFKA \(BROKERS \('localhost:9092' USING SSH TUNNEL "database"."schema"."ssh_conn", 'localhost:9093' USING SSH TUNNEL "database"."schema"."ssh_conn"\), PROGRESS TOPIC 'topic', SASL MECHANISMS = 'PLAIN', SASL USERNAME = 'user', SASL PASSWORD = SECRET "database"."schema"."password"\);`,
		).WillReturnResult(sqlmock.NewResult(1, 1))

		b := NewConnectionKafkaBuilder(db, connKafka)
		b.KafkaBrokers([]KafkaBroker{
			{
				Broker:    "localhost:9092",
				SSHTunnel: IdentifierSchemaStruct{Name: "ssh_conn", DatabaseName: "database", SchemaName: "schema"},
			},
			{
				Broker:    "localhost:9093",
				SSHTunnel: IdentifierSchemaStruct{Name: "ssh_conn", DatabaseName: "database", SchemaName: "schema"},
			},
		})
		b.KafkaProgressTopic("topic")
		b.KafkaSASLMechanisms("PLAIN")
		b.KafkaSASLUsername(ValueSecretStruct{Text: "user"})
		b.KafkaSASLPassword(IdentifierSchemaStruct{Name: "password", DatabaseName: "database", SchemaName: "schema"})
		b.Validate(true)

		if err := b.Create(); err != nil {
			t.Fatal(err)
		}
	})
}

func TestConnectionKafkaSslCreate(t *testing.T) {
	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		mock.ExpectExec(
			`CREATE CONNECTION "database"."schema"."kafka_conn" TO KAFKA \(BROKERS \('localhost:9092'\), SECURITY PROTOCOL = 'SSL', PROGRESS TOPIC 'topic', SSL CERTIFICATE AUTHORITY = SECRET "database"."schema"."ca", SSL CERTIFICATE = SECRET "database"."schema"."cert", SSL KEY = SECRET "database"."schema"."key"\);`,
		).WillReturnResult(sqlmock.NewResult(1, 1))

		b := NewConnectionKafkaBuilder(db, connKafka)
		b.KafkaBrokers([]KafkaBroker{
			{
				Broker: "localhost:9092",
			},
		})
		b.KafkaProgressTopic("topic")
		b.KafkaSecurityProtocol("SSL")
		b.KafkaSSLKey(IdentifierSchemaStruct{SchemaName: "schema", Name: "key", DatabaseName: "database"})
		b.KafkaSSLCert(ValueSecretStruct{Secret: IdentifierSchemaStruct{SchemaName: "schema", Name: "cert", DatabaseName: "database"}})
		b.KafkaSSLCa(ValueSecretStruct{Secret: IdentifierSchemaStruct{SchemaName: "schema", Name: "ca", DatabaseName: "database"}})
		b.Validate(true)

		if err := b.Create(); err != nil {
			t.Fatal(err)
		}
	})
}

func TestConnectionKafkaAwsPrivatelinkCreate(t *testing.T) {
	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		mock.ExpectExec(
			`CREATE CONNECTION "database"."schema"."kafka_conn" TO KAFKA \(BROKERS \('b-1.hostname-1:9096' USING AWS PRIVATELINK "database"."schema"."privatelink_conn" \(PORT 9001, AVAILABILITY ZONE 'use1-az1'\), 'b-1.hostname-1:9097' USING AWS PRIVATELINK "database"."schema"."privatelink_conn" \(PORT 9002, AVAILABILITY ZONE 'use1-az2'\)\), SASL MECHANISMS = 'PLAIN', SASL USERNAME = 'user', SASL PASSWORD = SECRET "database"."schema"."password"\);`,
		).WillReturnResult(sqlmock.NewResult(1, 1))

		b := NewConnectionKafkaBuilder(db, connKafka)
		b.KafkaBrokers([]KafkaBroker{
			{
				Broker:                "b-1.hostname-1:9096",
				TargetGroupPort:       9001,
				AvailabilityZone:      "use1-az1",
				PrivateLinkConnection: IdentifierSchemaStruct{SchemaName: "schema", Name: "privatelink_conn", DatabaseName: "database"},
			},
			{
				Broker:                "b-1.hostname-1:9097",
				TargetGroupPort:       9002,
				AvailabilityZone:      "use1-az2",
				PrivateLinkConnection: IdentifierSchemaStruct{SchemaName: "schema", Name: "privatelink_conn", DatabaseName: "database"},
			},
		})
		b.KafkaSASLMechanisms("PLAIN")
		b.KafkaSASLUsername(ValueSecretStruct{Text: "user"})
		b.KafkaSASLPassword(IdentifierSchemaStruct{SchemaName: "schema", Name: "password", DatabaseName: "database"})
		b.Validate(true)

		if err := b.Create(); err != nil {
			t.Fatal(err)
		}
	})

}

func TestConnectionKafkaAwsPrivateLinkCreate(t *testing.T) {
	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		expectedSQL := `CREATE CONNECTION "database"."schema"."kafka_conn" TO KAFKA \( AWS PRIVATELINK "database"."schema"."privatelink_conn" \(PORT 9000\), SASL MECHANISMS = 'PLAIN', SASL USERNAME = 'user', SASL PASSWORD = SECRET "database"."schema"."password"\);`

		mock.ExpectExec(expectedSQL).WillReturnResult(sqlmock.NewResult(1, 1))
		b := NewConnectionKafkaBuilder(db, connKafka)
		awsPrivateLink := awsPrivateLinkConnection{
			PrivateLinkConnection: IdentifierSchemaStruct{SchemaName: "schema", Name: "privatelink_conn", DatabaseName: "database"},
			PrivateLinkPort:       9000,
		}
		b.KafkaAwsPrivateLink(awsPrivateLink)
		b.KafkaSASLMechanisms("PLAIN")
		b.KafkaSASLUsername(ValueSecretStruct{Text: "user"})
		b.KafkaSASLPassword(IdentifierSchemaStruct{SchemaName: "schema", Name: "password", DatabaseName: "database"})
		b.Validate(true)

		if err := b.Create(); err != nil {
			t.Fatalf("Failed to create Kafka connection with AWS PrivateLink: %v", err)
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("Not all expectations were met: %v", err)
		}
	})
}

func TestConnectionKafkaProgressTopicReplicationFactorCreate(t *testing.T) {
	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		mock.ExpectExec(
			`CREATE CONNECTION "database"."schema"."kafka_conn" TO KAFKA \(BROKERS \('localhost:9092'\), SECURITY PROTOCOL = 'PLAIN', PROGRESS TOPIC 'topic', PROGRESS TOPIC REPLICATION FACTOR 3, SASL MECHANISMS = 'PLAIN', SASL USERNAME = 'user', SASL PASSWORD = SECRET "database"."schema"."password"\);`,
		).WillReturnResult(sqlmock.NewResult(1, 1))

		b := NewConnectionKafkaBuilder(db, connKafka)
		b.KafkaBrokers([]KafkaBroker{
			{
				Broker: "localhost:9092",
			},
		})
		b.KafkaProgressTopic("topic")
		b.KafkaProgressTopicReplicationFactor(3)
		b.KafkaSecurityProtocol("PLAIN")
		b.KafkaSASLMechanisms("PLAIN")
		b.KafkaSASLUsername(ValueSecretStruct{Text: "user"})
		b.KafkaSASLPassword(IdentifierSchemaStruct{Name: "password", DatabaseName: "database", SchemaName: "schema"})
		b.Validate(true)

		if err := b.Create(); err != nil {
			t.Fatal(err)
		}
	})
}

func TestConnectionKafkaAwsIAMAuthCreate(t *testing.T) {
	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		expectedSQL := `CREATE CONNECTION "database"."schema"."kafka_conn" TO KAFKA \(BROKERS \('broker1:9092', 'broker2:9092'\), SECURITY PROTOCOL = 'SASL_SSL', AWS CONNECTION = "database"."schema"."aws_conn"\);`

		mock.ExpectExec(expectedSQL).WillReturnResult(sqlmock.NewResult(1, 1))

		b := NewConnectionKafkaBuilder(db, connKafka)
		b.KafkaBrokers([]KafkaBroker{
			{Broker: "broker1:9092"},
			{Broker: "broker2:9092"},
		})
		b.KafkaSecurityProtocol("SASL_SSL")
		b.AwsConnection(IdentifierSchemaStruct{
			Name:         "aws_conn",
			SchemaName:   "schema",
			DatabaseName: "database",
		})
		b.Validate(true)

		if err := b.Create(); err != nil {
			t.Fatalf("Failed to create Kafka connection with AWS IAM authentication: %v", err)
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("Not all expectations were met: %v", err)
		}
	})
}
