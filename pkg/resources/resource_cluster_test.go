package resources

import (
	"context"
	"testing"

	"github.com/MaterializeInc/terraform-provider-materialize/pkg/testhelpers"
	"github.com/MaterializeInc/terraform-provider-materialize/pkg/utils"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/require"
)

var inCluster = map[string]interface{}{
	"name":                    "cluster",
	"size":                    "3xsmall",
	"replication_factor":      2,
	"availability_zones":      []interface{}{"use1-az1", "use1-az2"},
	"introspection_interval":  "10s",
	"introspection_debugging": true,
	"ownership_role":          "joe",
}

func TestResourceClusterCreate(t *testing.T) {
	r := require.New(t)

	d := schema.TestResourceDataRaw(t, Cluster().Schema, inCluster)
	r.NotNil(d)

	testhelpers.WithMockProviderMeta(t, func(db *utils.ProviderMeta, mock sqlmock.Sqlmock) {
		// Create
		mock.ExpectExec(`
			CREATE CLUSTER "cluster"
			SIZE '3xsmall',
			REPLICATION FACTOR 2,
			AVAILABILITY ZONES = \['use1-az1','use1-az2'\],
			INTROSPECTION INTERVAL = '10s',
			INTROSPECTION DEBUGGING = TRUE;
		`).WillReturnResult(sqlmock.NewResult(1, 1))

		// Ownership
		mock.ExpectExec(`ALTER CLUSTER "cluster" OWNER TO "joe";`).WillReturnResult(sqlmock.NewResult(1, 1))

		// Query Id
		ip := `WHERE mz_clusters.name = 'cluster'`
		testhelpers.MockClusterScan(mock, ip)

		// Query Params
		pp := `WHERE mz_clusters.id = 'u1'`
		testhelpers.MockClusterScan(mock, pp)

		if err := clusterCreate(context.TODO(), d, db); err != nil {
			t.Fatal(err)
		}
	})
}

// Confirm id is updated with region for 0.4.0
func TestResourceClusterReadIdMigration(t *testing.T) {
	utils.SetDefaultRegion("aws/us-east-1")
	r := require.New(t)

	in := map[string]interface{}{
		"name": "cluster",
	}
	d := schema.TestResourceDataRaw(t, Cluster().Schema, in)
	r.NotNil(d)

	// Set id before migration
	d.SetId("u1")

	testhelpers.WithMockProviderMeta(t, func(db *utils.ProviderMeta, mock sqlmock.Sqlmock) {
		// Query Params
		pp := `WHERE mz_clusters.id = 'u1'`
		testhelpers.MockClusterScan(mock, pp)

		if err := clusterRead(context.TODO(), d, db); err != nil {
			t.Fatal(err)
		}

		if d.Id() != "aws/us-east-1:u1" {
			t.Fatalf("unexpected id of %s", d.Id())
		}
	})
}

func TestResourceClusterZeroReplicationCreate(t *testing.T) {
	r := require.New(t)

	var inClusterZeroReplication = map[string]interface{}{
		"name":               "cluster",
		"size":               "3xsmall",
		"replication_factor": 0,
	}
	d := schema.TestResourceDataRaw(t, Cluster().Schema, inClusterZeroReplication)
	r.NotNil(d)

	testhelpers.WithMockProviderMeta(t, func(db *utils.ProviderMeta, mock sqlmock.Sqlmock) {
		// Create
		mock.ExpectExec(`
			CREATE CLUSTER "cluster"
			SIZE '3xsmall',
			REPLICATION FACTOR 0,
			INTROSPECTION INTERVAL = '1s';
		`).WillReturnResult(sqlmock.NewResult(1, 1))

		// Query Id
		ip := `WHERE mz_clusters.name = 'cluster'`
		testhelpers.MockClusterScan(mock, ip)

		// Query Params
		pp := `WHERE mz_clusters.id = 'u1'`
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

	testhelpers.WithMockProviderMeta(t, func(db *utils.ProviderMeta, mock sqlmock.Sqlmock) {
		mock.ExpectExec(`DROP CLUSTER "cluster";`).WillReturnResult(sqlmock.NewResult(1, 1))

		if err := clusterDelete(context.TODO(), d, db); err != nil {
			t.Fatal(err)
		}
	})
}
