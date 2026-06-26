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

// TestBuildBrokersStringPermutations exercises every combination of static
// brokers and wildcard MATCHING rules that BuildBrokersString must emit. The
// same method backs both Create and the BROKERS clause rebuilt on update, so
// these cases cover the syntax for add/remove/change of any broker or rule.
func TestBuildBrokersStringPermutations(t *testing.T) {
	plConn := IdentifierSchemaStruct{SchemaName: "schema", Name: "pl_conn", DatabaseName: "database"}
	plConn2 := IdentifierSchemaStruct{SchemaName: "schema", Name: "pl_conn_2", DatabaseName: "database"}

	cases := []struct {
		name     string
		brokers  []KafkaBroker
		rules    []KafkaBrokerMatchingRule
		expected string
	}{
		{
			name:     "only static brokers, no rules",
			brokers:  []KafkaBroker{{Broker: "localhost:9092"}, {Broker: "localhost:9093"}},
			expected: `'localhost:9092', 'localhost:9093'`,
		},
		{
			name:    "static broker with privatelink plus single rule, az only",
			brokers: []KafkaBroker{{Broker: "boot:9092", PrivateLinkConnection: plConn}},
			rules: []KafkaBrokerMatchingRule{
				{Pattern: "*.use1-az1.*", AvailabilityZone: "use1-az1", PrivateLinkConnection: plConn},
			},
			expected: `'boot:9092' USING AWS PRIVATELINK "database"."schema"."pl_conn", MATCHING '*.use1-az1.*' USING AWS PRIVATELINK "database"."schema"."pl_conn" (AVAILABILITY ZONE 'use1-az1')`,
		},
		{
			name:    "static broker plus multiple rules preserve order",
			brokers: []KafkaBroker{{Broker: "boot:9092", PrivateLinkConnection: plConn}},
			rules: []KafkaBrokerMatchingRule{
				{Pattern: "*.use1-az1.*", AvailabilityZone: "use1-az1", PrivateLinkConnection: plConn},
				{Pattern: "*.use1-az4.*", AvailabilityZone: "use1-az4", PrivateLinkConnection: plConn},
				{Pattern: "*.use1-az6.*", AvailabilityZone: "use1-az6", PrivateLinkConnection: plConn},
			},
			expected: `'boot:9092' USING AWS PRIVATELINK "database"."schema"."pl_conn", MATCHING '*.use1-az1.*' USING AWS PRIVATELINK "database"."schema"."pl_conn" (AVAILABILITY ZONE 'use1-az1'), MATCHING '*.use1-az4.*' USING AWS PRIVATELINK "database"."schema"."pl_conn" (AVAILABILITY ZONE 'use1-az4'), MATCHING '*.use1-az6.*' USING AWS PRIVATELINK "database"."schema"."pl_conn" (AVAILABILITY ZONE 'use1-az6')`,
		},
		{
			name:    "rule with port only",
			brokers: []KafkaBroker{{Broker: "boot:9092"}},
			rules: []KafkaBrokerMatchingRule{
				{Pattern: "*az1*", TargetGroupPort: 9001, PrivateLinkConnection: plConn},
			},
			expected: `'boot:9092', MATCHING '*az1*' USING AWS PRIVATELINK "database"."schema"."pl_conn" (PORT 9001)`,
		},
		{
			name:    "rule with port and az",
			brokers: []KafkaBroker{{Broker: "boot:9092"}},
			rules: []KafkaBrokerMatchingRule{
				{Pattern: "*az1*", TargetGroupPort: 9001, AvailabilityZone: "use1-az1", PrivateLinkConnection: plConn},
			},
			expected: `'boot:9092', MATCHING '*az1*' USING AWS PRIVATELINK "database"."schema"."pl_conn" (PORT 9001, AVAILABILITY ZONE 'use1-az1')`,
		},
		{
			name:    "rule with neither port nor az",
			brokers: []KafkaBroker{{Broker: "boot:9092"}},
			rules: []KafkaBrokerMatchingRule{
				{Pattern: "*az1*", PrivateLinkConnection: plConn},
			},
			expected: `'boot:9092', MATCHING '*az1*' USING AWS PRIVATELINK "database"."schema"."pl_conn"`,
		},
		{
			name:    "rules routed through different privatelink connections",
			brokers: []KafkaBroker{{Broker: "boot:9092", PrivateLinkConnection: plConn}},
			rules: []KafkaBrokerMatchingRule{
				{Pattern: "*az1*", AvailabilityZone: "use1-az1", PrivateLinkConnection: plConn},
				{Pattern: "*az4*", AvailabilityZone: "use1-az4", PrivateLinkConnection: plConn2},
			},
			expected: `'boot:9092' USING AWS PRIVATELINK "database"."schema"."pl_conn", MATCHING '*az1*' USING AWS PRIVATELINK "database"."schema"."pl_conn" (AVAILABILITY ZONE 'use1-az1'), MATCHING '*az4*' USING AWS PRIVATELINK "database"."schema"."pl_conn_2" (AVAILABILITY ZONE 'use1-az4')`,
		},
		{
			name:     "no brokers, no rules yields empty string",
			expected: ``,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			b := &ConnectionKafkaBuilder{}
			b.KafkaBrokers(tc.brokers)
			b.KafkaBrokerMatchingRules(tc.rules)
			if got := b.BuildBrokersString(); got != tc.expected {
				t.Errorf("BuildBrokersString()\n got: %s\nwant: %s", got, tc.expected)
			}
		})
	}
}

