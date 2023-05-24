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
