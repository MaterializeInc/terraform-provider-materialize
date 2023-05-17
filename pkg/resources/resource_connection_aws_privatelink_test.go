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

var inAwsPrivatelink = map[string]interface{}{
	"name":               "conn",
	"schema_name":        "schema",
	"database_name":      "database",
	"service_name":       "service",
	"availability_zones": []interface{}{"use1-az1", "use1-az2"},
}

func TestResourceAwsPrivatelinkRead(t *testing.T) {
	r := require.New(t)
	d := schema.TestResourceDataRaw(t, ConnectionAwsPrivatelink().Schema, inAwsPrivatelink)
	r.NotNil(d)
	d.SetId("u1")

	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		mockConnectionAwsPrivatelinkParams(mock)

		if err := connectionAwsPrivatelinkRead(context.TODO(), d, db); err != nil {
			t.Fatal(err)
		}

		// Ensure parameters set
		expectedParams := map[string]string{
			"name":               "conn",
			"schema_name":        "schema",
			"database_name":      "database",
			"principal":          "principal",
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

func TestResourceAwsPrivatelinkCreate(t *testing.T) {
	r := require.New(t)
	d := schema.TestResourceDataRaw(t, ConnectionAwsPrivatelink().Schema, inAwsPrivatelink)
	r.NotNil(d)

	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		// Create
		mock.ExpectExec(
			`CREATE CONNECTION "database"."schema"."conn" TO AWS PRIVATELINK \(SERVICE NAME 'service',AVAILABILITY ZONES \('use1-az1', 'use1-az2'\)\)`,
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
		mockConnectionAwsPrivatelinkParams(mock)

		if err := connectionAwsPrivatelinkCreate(context.TODO(), d, db); err != nil {
			t.Fatal(err)
		}
	})

}

func TestResourceAwsPrivatelinkUpdate(t *testing.T) {
	r := require.New(t)
	d := schema.TestResourceDataRaw(t, ConnectionAwsPrivatelink().Schema, inAwsPrivatelink)

	// Set current state
	d.SetId("u1")
	d.Set("name", "old_conn")
	r.NotNil(d)

	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		mock.ExpectExec(`ALTER CONNECTION "database"."schema"."old_conn" RENAME TO "database"."schema"."conn";`).WillReturnResult(sqlmock.NewResult(1, 1))

		// Query Params
		mockConnectionAwsPrivatelinkParams(mock)

		if err := connectionAwsPrivatelinkUpdate(context.TODO(), d, db); err != nil {
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

	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		mock.ExpectExec(`DROP CONNECTION "database"."schema"."conn";`).WillReturnResult(sqlmock.NewResult(1, 1))

		if err := connectionDelete(context.TODO(), d, db); err != nil {
			t.Fatal(err)
		}
	})

}
