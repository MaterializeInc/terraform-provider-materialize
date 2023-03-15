package resources

import (
	"context"
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/require"
)

func TestResourcePostgresCreate(t *testing.T) {
	r := require.New(t)

	in := map[string]interface{}{
		"name":                      "conn",
		"schema_name":               "schema",
		"database_name":             "database",
		"database":                  "default",
		"host":                      "postgres_host",
		"port":                      5432,
		"user":                      []interface{}{map[string]interface{}{"secret": []interface{}{map[string]interface{}{"name": "user"}}}},
		"password":                  []interface{}{map[string]interface{}{"name": "password"}},
		"ssh_tunnel":                []interface{}{map[string]interface{}{"name": "ssh_conn"}},
		"ssl_certificate_authority": []interface{}{map[string]interface{}{"secret": []interface{}{map[string]interface{}{"name": "root"}}}},
		"ssl_certificate":           []interface{}{map[string]interface{}{"secret": []interface{}{map[string]interface{}{"name": "cert"}}}},
		"ssl_key":                   []interface{}{map[string]interface{}{"name": "key"}},
		"ssl_mode":                  "verify-full",
		"aws_privatelink":           []interface{}{map[string]interface{}{"name": "link"}},
	}
	d := schema.TestResourceDataRaw(t, ConnectionPostgres().Schema, in)
	r.NotNil(d)

	WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		// Create
		mock.ExpectExec(
			`CREATE CONNECTION "database"."schema"."conn" TO POSTGRES \(HOST 'postgres_host', PORT 5432, USER SECRET "database"."schema"."user", PASSWORD SECRET "database"."schema"."password", SSL MODE 'verify-full', SSH TUNNEL "database"."schema"."ssh_conn", SSL CERTIFICATE AUTHORITY SECRET "database"."schema"."root", SSL CERTIFICATE SECRET "database"."schema"."cert", SSL KEY SECRET "database"."schema"."key", AWS PRIVATELINK "database"."schema"."link", DATABASE 'default'\);`,
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

		if err := connectionPostgresCreate(context.TODO(), d, db); err != nil {
			t.Fatal(err)
		}
	})

}

func TestResourcePostgresDelete(t *testing.T) {
	r := require.New(t)

	in := map[string]interface{}{
		"name":          "conn",
		"schema_name":   "schema",
		"database_name": "database",
	}
	d := schema.TestResourceDataRaw(t, ConnectionPostgres().Schema, in)
	r.NotNil(d)

	WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		mock.ExpectExec(`DROP CONNECTION "database"."schema"."conn";`).WillReturnResult(sqlmock.NewResult(1, 1))

		if err := connectionPostgresDelete(context.TODO(), d, db); err != nil {
			t.Fatal(err)
		}
	})

}

