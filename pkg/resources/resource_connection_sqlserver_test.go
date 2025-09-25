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

var inSQLServer = map[string]interface{}{
	"name":          "conn",
	"schema_name":   "schema",
	"database_name": "database",
	"host":          "sqlserver_host",
	"port":          1433,
	"database":      "testdb",
	"user":          []interface{}{map[string]interface{}{"secret": []interface{}{map[string]interface{}{"name": "user"}}}},
	"password":      []interface{}{map[string]interface{}{"name": "password"}},
	"ssh_tunnel": []interface{}{
		map[string]interface{}{
			"name":          "ssh_conn",
			"schema_name":   "tunnel_schema",
			"database_name": "tunnel_database",
		},
	},
	"aws_privatelink": []interface{}{
		map[string]interface{}{
			"name":          "aws_conn",
			"schema_name":   "aws_schema",
			"database_name": "aws_database",
		},
	},
	"validate": true,
	"comment":  "SQL Server connection comment",
}

func TestResourceConnectionSQLServerCreate(t *testing.T) {
	r := require.New(t)
	d := schema.TestResourceDataRaw(t, ConnectionSQLServer().Schema, inSQLServer)
	r.NotNil(d)

	testhelpers.WithMockProviderMeta(t, func(db *utils.ProviderMeta, mock sqlmock.Sqlmock) {
		// Create
		mock.ExpectExec(
			`CREATE CONNECTION "database"."schema"."conn" TO SQL SERVER \(HOST 'sqlserver_host', PORT 1433, USER SECRET "materialize"."public"."user", PASSWORD SECRET "materialize"."public"."password", SSH TUNNEL "tunnel_database"."tunnel_schema"."ssh_conn", AWS PRIVATELINK "aws_database"."aws_schema"."aws_conn", DATABASE 'testdb'\);`,
		).WillReturnResult(sqlmock.NewResult(1, 1))

		// Comment
		mock.ExpectExec(`COMMENT ON CONNECTION "database"."schema"."conn" IS 'SQL Server connection comment';`).WillReturnResult(sqlmock.NewResult(1, 1))

		// Query Id
		ip := `WHERE mz_connections.name = 'conn' AND mz_databases.name = 'database' AND mz_schemas.name = 'schema'`
		testhelpers.MockConnectionScan(mock, ip)

		// Query Params
		pp := `WHERE mz_connections.id = 'u1'`
		testhelpers.MockConnectionScan(mock, pp)

		if err := connectionSQLServerCreate(context.TODO(), d, db); err != nil {
			t.Fatal(err)
		}
	})
}

func TestResourceConnectionSQLServerCreateMinimal(t *testing.T) {
	r := require.New(t)
	minimalInput := map[string]interface{}{
		"name":          "minimal_conn",
		"schema_name":   "schema",
		"database_name": "database",
		"host":          "sqlserver_host",
		"database":      "testdb",
		"user":          []interface{}{map[string]interface{}{"text": "plaintext_user"}},
	}
	d := schema.TestResourceDataRaw(t, ConnectionSQLServer().Schema, minimalInput)
	r.NotNil(d)

	testhelpers.WithMockProviderMeta(t, func(db *utils.ProviderMeta, mock sqlmock.Sqlmock) {
		// Create with minimal configuration (should use default port 1433)
		mock.ExpectExec(
			`CREATE CONNECTION "database"."schema"."minimal_conn" TO SQL SERVER \(HOST 'sqlserver_host', PORT 1433, USER 'plaintext_user', DATABASE 'testdb'\);`,
		).WillReturnResult(sqlmock.NewResult(1, 1))

		// Query Id
		ip := `WHERE mz_connections.name = 'minimal_conn' AND mz_databases.name = 'database' AND mz_schemas.name = 'schema'`
		testhelpers.MockConnectionScan(mock, ip)

		// Query Params
		pp := `WHERE mz_connections.id = 'u1'`
		testhelpers.MockConnectionScan(mock, pp)

		if err := connectionSQLServerCreate(context.TODO(), d, db); err != nil {
			t.Fatal(err)
		}
	})
}

