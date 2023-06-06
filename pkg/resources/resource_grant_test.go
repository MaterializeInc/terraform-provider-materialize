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

func TestResourceGrantCreate(t *testing.T) {
	r := require.New(t)

	in := map[string]interface{}{
		"role_name": "joe",
		"privilege": "CREATE",
		"object":    []interface{}{map[string]interface{}{"type": "DATABASE", "name": "materialize"}},
	}
	d := schema.TestResourceDataRaw(t, Grant().Schema, in)
	r.NotNil(d)

	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		// Create
		mock.ExpectExec(
			`GRANT CREATE ON DATABASE "materialize" TO joe;`,
		).WillReturnResult(sqlmock.NewResult(1, 1))

		// Query Role Id
		rp := `WHERE mz_roles.name = 'joe'`
		testhelpers.MockRoleScan(mock, rp)

		// Query Grant Id
		gp := `WHERE name = 'materialize'`
		testhelpers.MockDatabaseScan(mock, gp)

		// Query Params
		pp := `WHERE id = 'u1'`
		testhelpers.MockDatabaseScan(mock, pp)

		if err := grantCreate(context.TODO(), d, db); err != nil {
			t.Fatal(err)
		}

		if d.Id() != "GRANT|DATABASE|u1|u1|CREATE" {
			t.Fatalf("unexpected id of %s", d.Id())
		}
	})
}

func TestResourceGrantDelete(t *testing.T) {
	r := require.New(t)

	in := map[string]interface{}{
		"role_name": "joe",
		"privilege": "CREATE",
		"object":    []interface{}{map[string]interface{}{"type": "DATABASE", "name": "materialize"}},
	}
	d := schema.TestResourceDataRaw(t, Grant().Schema, in)
	r.NotNil(d)

	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		mock.ExpectExec(`REVOKE CREATE ON DATABASE "materialize" FROM joe;`).WillReturnResult(sqlmock.NewResult(1, 1))

		if err := grantDelete(context.TODO(), d, db); err != nil {
			t.Fatal(err)
		}
	})
}
