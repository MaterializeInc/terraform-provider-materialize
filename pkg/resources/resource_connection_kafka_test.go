package resources

import (
	"context"
	"testing"

	"github.com/MaterializeInc/terraform-provider-materialize/pkg/testhelpers"
	"github.com/MaterializeInc/terraform-provider-materialize/pkg/utils"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/require"
)

var inKafka = map[string]interface{}{
	"name":          "conn",
	"schema_name":   "schema",
	"database_name": "database",
	"service_name":  "service",
	"kafka_broker": []interface{}{map[string]interface{}{
		"broker":                 "b-1.hostname-1:9096",
		"target_group_port":      9001,
		"availability_zone":      "use1-az1",
		"privatelink_connection": []interface{}{map[string]interface{}{"name": "pl_conn"}},
		"ssh_tunnel":             []interface{}{map[string]interface{}{"name": "ssh"}},
	}},
	"aws_privatelink": []interface{}{map[string]interface{}{
		"privatelink_connection":      []interface{}{map[string]interface{}{"name": "pl_conn"}},
		"privatelink_connection_port": 9001,
	}},
	"security_protocol":                 "SASL_PLAINTEXT",
	"progress_topic":                    "topic",
	"progress_topic_replication_factor": 3,
	"ssl_certificate_authority":         []interface{}{map[string]interface{}{"text": "key"}},
	"ssl_certificate":                   []interface{}{map[string]interface{}{"secret": []interface{}{map[string]interface{}{"name": "cert"}}}},
	"ssl_key":                           []interface{}{map[string]interface{}{"name": "key"}},
	"sasl_mechanisms":                   "PLAIN",
	"sasl_username":                     []interface{}{map[string]interface{}{"text": "username"}},
	"sasl_password":                     []interface{}{map[string]interface{}{"name": "password"}},
	"ssh_tunnel":                        []interface{}{map[string]interface{}{"name": "tunnel"}},
	"comment":                           "object comment",
}

func TestResourceConnectionKafkaCreate(t *testing.T) {
	r := require.New(t)

	d := schema.TestResourceDataRaw(t, ConnectionKafka().Schema, inKafka)
	r.NotNil(d)

	testhelpers.WithMockProviderMeta(t, func(db *utils.ProviderMeta, mock sqlmock.Sqlmock) {
		// Create
		mock.ExpectExec(
			`CREATE CONNECTION "database"."schema"."conn"
			TO KAFKA \(BROKERS
				\('b-1.hostname-1:9096'
				USING SSH TUNNEL "materialize"."public"."ssh"
				USING AWS PRIVATELINK "materialize"."public"."pl_conn"
				\(PORT 9001, AVAILABILITY ZONE 'use1-az1'\)\)
			AWS PRIVATELINK "materialize"."public"."pl_conn" \(PORT 9001\),
			SSH TUNNEL "materialize"."public"."tunnel",
			SECURITY PROTOCOL = 'SASL_PLAINTEXT',
			PROGRESS TOPIC 'topic',
			PROGRESS TOPIC REPLICATION FACTOR 3,
			SSL CERTIFICATE AUTHORITY = 'key',
			SSL CERTIFICATE = SECRET "materialize"."public"."cert",
			SSL KEY = SECRET "materialize"."public"."key",
			SASL MECHANISMS = 'PLAIN',
			SASL USERNAME = 'username',
			SASL PASSWORD = SECRET "materialize"."public"."password"\);`,
		).WillReturnResult(sqlmock.NewResult(1, 1))

		// Comment
		mock.ExpectExec(`COMMENT ON CONNECTION "database"."schema"."conn" IS 'object comment';`).WillReturnResult(sqlmock.NewResult(1, 1))

		// Query Id
		ip := `WHERE mz_connections.name = 'conn' AND mz_databases.name = 'database' AND mz_schemas.name = 'schema'`
		testhelpers.MockConnectionScan(mock, ip)

		// Query Params
		pp := `WHERE mz_connections.id = 'u1'`
		testhelpers.MockConnectionScan(mock, pp)

		if err := connectionKafkaCreate(context.TODO(), d, db); err != nil {
			t.Fatal(err)
		}
	})
}

func TestResourceConnectionKafkaCreateWithoutReplicationFactor(t *testing.T) {
	r := require.New(t)

	inKafkaWithoutReplicationFactor := map[string]interface{}{}
	for k, v := range inKafka {
		if k != "progress_topic_replication_factor" {
			inKafkaWithoutReplicationFactor[k] = v
		}
	}

	d := schema.TestResourceDataRaw(t, ConnectionKafka().Schema, inKafkaWithoutReplicationFactor)
	r.NotNil(d)

	testhelpers.WithMockProviderMeta(t, func(db *utils.ProviderMeta, mock sqlmock.Sqlmock) {
		// Create
		mock.ExpectExec(
			`CREATE CONNECTION "database"."schema"."conn"
			TO KAFKA \(BROKERS
				\('b-1.hostname-1:9096'
				USING SSH TUNNEL "materialize"."public"."ssh"
				USING AWS PRIVATELINK "materialize"."public"."pl_conn"
				\(PORT 9001, AVAILABILITY ZONE 'use1-az1'\)\)
			AWS PRIVATELINK "materialize"."public"."pl_conn" \(PORT 9001\),
			SSH TUNNEL "materialize"."public"."tunnel",
			SECURITY PROTOCOL = 'SASL_PLAINTEXT',
			PROGRESS TOPIC 'topic',
			SSL CERTIFICATE AUTHORITY = 'key',
			SSL CERTIFICATE = SECRET "materialize"."public"."cert",
			SSL KEY = SECRET "materialize"."public"."key",
			SASL MECHANISMS = 'PLAIN',
			SASL USERNAME = 'username',
			SASL PASSWORD = SECRET "materialize"."public"."password"\);`,
		).WillReturnResult(sqlmock.NewResult(1, 1))

		// Comment
		mock.ExpectExec(`COMMENT ON CONNECTION "database"."schema"."conn" IS 'object comment';`).WillReturnResult(sqlmock.NewResult(1, 1))

		// Query Id
		ip := `WHERE mz_connections.name = 'conn' AND mz_databases.name = 'database' AND mz_schemas.name = 'schema'`
		testhelpers.MockConnectionScan(mock, ip)

		// Query Params
		pp := `WHERE mz_connections.id = 'u1'`
		testhelpers.MockConnectionScan(mock, pp)

		if err := connectionKafkaCreate(context.TODO(), d, db); err != nil {
			t.Fatal(err)
		}
	})
}

