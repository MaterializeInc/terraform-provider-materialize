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

var inCluster = map[string]interface{}{
	"name":           "cluster",
	"ownership_role": "joe",
}

func TestResourceClusterCreate(t *testing.T) {
	r := require.New(t)

	d := schema.TestResourceDataRaw(t, Cluster().Schema, inCluster)
	r.NotNil(d)

	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		// Create
		mock.ExpectExec(`CREATE CLUSTER "cluster" REPLICAS \(\);`).WillReturnResult(sqlmock.NewResult(1, 1))

		// Ownership
		mock.ExpectExec(`ALTER CLUSTER "cluster" OWNER TO "joe";`).WillReturnResult(sqlmock.NewResult(1, 1))

		// Query Id
		ip := `WHERE name = 'cluster'`
		testhelpers.MockClusterScan(mock, ip)

		// Query Params
		pp := `WHERE id = 'u1'`
		testhelpers.MockClusterScan(mock, pp)

		if err := clusterCreate(context.TODO(), d, db); err != nil {
			t.Fatal(err)
		}
	})
}

func TestResourceClusterDelete(t *testing.T) {
	r := require.New(t)

	d := schema.TestResourceDataRaw(t, Cluster().Schema, inCluster)
	r.NotNil(d)

	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		mock.ExpectExec(`DROP CLUSTER "cluster";`).WillReturnResult(sqlmock.NewResult(1, 1))

		if err := clusterDelete(context.TODO(), d, db); err != nil {
			t.Fatal(err)
		}
	})
}
