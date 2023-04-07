package datasources

import (
	"context"
	"testing"

	"github.com/MaterializeInc/terraform-materialize-provider/pkg/testhelpers"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/require"
)

func TestConnectionDatasource(t *testing.T) {
	r := require.New(t)

	in := map[string]interface{}{
		"schema_name":   "schema",
		"database_name": "database",
	}
	d := schema.TestResourceDataRaw(t, Connection().Schema, in)
	r.NotNil(d)

	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		ir := mock.NewRows([]string{"id", "name", "schema", "database", "connection_type"}).
			AddRow("u1", "secret", "schema", "database", "connection_type")
		mock.ExpectQuery(`
			SELECT
				mz_connections.id,
				mz_connections.name,
				mz_schemas.name AS schema_name,
				mz_databases.name AS database_name,
				mz_connections.type
			FROM mz_connections
			JOIN mz_schemas
				ON mz_connections.schema_id = mz_schemas.id
			JOIN mz_databases
				ON mz_schemas.database_id = mz_databases.id
			WHERE mz_databases.name = 'database' AND mz_schemas.name = 'schema'`).WillReturnRows(ir)

		if err := connectionRead(context.TODO(), d, db); err != nil {
			t.Fatal(err)
		}
	})

}
