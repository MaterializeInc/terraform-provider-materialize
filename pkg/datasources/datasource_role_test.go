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

func TestRoleDatasource(t *testing.T) {
	r := require.New(t)

	in := map[string]interface{}{}
	d := schema.TestResourceDataRaw(t, Role().Schema, in)
	r.NotNil(d)

	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		ir := mock.NewRows([]string{"id", "role_name", "inherit", "create_role", "create_db", "create_cluster"}).
			AddRow("u1", "role", true, true, true, true)
		mock.ExpectQuery(`
			SELECT
				id,
				name AS role_name,
				inherit,
				create_role,
				create_db,
				create_cluster
			FROM mz_roles;`).WillReturnRows(ir)

		if err := roleRead(context.TODO(), d, db); err != nil {
			t.Fatal(err)
		}
	})

}
