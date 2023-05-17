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

func TestResourceSshTunnelRead(t *testing.T) {
	r := require.New(t)
	d := schema.TestResourceDataRaw(t, ConnectionSshTunnel().Schema, inSshTunnel)
	r.NotNil(d)
	d.SetId("u1")

	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		mockConnectionSshTunnelParams(mock)

		if err := connectionSshTunnelRead(context.TODO(), d, db); err != nil {
			t.Fatal(err)
		}

		// Ensure parameters set
		expectedParams := map[string]string{
			"name":               "conn",
			"schema_name":        "schema",
			"database_name":      "database",
			"public_key_1":       "pk1",
			"public_key_2":       "pk2",
			"qualified_sql_name": `"database"."schema"."conn"`,
		}

		for key, value := range expectedParams {
			v := d.Get(key).(string)
			if v != value {
				t.Fatalf("Unexpected parameter set for %s. Recieved: %s", key, v)
			}
		}
	})
}

func TestResourceSshTunnelCreate(t *testing.T) {
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
		mockConnectionSshTunnelParams(mock)

		if err := connectionSshTunnelCreate(context.TODO(), d, db); err != nil {
			t.Fatal(err)
		}
	})

}

func TestResourceSshTunnelUpdate(t *testing.T) {
	r := require.New(t)
	d := schema.TestResourceDataRaw(t, ConnectionSshTunnel().Schema, inSshTunnel)

	// Set current state
	d.SetId("u1")
	d.Set("name", "old_conn")
	r.NotNil(d)

	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		mock.ExpectExec(`ALTER CONNECTION "database"."schema"."old_conn" RENAME TO "database"."schema"."conn";`).WillReturnResult(sqlmock.NewResult(1, 1))

		// Query Params
		mockConnectionSshTunnelParams(mock)

		if err := connectionSshTunnelUpdate(context.TODO(), d, db); err != nil {
			t.Fatal(err)
		}
	})

}

func TestResourceSshTunnelDelete(t *testing.T) {
	r := require.New(t)

	in := map[string]interface{}{
		"name":          "conn",
		"schema_name":   "schema",
		"database_name": "database",
	}
	d := schema.TestResourceDataRaw(t, ConnectionSshTunnel().Schema, in)
	r.NotNil(d)

	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		mock.ExpectExec(`DROP CONNECTION "database"."schema"."conn";`).WillReturnResult(sqlmock.NewResult(1, 1))

		if err := connectionDelete(context.TODO(), d, db); err != nil {
			t.Fatal(err)
		}
	})

}
