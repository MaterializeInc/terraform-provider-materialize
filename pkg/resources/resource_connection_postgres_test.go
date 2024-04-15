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

var inPostgres = map[string]interface{}{
	"name":          "conn",
	"schema_name":   "schema",
	"database_name": "database",
	"database":      "default",
	"host":          "postgres_host",
	"port":          5432,
	"user":          []interface{}{map[string]interface{}{"secret": []interface{}{map[string]interface{}{"name": "user"}}}},
	"password":      []interface{}{map[string]interface{}{"name": "password"}},
	"ssh_tunnel": []interface{}{
		map[string]interface{}{
			"name":          "ssh_conn",
			"schema_name":   "tunnel_schema",
			"database_name": "tunnel_database",
		},
	},
	"ssl_certificate_authority": []interface{}{
		map[string]interface{}{
			"secret": []interface{}{map[string]interface{}{
				"name":          "root",
				"database_name": "ssl_database",
			}},
		},
	},
	"ssl_certificate": []interface{}{map[string]interface{}{"secret": []interface{}{map[string]interface{}{"name": "cert"}}}},
	"ssl_key":         []interface{}{map[string]interface{}{"name": "key"}},
	"ssl_mode":        "verify-full",
	"aws_privatelink": []interface{}{map[string]interface{}{"name": "link"}},
	"comment":         "object comment",
}

func TestResourceConnectionPostgresCreate(t *testing.T) {
	r := require.New(t)
	d := schema.TestResourceDataRaw(t, ConnectionPostgres().Schema, inPostgres)
	r.NotNil(d)

	testhelpers.WithMockProviderMeta(t, func(db *utils.ProviderMeta, mock sqlmock.Sqlmock) {
		// Create
		mock.ExpectExec(
			`CREATE CONNECTION "database"."schema"."conn" TO POSTGRES \(HOST 'postgres_host', PORT 5432, USER SECRET "materialize"."public"."user", PASSWORD SECRET "materialize"."public"."password", SSL MODE 'verify-full', SSH TUNNEL "tunnel_database"."tunnel_schema"."ssh_conn", SSL CERTIFICATE AUTHORITY SECRET "ssl_database"."public"."root", SSL CERTIFICATE SECRET "materialize"."public"."cert", SSL KEY SECRET "materialize"."public"."key", AWS PRIVATELINK "materialize"."public"."link", DATABASE 'default'\);`,
		).WillReturnResult(sqlmock.NewResult(1, 1))

		// Comment
		mock.ExpectExec(`COMMENT ON CONNECTION "database"."schema"."conn" IS 'object comment';`).WillReturnResult(sqlmock.NewResult(1, 1))

		// Query Id
		ip := `WHERE mz_connections.name = 'conn' AND mz_databases.name = 'database' AND mz_schemas.name = 'schema'`
		testhelpers.MockConnectionScan(mock, ip)

		// Query Params
		pp := `WHERE mz_connections.id = 'u1'`
		testhelpers.MockConnectionScan(mock, pp)

		if err := connectionPostgresCreate(context.TODO(), d, db); err != nil {
			t.Fatal(err)
		}
	})
}

func TestResourceConnectionPostgresUpdate(t *testing.T) {
	r := require.New(t)
	d := schema.TestResourceDataRaw(t, ConnectionPostgres().Schema, inPostgres)

	// Set current state
	d.SetId("u1")
	d.Set("name", "conn")
	r.NotNil(d)

	testhelpers.WithMockProviderMeta(t, func(db *utils.ProviderMeta, mock sqlmock.Sqlmock) {
		// Name Change
		mock.ExpectExec(`ALTER CONNECTION "database"."schema"."" RENAME TO "conn";`).WillReturnResult(sqlmock.NewResult(1, 1))

		// Host
		mock.ExpectExec(`ALTER CONNECTION "database"."schema"."conn" SET \(HOST = 'postgres_host'\);`).WillReturnResult(sqlmock.NewResult(1, 1))

		// Port
		mock.ExpectExec(`ALTER CONNECTION "database"."schema"."conn" SET \(PORT = 5432\);`).WillReturnResult(sqlmock.NewResult(1, 1))

		// User
		mock.ExpectExec(`ALTER CONNECTION "database"."schema"."conn" SET \(USER = SECRET "materialize"."public"."user"\);`).WillReturnResult(sqlmock.NewResult(1, 1))

		// Password
		mock.ExpectExec(`ALTER CONNECTION "database"."schema"."conn" SET \(PASSWORD = SECRET "materialize"."public"."password"\);`).WillReturnResult(sqlmock.NewResult(1, 1))

		// Database
		mock.ExpectExec(`ALTER CONNECTION "database"."schema"."conn" SET \(DATABASE = 'default'\);`).WillReturnResult(sqlmock.NewResult(1, 1))

		// SSH Tunnel
		mock.ExpectExec(`ALTER CONNECTION "database"."schema"."conn" SET \(SSH TUNNEL = "tunnel_database"."tunnel_schema"."ssh_conn"\);`).WillReturnResult(sqlmock.NewResult(1, 1))

		// SSL Certificate Authority
		mock.ExpectExec(`ALTER CONNECTION "database"."schema"."conn" SET \(SSL CERTIFICATE AUTHORITY = SECRET "ssl_database"."public"."root"\);`).WillReturnResult(sqlmock.NewResult(1, 1))

		// SSL Certificate and Key
		mock.ExpectExec(`ALTER CONNECTION "database"."schema"."conn" SET \(SSL CERTIFICATE = SECRET "materialize"."public"."cert"\), SET \(SSL KEY = SECRET "materialize"."public"."key"\);`).WillReturnResult(sqlmock.NewResult(1, 1))

		// SSL Mode
		mock.ExpectExec(`ALTER CONNECTION "database"."schema"."conn" SET \(SSL MODE = 'verify-full'\);`).WillReturnResult(sqlmock.NewResult(1, 1))

		// AWS Privatelink
		mock.ExpectExec(`ALTER CONNECTION "database"."schema"."conn" SET \(AWS PRIVATELINK = "materialize"."public"."link"\);`).WillReturnResult(sqlmock.NewResult(1, 1))

		// Comment
		mock.ExpectExec(`COMMENT ON CONNECTION "database"."schema"."conn" IS 'object comment';`).WillReturnResult(sqlmock.NewResult(1, 1))

		// Query Params
		p := `WHERE mz_connections.id = 'u1'`
		testhelpers.MockConnectionScan(mock, p)

		// Execute the update function
		if err := connectionUpdate(context.TODO(), d, db); err != nil {
			t.Fatal(err)
		}
	})
}
