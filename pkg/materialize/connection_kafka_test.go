package materialize

import (
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/MaterializeInc/terraform-provider-materialize/pkg/testhelpers"
	"github.com/jmoiron/sqlx"
)

func TestConnectionKafkaCreate(t *testing.T) {
	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		mock.ExpectExec(
			`CREATE CONNECTION "database"."schema"."kafka_conn" TO KAFKA \(BROKERS \('localhost:9092'\), PROGRESS TOPIC 'topic', SASL MECHANISMS = 'PLAIN', SASL USERNAME = 'user', SASL PASSWORD = SECRET "database"."schema"."password"\);`,
		).WillReturnResult(sqlmock.NewResult(1, 1))

		b := NewConnectionKafkaBuilder(db, "kafka_conn", "schema", "database")
		b.KafkaBrokers([]KafkaBroker{
			{
				Broker: "localhost:9092",
			},
		})
		b.KafkaProgressTopic("topic")
		b.KafkaSASLMechanisms("PLAIN")
		b.KafkaSASLUsername(ValueSecretStruct{Text: "user"})
		b.KafkaSASLPassword(IdentifierSchemaStruct{Name: "password", DatabaseName: "database", SchemaName: "schema"})

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

		b := NewConnectionKafkaBuilder(db, "kafka_conn", "schema", "database")
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

		b := NewConnectionKafkaBuilder(db, "kafka_conn", "schema", "database")
		b.KafkaBrokers([]KafkaBroker{
			{
				Broker: "localhost:9092",
			},
		})
		b.KafkaProgressTopic("topic")
		b.KafkaSASLMechanisms("PLAIN")
		b.KafkaSASLUsername(ValueSecretStruct{Text: "user"})
		b.KafkaSASLPassword(IdentifierSchemaStruct{Name: "password", DatabaseName: "database", SchemaName: "schema"})
		b.KafkaSSHTunnel(IdentifierSchemaStruct{Name: "ssh_conn", DatabaseName: "database", SchemaName: "schema"})

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

		b := NewConnectionKafkaBuilder(db, "kafka_conn", "schema", "database")
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

		if err := b.Create(); err != nil {
			t.Fatal(err)
		}
	})
}

func TestConnectionKafkaBrokersSshCreate(t *testing.T) {
	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		mock.ExpectExec(
			`CREATE CONNECTION "database"."schema"."kafka_conn" TO KAFKA \(BROKERS \('localhost:9092' USING SSH TUNNEL "database"."schema"."ssh_conn",'localhost:9093' USING SSH TUNNEL "database"."schema"."ssh_conn"\), PROGRESS TOPIC 'topic', SASL MECHANISMS = 'PLAIN', SASL USERNAME = 'user', SASL PASSWORD = SECRET "database"."schema"."password"\);`,
		).WillReturnResult(sqlmock.NewResult(1, 1))

		b := NewConnectionKafkaBuilder(db, "kafka_conn", "schema", "database")
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
		b.KafkaSSHTunnel(IdentifierSchemaStruct{Name: "ssh_conn", DatabaseName: "database", SchemaName: "schema"})

		if err := b.Create(); err != nil {
			t.Fatal(err)
		}
	})
}

func TestConnectionKafkaSslCreate(t *testing.T) {
	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		mock.ExpectExec(
			`CREATE CONNECTION "database"."schema"."kafka_conn" TO KAFKA \(BROKERS \('localhost:9092'\), PROGRESS TOPIC 'topic', SSL CERTIFICATE AUTHORITY = SECRET "database"."schema"."ca", SSL CERTIFICATE = SECRET "database"."schema"."cert", SSL KEY = SECRET "database"."schema"."key"\);`,
		).WillReturnResult(sqlmock.NewResult(1, 1))

		b := NewConnectionKafkaBuilder(db, "kafka_conn", "schema", "database")
		b.KafkaBrokers([]KafkaBroker{
			{
				Broker: "localhost:9092",
			},
		})
		b.KafkaProgressTopic("topic")
		b.KafkaSSLKey(IdentifierSchemaStruct{SchemaName: "schema", Name: "key", DatabaseName: "database"})
		b.KafkaSSLCert(ValueSecretStruct{Secret: IdentifierSchemaStruct{SchemaName: "schema", Name: "cert", DatabaseName: "database"}})
		b.KafkaSSLCa(ValueSecretStruct{Secret: IdentifierSchemaStruct{SchemaName: "schema", Name: "ca", DatabaseName: "database"}})
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

		b := NewConnectionKafkaBuilder(db, "kafka_conn", "schema", "database")
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

		if err := b.Create(); err != nil {
			t.Fatal(err)
		}
	})
}
