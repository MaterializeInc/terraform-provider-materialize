package resources

import (
	"context"
	"terraform-materialize/pkg/testhelpers"
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
		"sasl_mechanism":            "PLAIN",
		"sasl_username":             []interface{}{map[string]interface{}{"text": "username"}},
		"sasl_password":             []interface{}{map[string]interface{}{"name": "password"}},
		"ssh_tunnel":                []interface{}{map[string]interface{}{"name": "tunnel"}},
	}
	d := schema.TestResourceDataRaw(t, ConnectionKafka().Schema, in)
	r.NotNil(d)

	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
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
		ip := sqlmock.NewRows([]string{"name", "schema", "database"}).
			AddRow("conn", "schema", "database")
		mock.ExpectQuery(`
			SELECT
				mz_connections.name,
				mz_schemas.name,
				mz_databases.name
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

	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		mock.ExpectExec(`DROP CONNECTION "database"."schema"."conn";`).WillReturnResult(sqlmock.NewResult(1, 1))

		if err := connectionKafkaDelete(context.TODO(), d, db); err != nil {
			t.Fatal(err)
		}
	})

}
