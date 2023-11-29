package resources

import (
	"context"
	"testing"

	"github.com/MaterializeInc/terraform-provider-materialize/pkg/testhelpers"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/require"
)

var inSshTunnel = map[string]interface{}{
	"name":          "conn",
	"schema_name":   "schema",
	"database_name": "database",
	"database":      "default",
	"host":          "localhost",
	"port":          123,
	"user":          "user",
	"comment":       "object comment",
}

func TestResourceConnectionSshTunnelCreate(t *testing.T) {
	r := require.New(t)

	d := schema.TestResourceDataRaw(t, ConnectionSshTunnel().Schema, inSshTunnel)
	r.NotNil(d)

	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		// Create
		mock.ExpectExec(
			`CREATE CONNECTION "database"."schema"."conn" TO SSH TUNNEL \(HOST 'localhost', USER 'user', PORT 123\);`,
		).WillReturnResult(sqlmock.NewResult(1, 1))

		// Comment
		mock.ExpectExec(`COMMENT ON CONNECTION "database"."schema"."conn" IS 'object comment';`).WillReturnResult(sqlmock.NewResult(1, 1))

		// Query Id
		ip := `WHERE mz_connections.name = 'conn' AND mz_databases.name = 'database' AND mz_schemas.name = 'schema'`
		testhelpers.MockConnectionScan(mock, ip)

		// Query Params
		pp := `WHERE mz_connections.id = 'u1'`
		testhelpers.MockConnectionSshTunnelScan(mock, pp)

		if err := connectionSshTunnelCreate(context.TODO(), d, db); err != nil {
			t.Fatal(err)
		}
	})
}

func TestResourceConnectionSshTunnelUpdate(t *testing.T) {
	r := require.New(t)
	d := schema.TestResourceDataRaw(t, ConnectionSshTunnel().Schema, inSshTunnel)

	// Set current state
	d.SetId("u1")
	d.Set("name", "old_conn")
	r.NotNil(d)

	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		mock.ExpectExec(`ALTER CONNECTION "database"."schema"."" RENAME TO "conn";`).WillReturnResult(sqlmock.NewResult(1, 1))

		// Comment
		mock.ExpectExec(`COMMENT ON CONNECTION "database"."schema"."old_conn" IS 'object comment';`).WillReturnResult(sqlmock.NewResult(1, 1))

		// Query Params
		pp := `WHERE mz_connections.id = 'u1'`
		testhelpers.MockConnectionSshTunnelScan(mock, pp)

		if err := connectionSshTunnelUpdate(context.TODO(), d, db); err != nil {
			t.Fatal(err)
		}
	})
}
