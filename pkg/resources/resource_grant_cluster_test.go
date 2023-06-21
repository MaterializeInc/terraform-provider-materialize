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

func TestResourceGrantClusterCreate(t *testing.T) {
	r := require.New(t)

	in := map[string]interface{}{
		"role_name":    "joe",
		"privilege":    "CREATE",
		"cluster_name": "materialize",
	}
	d := schema.TestResourceDataRaw(t, GrantCluster().Schema, in)
	r.NotNil(d)

	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		// Create
		mock.ExpectExec(
			`GRANT CREATE ON CLUSTER "materialize" TO joe;`,
		).WillReturnResult(sqlmock.NewResult(1, 1))

		// Query Role Id
		rp := `WHERE mz_roles.name = 'joe'`
		testhelpers.MockRoleScan(mock, rp)

		// Query Grant Id
		gp := `WHERE mz_clusters.name = 'materialize'`
		testhelpers.MockClusterScan(mock, gp)

		// Query Params
		pp := `WHERE mz_clusters.id = 'u1'`
		testhelpers.MockClusterScan(mock, pp)

		if err := grantClusterCreate(context.TODO(), d, db); err != nil {
			t.Fatal(err)
		}

		if d.Id() != "GRANT|CLUSTER|u1|u1|CREATE" {
			t.Fatalf("unexpected id of %s", d.Id())
		}
	})
}

func TestResourceGrantClusterDelete(t *testing.T) {
	r := require.New(t)

	in := map[string]interface{}{
		"role_name":    "joe",
		"privilege":    "CREATE",
		"cluster_name": "materialize",
	}
	d := schema.TestResourceDataRaw(t, GrantCluster().Schema, in)
	r.NotNil(d)

	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		mock.ExpectExec(`REVOKE CREATE ON CLUSTER "materialize" FROM joe;`).WillReturnResult(sqlmock.NewResult(1, 1))

		if err := grantClusterDelete(context.TODO(), d, db); err != nil {
			t.Fatal(err)
		}
	})
}
