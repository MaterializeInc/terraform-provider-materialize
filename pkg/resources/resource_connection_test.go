package resources

import (
	"context"
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/MaterializeInc/terraform-provider-materialize/pkg/testhelpers"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/require"
)

var readConnection string = `
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
WHERE mz_connections.id = 'u1';`

var readConnectionAwsPrivatelink string = `
SELECT
	mz_connections.id,
	mz_connections.name AS connection_name,
	mz_schemas.name AS schema_name,
	mz_databases.name AS database_name,
	mz_aws_privatelink_connections.principal
FROM mz_connections
JOIN mz_schemas
	ON mz_connections.schema_id = mz_schemas.id
JOIN mz_databases
	ON mz_schemas.database_id = mz_databases.id
LEFT JOIN mz_aws_privatelink_connections
	ON mz_connections.id = mz_aws_privatelink_connections.id
WHERE mz_connections.id = 'u1';`

var readConnectionSshTunnellink string = `
SELECT
	mz_connections.id,
	mz_connections.name AS connection_name,
	mz_schemas.name AS schema_name,
	mz_databases.name AS database_name,
	mz_ssh_tunnel_connections.public_key_1,
	mz_ssh_tunnel_connections.public_key_2
FROM mz_connections
JOIN mz_schemas
	ON mz_connections.schema_id = mz_schemas.id
JOIN mz_databases
	ON mz_schemas.database_id = mz_databases.id
LEFT JOIN mz_ssh_tunnel_connections
	ON mz_connections.id = mz_ssh_tunnel_connections.id
WHERE mz_connections.id = 'u1';`

var inConnection = map[string]interface{}{
	"name":          "conn",
	"schema_name":   "schema",
	"database_name": "database",
}

func TestResourceConnectionUpdate(t *testing.T) {
	r := require.New(t)
	d := schema.TestResourceDataRaw(t, ConnectionKafka().Schema, inConnection)

	// Set current state
	d.SetId("u1")
	d.Set("name", "old_conn")
	r.NotNil(d)

	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		mock.ExpectExec(`ALTER CONNECTION "database"."schema"."old_conn" RENAME TO "database"."schema"."conn";`).WillReturnResult(sqlmock.NewResult(1, 1))

		// Query Params
		ip := sqlmock.NewRows([]string{"connection_name", "schema_name", "database_name"}).AddRow("conn", "schema", "database")
		mock.ExpectQuery(readConnection).WillReturnRows(ip)

		if err := connectionUpdate(context.TODO(), d, db); err != nil {
			t.Fatal(err)
		}
	})

}

func TestResourceConnectionDelete(t *testing.T) {
	r := require.New(t)

	d := schema.TestResourceDataRaw(t, ConnectionKafka().Schema, inConnection)
	r.NotNil(d)

	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		mock.ExpectExec(`DROP CONNECTION "database"."schema"."conn";`).WillReturnResult(sqlmock.NewResult(1, 1))

		if err := connectionDelete(context.TODO(), d, db); err != nil {
			t.Fatal(err)
		}
	})

}