func TestResourceConnectionKafkaCreateWithAwsConnection(t *testing.T) {
	r := require.New(t)

	inKafkaWithAwsConnection := map[string]interface{}{
		"name":          "conn",
		"schema_name":   "schema",
		"database_name": "database",
		"kafka_broker": []interface{}{map[string]interface{}{
			"broker": "b-1.hostname-1:9096",
		}},
		"security_protocol": "SASL_SSL",
		"aws_connection":    []interface{}{map[string]interface{}{"name": "aws_conn"}},
		"comment":           "object comment",
	}

	d := schema.TestResourceDataRaw(t, ConnectionKafka().Schema, inKafkaWithAwsConnection)
	r.NotNil(d)

	testhelpers.WithMockProviderMeta(t, func(db *utils.ProviderMeta, mock sqlmock.Sqlmock) {
		// Create
		mock.ExpectExec(
			`CREATE CONNECTION "database"."schema"."conn"
            TO KAFKA \(BROKERS \('b-1.hostname-1:9096'\),
            SECURITY PROTOCOL = 'SASL_SSL',
            AWS CONNECTION = "materialize"."public"."aws_conn"\);`,
		).WillReturnResult(sqlmock.NewResult(1, 1))

		// Comment
		mock.ExpectExec(`COMMENT ON CONNECTION "database"."schema"."conn" IS 'object comment';`).WillReturnResult(sqlmock.NewResult(1, 1))

		// Query Id
		ip := `WHERE mz_connections.name = 'conn' AND mz_databases.name = 'database' AND mz_schemas.name = 'schema'`
		testhelpers.MockConnectionScan(mock, ip)

		// Query Params
		pp := `WHERE mz_connections.id = 'u1'`
		testhelpers.MockConnectionScan(mock, pp)

		if err := connectionKafkaCreate(context.TODO(), d, db); err != nil {
			t.Fatal(err)
		}
	})
}

func TestResourceConnectionKafkaCreateWithAwsConnectionAndPrivateLink(t *testing.T) {
	r := require.New(t)

	inKafkaWithAwsConnectionAndPrivateLink := map[string]interface{}{
		"name":          "conn",
		"schema_name":   "schema",
		"database_name": "database",
		"kafka_broker": []interface{}{map[string]interface{}{
			"broker":                 "b-1.hostname-1:9096",
			"target_group_port":      9001,
			"availability_zone":      "use1-az1",
			"privatelink_connection": []interface{}{map[string]interface{}{"name": "pl_conn"}},
		}},
		"aws_privatelink": []interface{}{map[string]interface{}{
			"privatelink_connection":      []interface{}{map[string]interface{}{"name": "pl_conn"}},
			"privatelink_connection_port": 9001,
		}},
		"security_protocol": "SASL_SSL",
		"aws_connection":    []interface{}{map[string]interface{}{"name": "aws_conn"}},
		"comment":           "object comment",
	}

	d := schema.TestResourceDataRaw(t, ConnectionKafka().Schema, inKafkaWithAwsConnectionAndPrivateLink)
	r.NotNil(d)

	testhelpers.WithMockProviderMeta(t, func(db *utils.ProviderMeta, mock sqlmock.Sqlmock) {
		// Create
		mock.ExpectExec(
			`CREATE CONNECTION "database"."schema"."conn"
            TO KAFKA \(BROKERS
                \('b-1.hostname-1:9096'
                USING AWS PRIVATELINK "materialize"."public"."pl_conn"
                \(PORT 9001, AVAILABILITY ZONE 'use1-az1'\)\)
            AWS PRIVATELINK "materialize"."public"."pl_conn" \(PORT 9001\),
            SECURITY PROTOCOL = 'SASL_SSL',
            AWS CONNECTION = "materialize"."public"."aws_conn"\);`,
		).WillReturnResult(sqlmock.NewResult(1, 1))

		// Comment
		mock.ExpectExec(`COMMENT ON CONNECTION "database"."schema"."conn" IS 'object comment';`).WillReturnResult(sqlmock.NewResult(1, 1))

		// Query Id
		ip := `WHERE mz_connections.name = 'conn' AND mz_databases.name = 'database' AND mz_schemas.name = 'schema'`
		testhelpers.MockConnectionScan(mock, ip)

		// Query Params
		pp := `WHERE mz_connections.id = 'u1'`
		testhelpers.MockConnectionScan(mock, pp)

		if err := connectionKafkaCreate(context.TODO(), d, db); err != nil {
			t.Fatal(err)
		}
	})
}
