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

var inMySQL = map[string]interface{}{
	"name":          "conn",
	"schema_name":   "schema",
	"database_name": "database",
	"host":          "mysql_host",
	"port":          3306,
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
	"ssl_mode":        "verify-ca",
	"comment":         "object comment",
}

func TestResourceConnectionMySQLCreate(t *testing.T) {
	r := require.New(t)
	d := schema.TestResourceDataRaw(t, ConnectionMySQL().Schema, inMySQL)
	r.NotNil(d)

	testhelpers.WithMockProviderMeta(t, func(db *utils.ProviderMeta, mock sqlmock.Sqlmock) {
		// Create
		mock.ExpectExec(
			`CREATE CONNECTION "database"."schema"."conn" TO MYSQL \(HOST 'mysql_host', PORT 3306, USER SECRET "materialize"."public"."user", PASSWORD SECRET "materialize"."public"."password", SSL MODE 'verify-ca', SSH TUNNEL "tunnel_database"."tunnel_schema"."ssh_conn", SSL CERTIFICATE AUTHORITY SECRET "ssl_database"."public"."root", SSL CERTIFICATE SECRET "materialize"."public"."cert", SSL KEY SECRET "materialize"."public"."key"\);`,
		).WillReturnResult(sqlmock.NewResult(1, 1))

		// Comment
		mock.ExpectExec(`COMMENT ON CONNECTION "database"."schema"."conn" IS 'object comment';`).WillReturnResult(sqlmock.NewResult(1, 1))

		// Query Id
		ip := `WHERE mz_connections.name = 'conn' AND mz_databases.name = 'database' AND mz_schemas.name = 'schema'`
		testhelpers.MockConnectionScan(mock, ip)

		// Query Params
		pp := `WHERE mz_connections.id = 'u1'`
		testhelpers.MockConnectionScan(mock, pp)

		if err := connectionMySQLCreate(context.TODO(), d, db); err != nil {
			t.Fatal(err)
		}
	})
}