func TestResourceConnectionSQLServerCreateWithoutValidation(t *testing.T) {
	r := require.New(t)
	noValidateInput := map[string]interface{}{
		"name":          "no_validate_conn",
		"schema_name":   "schema",
		"database_name": "database",
		"host":          "sqlserver_host",
		"port":          1433,
		"database":      "testdb",
		"user":          []interface{}{map[string]interface{}{"text": "user"}},
		"password":      []interface{}{map[string]interface{}{"name": "password"}},
		"validate":      false,
	}
	d := schema.TestResourceDataRaw(t, ConnectionSQLServer().Schema, noValidateInput)
	r.NotNil(d)

	testhelpers.WithMockProviderMeta(t, func(db *utils.ProviderMeta, mock sqlmock.Sqlmock) {
		// Create without validation
		mock.ExpectExec(
			`CREATE CONNECTION "database"."schema"."no_validate_conn" TO SQL SERVER \(HOST 'sqlserver_host', PORT 1433, USER 'user', PASSWORD SECRET "materialize"."public"."password", DATABASE 'testdb'\) WITH \(VALIDATE = false\);`,
		).WillReturnResult(sqlmock.NewResult(1, 1))

		// Query Id
		ip := `WHERE mz_connections.name = 'no_validate_conn' AND mz_databases.name = 'database' AND mz_schemas.name = 'schema'`
		testhelpers.MockConnectionScan(mock, ip)

		// Query Params
		pp := `WHERE mz_connections.id = 'u1'`
		testhelpers.MockConnectionScan(mock, pp)

		if err := connectionSQLServerCreate(context.TODO(), d, db); err != nil {
			t.Fatal(err)
		}
	})
}

func TestResourceConnectionSQLServerUpdate(t *testing.T) {
	r := require.New(t)
	d := schema.TestResourceDataRaw(t, ConnectionSQLServer().Schema, inSQLServer)

	// Set current state
	d.SetId("u1")
	d.Set("name", "conn")
	r.NotNil(d)

	testhelpers.WithMockProviderMeta(t, func(db *utils.ProviderMeta, mock sqlmock.Sqlmock) {
		// Name Change
		mock.ExpectExec(`ALTER CONNECTION "database"."schema"."" RENAME TO "conn";`).WillReturnResult(sqlmock.NewResult(1, 1))

		// Host
		mock.ExpectExec(`ALTER CONNECTION "database"."schema"."conn" SET \(HOST = 'sqlserver_host'\);`).WillReturnResult(sqlmock.NewResult(1, 1))

		// Port
		mock.ExpectExec(`ALTER CONNECTION "database"."schema"."conn" SET \(PORT = 1433\);`).WillReturnResult(sqlmock.NewResult(1, 1))

		// User
		mock.ExpectExec(`ALTER CONNECTION "database"."schema"."conn" SET \(USER = SECRET "materialize"."public"."user"\);`).WillReturnResult(sqlmock.NewResult(1, 1))

		// Password
		mock.ExpectExec(`ALTER CONNECTION "database"."schema"."conn" SET \(PASSWORD = SECRET "materialize"."public"."password"\);`).WillReturnResult(sqlmock.NewResult(1, 1))

		// Database
		mock.ExpectExec(`ALTER CONNECTION "database"."schema"."conn" SET \(DATABASE = 'testdb'\);`).WillReturnResult(sqlmock.NewResult(1, 1))

		// SSH Tunnel
		mock.ExpectExec(`ALTER CONNECTION "database"."schema"."conn" SET \(SSH TUNNEL = "tunnel_database"."tunnel_schema"."ssh_conn"\);`).WillReturnResult(sqlmock.NewResult(1, 1))

		// AWS PrivateLink
		mock.ExpectExec(`ALTER CONNECTION "database"."schema"."conn" SET \(AWS PRIVATELINK = "aws_database"."aws_schema"."aws_conn"\);`).WillReturnResult(sqlmock.NewResult(1, 1))

		// Comment
		mock.ExpectExec(`COMMENT ON CONNECTION "database"."schema"."conn" IS 'SQL Server connection comment';`).WillReturnResult(sqlmock.NewResult(1, 1))

		// Query Params
		p := `WHERE mz_connections.id = 'u1'`
		testhelpers.MockConnectionScan(mock, p)

		// Execute the update function
		if err := connectionUpdate(context.TODO(), d, db); err != nil {
			t.Fatal(err)
		}
	})
}

func TestResourceConnectionSQLServerCreateWithSSL(t *testing.T) {
	sslInput := map[string]interface{}{
		"name":          "ssl_conn",
		"schema_name":   "schema",
		"database_name": "database",
		"host":          "sqlserver_host",
		"port":          1433,
		"database":      "testdb",
		"user":          []interface{}{map[string]interface{}{"secret": []interface{}{map[string]interface{}{"name": "user"}}}},
		"password":      []interface{}{map[string]interface{}{"name": "password"}},
		"ssl_mode":      "require",
		"ssl_certificate_authority": []interface{}{
			map[string]interface{}{
				"text": "-----BEGIN CERTIFICATE-----",
			},
		},
		"validate": true,
		"comment":  "SSL connection comment",
	}

	r := require.New(t)
	d := schema.TestResourceDataRaw(t, ConnectionSQLServer().Schema, sslInput)
	r.NotNil(d)

	testhelpers.WithMockProviderMeta(t, func(db *utils.ProviderMeta, mock sqlmock.Sqlmock) {
		// Create with SSL
		mock.ExpectExec(
			`CREATE CONNECTION "database"."schema"."ssl_conn" TO SQL SERVER \(HOST 'sqlserver_host', PORT 1433, USER SECRET "materialize"."public"."user", PASSWORD SECRET "materialize"."public"."password", SSL MODE 'require', SSL CERTIFICATE AUTHORITY '-----BEGIN CERTIFICATE-----', DATABASE 'testdb'\);`,
		).WillReturnResult(sqlmock.NewResult(1, 1))

		// Comment
		mock.ExpectExec(`COMMENT ON CONNECTION "database"."schema"."ssl_conn" IS 'SSL connection comment';`).WillReturnResult(sqlmock.NewResult(1, 1))

		// Query Id
		ip := `WHERE mz_connections.name = 'ssl_conn' AND mz_databases.name = 'database' AND mz_schemas.name = 'schema'`
		testhelpers.MockConnectionScan(mock, ip)

		// Query Params
		pp := `WHERE mz_connections.id = 'u1'`
		testhelpers.MockConnectionScan(mock, pp)

		if err := connectionSQLServerCreate(context.TODO(), d, db); err != nil {
			t.Fatal(err)
		}
	})
}

