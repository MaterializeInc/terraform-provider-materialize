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

var inConfluentSchemaRegistry = map[string]interface{}{
	"name":                      "conn",
	"schema_name":               "schema",
	"database_name":             "database",
	"service_name":              "service",
	"url":                       "http://localhost:8081",
	"ssl_certificate_authority": []interface{}{map[string]interface{}{"secret": []interface{}{map[string]interface{}{"name": "ssl"}}}},
	"ssl_certificate":           []interface{}{map[string]interface{}{"secret": []interface{}{map[string]interface{}{"name": "ssl"}}}},
	"ssl_key": []interface{}{
		map[string]interface{}{
			"name":          "ssl",
			"database_name": "ssl_key",
		},
	},
	"password": []interface{}{map[string]interface{}{"name": "password"}},
	"username": []interface{}{map[string]interface{}{"text": "user"}},
	"ssh_tunnel": []interface{}{
		map[string]interface{}{
			"name":        "tunnel",
			"schema_name": "tunnel_schema",
		},
	},
	"aws_privatelink": []interface{}{map[string]interface{}{"name": "privatelink"}},
	"comment":         "object comment",
}

func TestResourceConnectionConfluentSchemaRegistryCreate(t *testing.T) {
	r := require.New(t)
	d := schema.TestResourceDataRaw(t, ConnectionConfluentSchemaRegistry().Schema, inConfluentSchemaRegistry)
	r.NotNil(d)

	testhelpers.WithMockProviderMeta(t, func(db *utils.ProviderMeta, mock sqlmock.Sqlmock) {
		// Create
		mock.ExpectExec(
			`CREATE CONNECTION "database"."schema"."conn" TO CONFLUENT SCHEMA REGISTRY \(URL 'http://localhost:8081', USERNAME = 'user', PASSWORD = SECRET "materialize"."public"."password", SSL CERTIFICATE AUTHORITY = SECRET "materialize"."public"."ssl", SSL CERTIFICATE = SECRET "materialize"."public"."ssl", SSL KEY = SECRET "ssl_key"."public"."ssl", AWS PRIVATELINK "materialize"."public"."privatelink", SSH TUNNEL "materialize"."tunnel_schema"."tunnel"\)`,
		).WillReturnResult(sqlmock.NewResult(1, 1))

		// Comment
		mock.ExpectExec(`COMMENT ON CONNECTION "database"."schema"."conn" IS 'object comment';`).WillReturnResult(sqlmock.NewResult(1, 1))

		// Query Id
		ip := `WHERE mz_connections.name = 'conn' AND mz_databases.name = 'database' AND mz_schemas.name = 'schema'`
		testhelpers.MockConnectionScan(mock, ip)

		// Query Params
		pp := `WHERE mz_connections.id = 'u1'`
		testhelpers.MockConnectionScan(mock, pp)

		if err := connectionConfluentSchemaRegistryCreate(context.TODO(), d, db); err != nil {
			t.Fatal(err)
		}
	})
}

func TestResourceConnectionConfluentSchemaRegistryUpdate(t *testing.T) {
	r := require.New(t)
	d := schema.TestResourceDataRaw(t, ConnectionConfluentSchemaRegistry().Schema, inConfluentSchemaRegistry)

	// Set current state
	d.SetId("u1")
	d.Set("name", "old_conn")
	r.NotNil(d)

	testhelpers.WithMockProviderMeta(t, func(db *utils.ProviderMeta, mock sqlmock.Sqlmock) {
		mock.ExpectExec(`ALTER CONNECTION "database"."schema"."" RENAME TO "conn";`).WillReturnResult(sqlmock.NewResult(1, 1))

		mock.ExpectExec(`ALTER CONNECTION "database"."schema"."old_conn" SET \(URL = 'http://localhost:8081'\);`).WillReturnResult(sqlmock.NewResult(1, 1))

		mock.ExpectExec(`ALTER CONNECTION "database"."schema"."old_conn" SET \(USER = 'user'\);`).WillReturnResult(sqlmock.NewResult(1, 1))

		mock.ExpectExec(`ALTER CONNECTION "database"."schema"."old_conn" SET \(PASSWORD = SECRET "materialize"."public"."password"\);`).WillReturnResult(sqlmock.NewResult(1, 1))

		mock.ExpectExec(`ALTER CONNECTION "database"."schema"."old_conn" SET \(SSL CERTIFICATE AUTHORITY = SECRET "materialize"."public"."ssl"\);`).WillReturnResult(sqlmock.NewResult(1, 1))

		mock.ExpectExec(`ALTER CONNECTION "database"."schema"."old_conn" SET \((SSL CERTIFICATE = SECRET "materialize"."public"."ssl"\), SET \(SSL KEY = SECRET "ssl_key"."public"."ssl"|SSL KEY = SECRET "ssl_key"."public"."ssl"\), SET \(SSL CERTIFICATE = SECRET "materialize"."public"."ssl")\);`).WillReturnResult(sqlmock.NewResult(1, 1))

		// Comment
		mock.ExpectExec(`COMMENT ON CONNECTION "database"."schema"."old_conn" IS 'object comment';`).WillReturnResult(sqlmock.NewResult(1, 1))

		// Query Params
		pp := `WHERE mz_connections.id = 'u1'`
		testhelpers.MockConnectionScan(mock, pp)

		// Execute the update function
		if err := connectionConfluentSchemaRegistryUpdate(context.TODO(), d, db); err != nil {
			t.Fatal(err)
		}

		// Check that all expectations were met
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("there were unfulfilled expectations: %s", err)
		}
	})
}
