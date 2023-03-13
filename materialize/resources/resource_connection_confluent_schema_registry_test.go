package resources

import (
	"context"
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/require"
)

func TestResourceConfluentSchemaRegistryCreate(t *testing.T) {
	r := require.New(t)

	in := map[string]interface{}{
		"name":                             "conn",
		"schema_name":                      "schema",
		"database_name":                    "database",
		"service_name":                     "service",
		"url":                              "http://localhost:8081",
		"ssl_certificate_authority_secret": "ssl",
		"ssl_certificate_secret":           "ssl",
		"ssl_key":                          "ssl",
		"password":                         "password",
		"username":                         "user",
		"ssh_tunnel":                       "tunnel",
		"aws_privatelink":                  "privatelink",
	}
	d := schema.TestResourceDataRaw(t, ConnectionConfluentSchemaRegistry().Schema, in)
	r.NotNil(d)

	WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		// Create
		mock.ExpectExec(
			`CREATE CONNECTION "database"."schema"."conn" TO CONFLUENT SCHEMA REGISTRY \(URL 'http://localhost:8081', USERNAME = 'user', PASSWORD = SECRET password, SSL CERTIFICATE AUTHORITY = SECRET ssl, SSL CERTIFICATE = SECRET ssl, SSL KEY = SECRET ssl, AWS PRIVATELINK privatelink, SSH TUNNEL tunnel\)`,
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

		if err := connectionConfluentSchemaRegistryCreate(context.TODO(), d, db); err != nil {
			t.Fatal(err)
		}
	})

}

func TestResourceConfluentSchemaRegistrykDelete(t *testing.T) {
	r := require.New(t)

	in := map[string]interface{}{
		"name":          "conn",
		"schema_name":   "schema",
		"database_name": "database",
	}
	d := schema.TestResourceDataRaw(t, ConnectionAwsPrivatelink().Schema, in)
	r.NotNil(d)

	WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		mock.ExpectExec(`DROP CONNECTION "database"."schema"."conn";`).WillReturnResult(sqlmock.NewResult(1, 1))

		if err := connectionConfluentSchemaRegistryDelete(context.TODO(), d, db); err != nil {
			t.Fatal(err)
		}
	})

}

func TestResourceConnectionConfluentSchemaRegistryReadId(t *testing.T) {
	r := require.New(t)
	b := newConnectionConfluentSchemaRegistryBuilder("connection", "schema", "database")
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

func TestConnectionConfluentSchemaRegistryRenameQuery(t *testing.T) {
	r := require.New(t)
	b := newConnectionConfluentSchemaRegistryBuilder("connection", "schema", "database")
	r.Equal(`ALTER CONNECTION "database"."schema"."connection" RENAME TO "database"."schema"."new_connection";`, b.Rename("new_connection"))
}

func TestConnectionConfluentSchemaRegistryDropQuery(t *testing.T) {
	r := require.New(t)
	b := newConnectionConfluentSchemaRegistryBuilder("connection", "schema", "database")
	r.Equal(`DROP CONNECTION "database"."schema"."connection";`, b.Drop())
}

func TestConnectionConfluentSchemaRegistryReadParamsQuery(t *testing.T) {
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

func TestConnectionCreateConfluentSchemaRegistryQuery(t *testing.T) {
	r := require.New(t)
	b := newConnectionConfluentSchemaRegistryBuilder("csr_conn", "schema", "database")
	b.ConfluentSchemaRegistryUrl("http://localhost:8081")
	b.ConfluentSchemaRegistryUsername("user")
	b.ConfluentSchemaRegistryPassword("password")
	r.Equal(`CREATE CONNECTION "database"."schema"."csr_conn" TO CONFLUENT SCHEMA REGISTRY (URL 'http://localhost:8081', USERNAME = 'user', PASSWORD = SECRET password);`, b.Create())
}

func TestConnectionCreateConfluentSchemaRegistryQueryUsernameSecret(t *testing.T) {
	r := require.New(t)
	b := newConnectionConfluentSchemaRegistryBuilder("csr_conn", "schema", "database")
	b.ConfluentSchemaRegistryUrl("http://localhost:8081")
	b.ConfluentSchemaRegistryUsernameSecret("user")
	b.ConfluentSchemaRegistryPassword("password")
	r.Equal(`CREATE CONNECTION "database"."schema"."csr_conn" TO CONFLUENT SCHEMA REGISTRY (URL 'http://localhost:8081', USERNAME = SECRET user, PASSWORD = SECRET password);`, b.Create())
}