func TestConnectoinPostgresReadIdQuery(t *testing.T) {
	r := require.New(t)
	b := newConnectionKafkaBuilder("connection", "schema", "database")
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

func TestConnectionPostgresRenameQuery(t *testing.T) {
	r := require.New(t)
	b := newConnectionPostgresBuilder("connection", "schema", "database")
	r.Equal(`ALTER CONNECTION "database"."schema"."connection" RENAME TO "database"."schema"."new_connection";`, b.Rename("new_connection"))
}

func TestConnectionPostgresDropQuery(t *testing.T) {
	r := require.New(t)
	b := newConnectionPostgresBuilder("connection", "schema", "database")
	r.Equal(`DROP CONNECTION "database"."schema"."connection";`, b.Drop())
}

func TestConnectionPostgresReadParamsQuery(t *testing.T) {
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

func TestConnectionPostgresCreateQuery(t *testing.T) {
	r := require.New(t)
	b := newConnectionPostgresBuilder("postgres_conn", "schema", "database")
	b.PostgresHost("postgres_host")
	b.PostgresPort(5432)
	b.PostgresUser(ValueSecretStruct{Text: "user"})
	b.PostgresPassword(IdentifierSchemaStruct{Name: "password", SchemaName: "schema", DatabaseName: "database"})
	b.PostgresDatabase("default")
	r.Equal(`CREATE CONNECTION "database"."schema"."postgres_conn" TO POSTGRES (HOST 'postgres_host', PORT 5432, USER 'user', PASSWORD SECRET "database"."schema"."password", DATABASE 'default');`, b.Create())
}

func TestConnectionPostgresCreateSshQuery(t *testing.T) {
	r := require.New(t)
	b := newConnectionPostgresBuilder("postgres_conn", "schema", "database")
	b.PostgresHost("postgres_host")
	b.PostgresPort(5432)
	b.PostgresUser(ValueSecretStruct{Text: "user"})
	b.PostgresPassword(IdentifierSchemaStruct{Name: "password", SchemaName: "schema", DatabaseName: "database"})
	b.PostgresDatabase("default")
	b.PostgresSSHTunnel(IdentifierSchemaStruct{Name: "ssh_conn", SchemaName: "schema", DatabaseName: "database"})
	r.Equal(`CREATE CONNECTION "database"."schema"."postgres_conn" TO POSTGRES (HOST 'postgres_host', PORT 5432, USER 'user', PASSWORD SECRET "database"."schema"."password", SSH TUNNEL "database"."schema"."ssh_conn", DATABASE 'default');`, b.Create())
}

func TestConnectionPostgresCreatePrivateLinkQuery(t *testing.T) {
	r := require.New(t)
	b := newConnectionPostgresBuilder("postgres_conn", "schema", "database")
	b.PostgresHost("postgres_host")
	b.PostgresPort(5432)
	b.PostgresUser(ValueSecretStruct{Text: "user"})
	b.PostgresPassword(IdentifierSchemaStruct{Name: "password", SchemaName: "schema", DatabaseName: "database"})
	b.PostgresDatabase("default")
	b.PostgresAWSPrivateLink(IdentifierSchemaStruct{Name: "private_link", SchemaName: "schema", DatabaseName: "database"})
	r.Equal(`CREATE CONNECTION "database"."schema"."postgres_conn" TO POSTGRES (HOST 'postgres_host', PORT 5432, USER 'user', PASSWORD SECRET "database"."schema"."password", AWS PRIVATELINK "database"."schema"."private_link", DATABASE 'default');`, b.Create())
}

func TestConnectionPostgresCreateSslQuery(t *testing.T) {
	r := require.New(t)
	b := newConnectionPostgresBuilder("postgres_conn", "schema", "database")
	b.PostgresHost("postgres_host")
	b.PostgresPort(5432)
	b.PostgresUser(ValueSecretStruct{Secret: IdentifierSchemaStruct{Name: "user", SchemaName: "schema", DatabaseName: "database"}})
	b.PostgresPassword(IdentifierSchemaStruct{Name: "password", SchemaName: "schema", DatabaseName: "database"})
	b.PostgresDatabase("default")
	b.PostgresSSLMode("verify-full")
	b.PostgresSSLCa(ValueSecretStruct{Secret: IdentifierSchemaStruct{Name: "root", SchemaName: "schema", DatabaseName: "database"}})
	b.PostgresSSLCert(ValueSecretStruct{Secret: IdentifierSchemaStruct{Name: "cert", SchemaName: "schema", DatabaseName: "database"}})
	b.PostgresSSLKey(IdentifierSchemaStruct{Name: "key", SchemaName: "schema", DatabaseName: "database"})
	r.Equal(`CREATE CONNECTION "database"."schema"."postgres_conn" TO POSTGRES (HOST 'postgres_host', PORT 5432, USER SECRET "database"."schema"."user", PASSWORD SECRET "database"."schema"."password", SSL MODE 'verify-full', SSL CERTIFICATE AUTHORITY SECRET "database"."schema"."root", SSL CERTIFICATE SECRET "database"."schema"."cert", SSL KEY SECRET "database"."schema"."key", DATABASE 'default');`, b.Create())
}
