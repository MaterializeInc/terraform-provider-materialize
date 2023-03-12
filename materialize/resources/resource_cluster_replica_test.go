package resources

import (
	"context"
	"testing"

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
	}
	d := schema.TestResourceDataRaw(t, ClusterReplica().Schema, in)
	r.NotNil(d)

	WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		// Create
		mock.ExpectExec(
			`CREATE CLUSTER REPLICA "cluster"."replica" SIZE = 'small', AVAILABILITY ZONE = 'use1-az1', INTROSPECTION INTERVAL = '10s', INTROSPECTION DEBUGGING = TRUE, IDLE ARRANGEMENT MERGE EFFORT = 100;`,
		).WillReturnResult(sqlmock.NewResult(1, 1))

		// Query Id
		ir := mock.NewRows([]string{"id"}).
			AddRow("u1")
		mock.ExpectQuery(`
			SELECT mz_cluster_replicas.id
			FROM mz_cluster_replicas
			JOIN mz_clusters
				ON mz_cluster_replicas.cluster_id = mz_clusters.id
			WHERE mz_cluster_replicas.name = 'replica'
			AND mz_clusters.name = 'cluster'`).WillReturnRows(ir)

		// Query Params
		ip := sqlmock.NewRows([]string{"name", "cluster", "size", "availability_zone"}).
			AddRow("replica", "cluster", "small", "us-east-1")
		mock.ExpectQuery(`
			SELECT
				mz_cluster_replicas.name,
				mz_clusters.name,
				mz_cluster_replicas.size,
				mz_cluster_replicas.availability_zone
			FROM mz_cluster_replicas
			JOIN mz_clusters
				ON mz_cluster_replicas.cluster_id = mz_clusters.id
			WHERE mz_cluster_replicas.id = 'u1';`).WillReturnRows(ip)

		if err := clusterReplicaCreate(context.TODO(), d, db); err != nil {
			t.Fatal(err)
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

	WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		mock.ExpectExec(`DROP CLUSTER REPLICA "cluster"."replica";`).WillReturnResult(sqlmock.NewResult(1, 1))

		if err := clusterReplicaDelete(context.TODO(), d, db); err != nil {
			t.Fatal(err)
		}
	})

}

func TestClusterReplicaCreateQuery(t *testing.T) {
	r := require.New(t)
	b := newClusterReplicaBuilder("replica", "cluster")
	r.Equal(`CREATE CLUSTER REPLICA "cluster"."replica";`, b.Create())

	b.Size("xsmall")
	r.Equal(`CREATE CLUSTER REPLICA "cluster"."replica" SIZE = 'xsmall';`, b.Create())

	b.AvailabilityZone("us-east-1")
	r.Equal(`CREATE CLUSTER REPLICA "cluster"."replica" SIZE = 'xsmall', AVAILABILITY ZONE = 'us-east-1';`, b.Create())

	b.IntrospectionInterval("1s")
	r.Equal(`CREATE CLUSTER REPLICA "cluster"."replica" SIZE = 'xsmall', AVAILABILITY ZONE = 'us-east-1', INTROSPECTION INTERVAL = '1s';`, b.Create())

	b.IntrospectionDebugging()
	r.Equal(`CREATE CLUSTER REPLICA "cluster"."replica" SIZE = 'xsmall', AVAILABILITY ZONE = 'us-east-1', INTROSPECTION INTERVAL = '1s', INTROSPECTION DEBUGGING = TRUE;`, b.Create())

	b.IdleArrangementMergeEffort(1)
	r.Equal(`CREATE CLUSTER REPLICA "cluster"."replica" SIZE = 'xsmall', AVAILABILITY ZONE = 'us-east-1', INTROSPECTION INTERVAL = '1s', INTROSPECTION DEBUGGING = TRUE, IDLE ARRANGEMENT MERGE EFFORT = 1;`, b.Create())
}

func TestClusterReplicaDropQuery(t *testing.T) {
	r := require.New(t)
	b := newClusterReplicaBuilder("replica", "cluster")
	r.Equal(`DROP CLUSTER REPLICA "cluster"."replica";`, b.Drop())
}

func TestClusterReplicaReadQuery(t *testing.T) {
	r := require.New(t)
	b := newClusterReplicaBuilder("replica", "cluster")
	r.Equal(`
		SELECT mz_cluster_replicas.id
		FROM mz_cluster_replicas
		JOIN mz_clusters
			ON mz_cluster_replicas.cluster_id = mz_clusters.id
		WHERE mz_cluster_replicas.name = 'replica'
		AND mz_clusters.name = 'cluster';`, b.ReadId())
}

func TestClusterReplicaReadParamsQuery(t *testing.T) {
	r := require.New(t)
	b := readClusterReplicaParams("u1")
	r.Equal(`
		SELECT
			mz_cluster_replicas.name,
			mz_clusters.name,
			mz_cluster_replicas.size,
			mz_cluster_replicas.availability_zone
		FROM mz_cluster_replicas
		JOIN mz_clusters
			ON mz_cluster_replicas.cluster_id = mz_clusters.id
		WHERE mz_cluster_replicas.id = 'u1';`, b)
}