func TestConnectionKafkaBrokerMatchingRulesCreate(t *testing.T) {
	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		mock.ExpectExec(
			`CREATE CONNECTION "database"."schema"."kafka_conn" TO KAFKA \(BROKERS \('lkc-825730.endpoint.cloud:9092' USING AWS PRIVATELINK "database"."schema"."privatelink_conn", MATCHING '\*.use1-az1.\*' USING AWS PRIVATELINK "database"."schema"."privatelink_conn" \(AVAILABILITY ZONE 'use1-az1'\), MATCHING '\*.use1-az4.\*' USING AWS PRIVATELINK "database"."schema"."privatelink_conn" \(AVAILABILITY ZONE 'use1-az4'\)\), SASL MECHANISMS = 'PLAIN', SASL USERNAME = 'user', SASL PASSWORD = SECRET "database"."schema"."password"\);`,
		).WillReturnResult(sqlmock.NewResult(1, 1))

		b := NewConnectionKafkaBuilder(db, connKafka)
		b.KafkaBrokers([]KafkaBroker{
			{
				Broker:                "lkc-825730.endpoint.cloud:9092",
				PrivateLinkConnection: IdentifierSchemaStruct{SchemaName: "schema", Name: "privatelink_conn", DatabaseName: "database"},
			},
		})
		b.KafkaBrokerMatchingRules([]KafkaBrokerMatchingRule{
			{
				Pattern:               "*.use1-az1.*",
				AvailabilityZone:      "use1-az1",
				PrivateLinkConnection: IdentifierSchemaStruct{SchemaName: "schema", Name: "privatelink_conn", DatabaseName: "database"},
			},
			{
				Pattern:               "*.use1-az4.*",
				AvailabilityZone:      "use1-az4",
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

func TestConnectionKafkaBrokerMatchingRulesWithPortCreate(t *testing.T) {
	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		mock.ExpectExec(
			`CREATE CONNECTION "database"."schema"."kafka_conn" TO KAFKA \(BROKERS \('broker:9092', MATCHING '\*.use1-az1.\*' USING AWS PRIVATELINK "database"."schema"."privatelink_conn" \(PORT 9001, AVAILABILITY ZONE 'use1-az1'\)\)\) WITH \(VALIDATE = false\);`,
		).WillReturnResult(sqlmock.NewResult(1, 1))

		b := NewConnectionKafkaBuilder(db, connKafka)
		b.KafkaBrokers([]KafkaBroker{
			{Broker: "broker:9092"},
		})
		b.KafkaBrokerMatchingRules([]KafkaBrokerMatchingRule{
			{
				Pattern:               "*.use1-az1.*",
				TargetGroupPort:       9001,
				AvailabilityZone:      "use1-az1",
				PrivateLinkConnection: IdentifierSchemaStruct{SchemaName: "schema", Name: "privatelink_conn", DatabaseName: "database"},
			},
		})
		b.Validate(false)

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
