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

func TestSchemaDatasource(t *testing.T) {
	r := require.New(t)

	in := map[string]interface{}{
		"database_name": "database",
	}
	d := schema.TestResourceDataRaw(t, Schema().Schema, in)
	r.NotNil(d)

	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		ir := mock.NewRows([]string{"id", "schema", "database"}).
			AddRow("u1", "schema", "database")
		mock.ExpectQuery(`
		SELECT
			mz_schemas.id,
			mz_schemas.name,
			mz_databases.name
		FROM mz_schemas JOIN mz_databases
			ON mz_schemas.database_id = mz_databases.id
		WHERE mz_databases.name = 'database`).WillReturnRows(ir)

		if err := schemaRead(context.TODO(), d, db); err != nil {
			t.Fatal(err)
		}
	})

}
