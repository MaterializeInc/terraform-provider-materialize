package resources

import (
	"context"
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/require"
)

func TestResourceAwsPrivatelinkCreate(t *testing.T) {
	r := require.New(t)

	in := map[string]interface{}{
		"name":               "conn",
		"schema_name":        "schema",
		"database_name":      "database",
		"service_name":       "service",
		"availability_zones": []interface{}{"use1-az1", "use1-az2"},
	}
	d := schema.TestResourceDataRaw(t, ConnectionAwsPrivatelink().Schema, in)
	r.NotNil(d)

	WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		// Create
		mock.ExpectExec(
			`CREATE CONNECTION database.schema.conn TO AWS PRIVATELINK \(SERVICE NAME 'service',AVAILABILITY ZONES \('use1-az1', 'use1-az2'\)\)`,
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

		if err := connectionAwsPrivatelinkCreate(context.TODO(), d, db); err != nil {
			t.Fatal(err)
		}
	})

}

func TestResourceAwsPrivatelinkDelete(t *testing.T) {
	r := require.New(t)

	in := map[string]interface{}{
		"name":          "conn",
		"schema_name":   "schema",
		"database_name": "database",
	}
	d := schema.TestResourceDataRaw(t, ConnectionAwsPrivatelink().Schema, in)
	r.NotNil(d)

	WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		mock.ExpectExec(`DROP CONNECTION database.schema.conn;`).WillReturnResult(sqlmock.NewResult(1, 1))

		if err := connectionAwsPrivatelinkDelete(context.TODO(), d, db); err != nil {
			t.Fatal(err)
		}
	})

}

func TestConnectionAwsPrivatelinkReadIdQuery(t *testing.T) {
	r := require.New(t)
	b := newConnectionAwsPrivatelinkBuilder("connection", "schema", "database")
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

func TestConnectionAwsPrivatelinkRenameQuery(t *testing.T) {
	r := require.New(t)
	b := newConnectionAwsPrivatelinkBuilder("connection", "schema", "database")
	r.Equal(`ALTER CONNECTION database.schema.connection RENAME TO database.schema.new_connection;`, b.Rename("new_connection"))
}

func TestConnectionAwsPrivatelinkDropQuery(t *testing.T) {
	r := require.New(t)
	b := newConnectionAwsPrivatelinkBuilder("connection", "schema", "database")
	r.Equal(`DROP CONNECTION database.schema.connection;`, b.Drop())
}

func TestConnectionAwsPrivatelinkReadParamsQuery(t *testing.T) {
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

func TestConnectionCreateAwsPrivateLinkQuery(t *testing.T) {
	r := require.New(t)

	b := newConnectionAwsPrivatelinkBuilder("privatelink_conn", "schema", "database")
	b.PrivateLinkServiceName("com.amazonaws.us-east-1.materialize.example")
	b.PrivateLinkAvailabilityZones([]string{"use1-az1", "use1-az2"})
	r.Equal(`CREATE CONNECTION database.schema.privatelink_conn TO AWS PRIVATELINK (SERVICE NAME 'com.amazonaws.us-east-1.materialize.example',AVAILABILITY ZONES ('use1-az1', 'use1-az2'));`, b.Create())
}
