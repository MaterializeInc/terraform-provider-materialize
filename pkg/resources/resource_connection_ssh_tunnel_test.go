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

		// Query Id
		ir := mock.NewRows([]string{"id"}).AddRow("u1")
		mock.ExpectQuery(`
			SELECT
				mz_connections.id,
				mz_connections.name AS connection_name,
				mz_schemas.name AS schema_name,
				mz_databases.name AS database_name,
				mz_connections.type AS connection_type
			FROM mz_connections
			JOIN mz_schemas
				ON mz_connections.schema_id = mz_schemas.id
			JOIN mz_databases
				ON mz_schemas.database_id = mz_databases.id
			WHERE mz_connections.name = 'conn'
			AND mz_databases.name = 'database'
			AND mz_schemas.name = 'schema';`).WillReturnRows(ir)

		// Query Params
		ip := sqlmock.NewRows([]string{"connection_name", "schema_name", "database_name", "public_key_1", "public_key_2"}).
			AddRow("conn", "schema", "database", "pk1", "pk2")
		mock.ExpectQuery(readConnectionSshTunnellink).WillReturnRows(ip)

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
		mock.ExpectExec(`ALTER CONNECTION "database"."schema"."old_conn" RENAME TO "conn";`).WillReturnResult(sqlmock.NewResult(1, 1))

		// Query Params
		ip := sqlmock.NewRows([]string{"connection_name", "schema_name", "database_name", "public_key_1", "public_key_2"}).
			AddRow("conn", "schema", "database", "pk1", "pk2")
		mock.ExpectQuery(readConnectionSshTunnellink).WillReturnRows(ip)

		if err := connectionSshTunnelUpdate(context.TODO(), d, db); err != nil {
			t.Fatal(err)
		}
	})

}
