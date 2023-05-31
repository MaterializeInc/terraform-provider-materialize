package datasources

import (
	"context"
	"testing"

	"github.com/MaterializeInc/terraform-provider-materialize/pkg/testhelpers"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/require"
)

func TestTypesDatasource(t *testing.T) {
	r := require.New(t)

	in := map[string]interface{}{
		"schema_name":   "schema",
		"database_name": "database",
	}
	d := schema.TestResourceDataRaw(t, Type().Schema, in)
	r.NotNil(d)

	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		ir := mock.NewRows([]string{"id", "name", "schema_name", "database_name", "category"}).
			AddRow("u1", "type", "schema", "database", "category")
		mock.ExpectQuery(`
			SELECT
				mz_types.id,
				mz_types.name,
				mz_schemas.name AS schema_name,
				mz_databases.name AS database_name,
				mz_types.category
			FROM mz_types
			JOIN mz_schemas
				ON mz_types.schema_id = mz_schemas.id
			JOIN mz_databases
				ON mz_schemas.database_id = mz_databases.id
			WHERE mz_databases.name = 'database'
			AND mz_schemas.name = 'schema';`).WillReturnRows(ir)

		if err := typeRead(context.TODO(), d, db); err != nil {
			t.Fatal(err)
		}
	})

}
