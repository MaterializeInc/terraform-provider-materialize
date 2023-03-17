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
		"ssl_certificate_authority": []interface{}{map[string]interface{}{"text": "key"}},
		"ssl_certificate":           []interface{}{map[string]interface{}{"secret": []interface{}{map[string]interface{}{"name": "cert"}}}},
		"ssl_key":                   []interface{}{map[string]interface{}{"name": "key"}},
		"sasl_mechanisms":           "PLAIN",
		"sasl_username":             []interface{}{map[string]interface{}{"text": "username"}},
		"sasl_password":             []interface{}{map[string]interface{}{"name": "password"}},
		"ssh_tunnel":                []interface{}{map[string]interface{}{"name": "tunnel"}},
	}
	d := schema.TestResourceDataRaw(t, ConnectionKafka().Schema, in)
	r.NotNil(d)

	WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		// Create
		mock.ExpectExec(
			`CREATE CONNECTION "database"."schema"."conn" TO KAFKA \(BROKERS \('b-1.hostname-1:9096' USING SSH TUNNEL "database"."schema"."tunnel"\), PROGRESS TOPIC 'topic', SSL CERTIFICATE AUTHORITY = 'key', SSL CERTIFICATE = SECRET "database"."schema"."cert", SSL KEY = SECRET "database"."schema"."key", SASL MECHANISMS = 'PLAIN', SASL USERNAME = 'username', SASL PASSWORD = SECRET "database"."schema"."password"\);`,
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
		mock.ExpectExec(`DROP CONNECTION "database"."schema"."conn";`).WillReturnResult(sqlmock.NewResult(1, 1))

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
	r.Equal(`ALTER CONNECTION "database"."schema"."connection" RENAME TO "database"."schema"."new_connection";`, b.Rename("new_connection"))
}

func TestConnectionKafkaDropQuery(t *testing.T) {
	r := require.New(t)
	b := newConnectionKafkaBuilder("connection", "schema", "database")
	r.Equal(`DROP CONNECTION "database"."schema"."connection";`, b.Drop())
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
	b.KafkaSASLUsername(ValueSecretStruct{Text: "user"})
	b.KafkaSASLPassword(IdentifierSchemaStruct{Name: "password", DatabaseName: "database", SchemaName: "schema"})
	r.Equal(`CREATE CONNECTION "database"."schema"."kafka_conn" TO KAFKA (BROKERS ('localhost:9092'), PROGRESS TOPIC 'topic', SASL MECHANISMS = 'PLAIN', SASL USERNAME = 'user', SASL PASSWORD = SECRET "database"."schema"."password");`, b.Create())
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
	b.KafkaSASLUsername(ValueSecretStruct{Text: "user"})
	b.KafkaSASLPassword(IdentifierSchemaStruct{Name: "password", DatabaseName: "database", SchemaName: "schema"})
	r.Equal(`CREATE CONNECTION "database"."schema"."kafka_conn" TO KAFKA (BROKERS ('localhost:9092', 'localhost:9093'), PROGRESS TOPIC 'topic', SASL MECHANISMS = 'PLAIN', SASL USERNAME = 'user', SASL PASSWORD = SECRET "database"."schema"."password");`, b.Create())
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
	b.KafkaSASLUsername(ValueSecretStruct{Text: "user"})
	b.KafkaSASLPassword(IdentifierSchemaStruct{Name: "password", DatabaseName: "database", SchemaName: "schema"})
	b.KafkaSSHTunnel(IdentifierSchemaStruct{Name: "ssh_conn", DatabaseName: "database", SchemaName: "schema"})
	r.Equal(`CREATE CONNECTION "database"."schema"."kafka_conn" TO KAFKA (BROKERS ('localhost:9092' USING SSH TUNNEL "database"."schema"."ssh_conn"), PROGRESS TOPIC 'topic', SASL MECHANISMS = 'PLAIN', SASL USERNAME = 'user', SASL PASSWORD = SECRET "database"."schema"."password");`, b.Create())
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
	b.KafkaSASLUsername(ValueSecretStruct{Text: "user"})
	b.KafkaSASLPassword(IdentifierSchemaStruct{Name: "password", DatabaseName: "database", SchemaName: "schema"})
	r.Equal(`CREATE CONNECTION "database"."schema"."kafka_conn" TO KAFKA (BROKERS ('localhost:9092', 'localhost:9093'), PROGRESS TOPIC 'topic', SASL MECHANISMS = 'PLAIN', SASL USERNAME = 'user', SASL PASSWORD = SECRET "database"."schema"."password");`, b.Create())
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
	b.KafkaSASLUsername(ValueSecretStruct{Text: "user"})
	b.KafkaSASLPassword(IdentifierSchemaStruct{Name: "password", DatabaseName: "database", SchemaName: "schema"})
	b.KafkaSSHTunnel(IdentifierSchemaStruct{Name: "ssh_conn", DatabaseName: "database", SchemaName: "schema"})
	r.Equal(`CREATE CONNECTION "database"."schema"."kafka_conn" TO KAFKA (BROKERS ('localhost:9092' USING SSH TUNNEL "database"."schema"."ssh_conn",'localhost:9093' USING SSH TUNNEL "database"."schema"."ssh_conn"), PROGRESS TOPIC 'topic', SASL MECHANISMS = 'PLAIN', SASL USERNAME = 'user', SASL PASSWORD = SECRET "database"."schema"."password");`, b.Create())
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
	b.KafkaSSLKey(IdentifierSchemaStruct{SchemaName: "schema", Name: "key", DatabaseName: "database"})
	b.KafkaSSLCert(ValueSecretStruct{Secret: IdentifierSchemaStruct{SchemaName: "schema", Name: "cert", DatabaseName: "database"}})
	b.KafkaSSLCa(ValueSecretStruct{Secret: IdentifierSchemaStruct{SchemaName: "schema", Name: "ca", DatabaseName: "database"}})
	r.Equal(`CREATE CONNECTION "database"."schema"."kafka_conn" TO KAFKA (BROKERS ('localhost:9092'), PROGRESS TOPIC 'topic', SSL CERTIFICATE AUTHORITY = SECRET "database"."schema"."ca", SSL CERTIFICATE = SECRET "database"."schema"."cert", SSL KEY = SECRET "database"."schema"."key");`, b.Create())
}

func TestConnectionKafkaAwsPrivatelinkQuery(t *testing.T) {
	r := require.New(t)
	b := newConnectionKafkaBuilder("kafka_conn", "schema", "database")
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
	r.Equal(`CREATE CONNECTION "database"."schema"."kafka_conn" TO KAFKA (BROKERS ('b-1.hostname-1:9096' USING AWS PRIVATELINK "database"."schema"."privatelink_conn" (PORT 9001, AVAILABILITY ZONE 'use1-az1'), 'b-1.hostname-1:9097' USING AWS PRIVATELINK "database"."schema"."privatelink_conn" (PORT 9002, AVAILABILITY ZONE 'use1-az2')), SASL MECHANISMS = 'PLAIN', SASL USERNAME = 'user', SASL PASSWORD = SECRET "database"."schema"."password");`, b.Create())
}