func TestResourceConnectionSQLServerCreateWithSSLSecret(t *testing.T) {
	sslSecretInput := map[string]interface{}{
		"name":          "ssl_secret_conn",
		"schema_name":   "schema",
		"database_name": "database",
		"host":          "sqlserver_host",
		"port":          1433,
		"database":      "testdb",
		"user":          []interface{}{map[string]interface{}{"secret": []interface{}{map[string]interface{}{"name": "user"}}}},
		"password":      []interface{}{map[string]interface{}{"name": "password"}},
		"ssl_mode":      "verify-ca",
		"ssl_certificate_authority": []interface{}{
			map[string]interface{}{
				"secret": []interface{}{map[string]interface{}{
					"name":          "ssl_ca_secret",
					"schema_name":   "schema",
					"database_name": "database",
				}},
			},
		},
		"validate": true,
		"comment":  "SSL secret connection comment",
	}

	r := require.New(t)
	d := schema.TestResourceDataRaw(t, ConnectionSQLServer().Schema, sslSecretInput)
	r.NotNil(d)

	testhelpers.WithMockProviderMeta(t, func(db *utils.ProviderMeta, mock sqlmock.Sqlmock) {
		// Create with SSL secret
		mock.ExpectExec(
			`CREATE CONNECTION "database"."schema"."ssl_secret_conn" TO SQL SERVER \(HOST 'sqlserver_host', PORT 1433, USER SECRET "materialize"."public"."user", PASSWORD SECRET "materialize"."public"."password", SSL MODE 'verify-ca', SSL CERTIFICATE AUTHORITY SECRET "database"."schema"."ssl_ca_secret", DATABASE 'testdb'\);`,
		).WillReturnResult(sqlmock.NewResult(1, 1))

		// Comment
		mock.ExpectExec(`COMMENT ON CONNECTION "database"."schema"."ssl_secret_conn" IS 'SSL secret connection comment';`).WillReturnResult(sqlmock.NewResult(1, 1))

		// Query Id
		ip := `WHERE mz_connections.name = 'ssl_secret_conn' AND mz_databases.name = 'database' AND mz_schemas.name = 'schema'`
		testhelpers.MockConnectionScan(mock, ip)

		// Query Params
		pp := `WHERE mz_connections.id = 'u1'`
		testhelpers.MockConnectionScan(mock, pp)

		if err := connectionSQLServerCreate(context.TODO(), d, db); err != nil {
			t.Fatal(err)
		}
	})
}

func TestResourceConnectionSQLServerRead(t *testing.T) {
	r := require.New(t)
	d := schema.TestResourceDataRaw(t, ConnectionSQLServer().Schema, inSQLServer)
	d.SetId("aws/us-east-1:u1")

	testhelpers.WithMockProviderMeta(t, func(db *utils.ProviderMeta, mock sqlmock.Sqlmock) {
		// Query Params
		p := `WHERE mz_connections.id = 'u1'`
		testhelpers.MockConnectionScan(mock, p)

		if err := connectionRead(context.TODO(), d, db); err != nil {
			t.Fatal(err)
		}

		r.Equal("connection", d.Get("name"))
		r.Equal("schema", d.Get("schema_name"))
		r.Equal("database", d.Get("database_name"))
	})
}

func TestResourceConnectionSQLServerDelete(t *testing.T) {
	r := require.New(t)
	d := schema.TestResourceDataRaw(t, ConnectionSQLServer().Schema, inSQLServer)
	d.SetId("u1")
	r.NotNil(d)

	testhelpers.WithMockProviderMeta(t, func(db *utils.ProviderMeta, mock sqlmock.Sqlmock) {
		mock.ExpectExec(`DROP CONNECTION "database"."schema"."conn";`).WillReturnResult(sqlmock.NewResult(1, 1))

		if err := connectionDelete(context.TODO(), d, db); err != nil {
			t.Fatal(err)
		}
	})
}
