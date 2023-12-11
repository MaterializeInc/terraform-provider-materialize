package resources

import (
	"context"
	"testing"

	"github.com/MaterializeInc/terraform-provider-materialize/pkg/testhelpers"
	"github.com/MaterializeInc/terraform-provider-materialize/pkg/utils"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/require"
)

func TestResourceClusterReplicaCreate(t *testing.T) {
	r := require.New(t)

	in := map[string]interface{}{
		"name":                          "replica",
		"cluster_name":                  "cluster",
		"size":                          "small",
		"availability_zone":             "use1-az1",
		"introspection_interval":        "10s",
		"introspection_debugging":       true,
		"idle_arrangement_merge_effort": 100,
		"comment":                       "object comment",
	}
	d := schema.TestResourceDataRaw(t, ClusterReplica().Schema, in)
	r.NotNil(d)

	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		// Create
		mock.ExpectExec(`
			CREATE CLUSTER REPLICA "cluster"."replica"
			SIZE = 'small',
			AVAILABILITY ZONE = 'use1-az1',
			INTROSPECTION INTERVAL = '10s',
			INTROSPECTION DEBUGGING = TRUE,
			IDLE ARRANGEMENT MERGE EFFORT = 100;`,
		).WillReturnResult(sqlmock.NewResult(1, 1))

		// Comment
		mock.ExpectExec(`COMMENT ON CLUSTER REPLICA "cluster"."replica" IS 'object comment';`).WillReturnResult(sqlmock.NewResult(1, 1))

		// Query Id
		ip := `WHERE mz_cluster_replicas.name = 'replica' AND mz_clusters.name = 'cluster'`
		testhelpers.MockClusterReplicaScan(mock, ip)

		// Query Params
		pp := `WHERE mz_cluster_replicas.id = 'u1'`
		testhelpers.MockClusterReplicaScan(mock, pp)

		if err := clusterReplicaCreate(context.TODO(), d, db); err != nil {
			t.Fatal(err)
		}
	})
}

// Confirm id is updated with region for 0.4.0
func TestResourceClusterReplicaReadIdMigration(t *testing.T) {
	utils.SetRegionFromHostname("localhost")
	r := require.New(t)

	in := map[string]interface{}{
		"name": "replica",
	}
	d := schema.TestResourceDataRaw(t, ClusterReplica().Schema, in)
	r.NotNil(d)

	// Set id before migration
	d.SetId("u1")

	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		// Query Params
		pp := `WHERE mz_cluster_replicas.id = 'u1'`
		testhelpers.MockClusterReplicaScan(mock, pp)

		if err := clusterReplicaRead(context.TODO(), d, db); err != nil {
			t.Fatal(err)
		}

		if d.Id() != "aws/us-east-1:u1" {
			t.Fatalf("unexpected id of %s", d.Id())
		}
	})
}

func TestResourceClusterReplicaDelete(t *testing.T) {
	r := require.New(t)

	in := map[string]interface{}{
		"name":         "replica",
		"cluster_name": "cluster",
	}
	d := schema.TestResourceDataRaw(t, ClusterReplica().Schema, in)
	r.NotNil(d)

	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		mock.ExpectExec(`DROP CLUSTER REPLICA "cluster"."replica";`).WillReturnResult(sqlmock.NewResult(1, 1))

		if err := clusterReplicaDelete(context.TODO(), d, db); err != nil {
			t.Fatal(err)
		}
	})
}
