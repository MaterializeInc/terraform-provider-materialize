package resources

import (
	"context"
	"terraform-materialize/pkg/testhelpers"
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

	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
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
		ip := sqlmock.NewRows([]string{"name", "schema", "database"}).
			AddRow("conn", "schema", "database")
		mock.ExpectQuery(`
			SELECT
				mz_connections.name,
				mz_schemas.name,
				mz_databases.name
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

	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		mock.ExpectExec(`DROP CONNECTION "database"."schema"."conn";`).WillReturnResult(sqlmock.NewResult(1, 1))

		if err := connectionPostgresDelete(context.TODO(), d, db); err != nil {
			t.Fatal(err)
		}
	})

}
