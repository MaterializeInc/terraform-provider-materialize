package resources

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestResourceClusterReplicaCreate(t *testing.T) {
	r := require.New(t)
	b := newClusterReplicaBuilder("cluster", "replica")
	r.Equal(`CREATE CLUSTER REPLICA cluster.replica;`, b.Create())

	b.Size("xsmall")
	r.Equal(`CREATE CLUSTER REPLICA cluster.replica SIZE = 'xsmall';`, b.Create())

	b.AvailabilityZone("us-east-1")
	r.Equal(`CREATE CLUSTER REPLICA cluster.replica SIZE = 'xsmall' AVAILABILITY ZONE = 'us-east-1';`, b.Create())

	b.IntrospectionInterval("1s")
	r.Equal(`CREATE CLUSTER REPLICA cluster.replica SIZE = 'xsmall' AVAILABILITY ZONE = 'us-east-1' INTROSPECTION INTERVAL = '1s';`, b.Create())

	b.IntrospectionDebugging()
	r.Equal(`CREATE CLUSTER REPLICA cluster.replica SIZE = 'xsmall' AVAILABILITY ZONE = 'us-east-1' INTROSPECTION INTERVAL = '1s' INTROSPECTION DEBUGGING = TRUE;`, b.Create())

	b.IdleArrangementMergeEffort(1)
	r.Equal(`CREATE CLUSTER REPLICA cluster.replica SIZE = 'xsmall' AVAILABILITY ZONE = 'us-east-1' INTROSPECTION INTERVAL = '1s' INTROSPECTION DEBUGGING = TRUE IDLE ARRANGEMENT MERGE EFFORT = 1;`, b.Create())
}

func TestResourceClusterReplicaRead(t *testing.T) {
	r := require.New(t)
	b := newClusterReplicaBuilder("cluster", "replica")
	r.Equal(`
		SELECT
			mz_cluster_replicas.id,
			mz_cluster_replicas.name,
			mz_clusters.name,
			mz_cluster_replicas.size,
			mz_cluster_replicas.availability_zone
		FROM mz_cluster_replicas
		JOIN mz_clusters
			ON mz_cluster_replicas.cluster_id = mz_clusters.id
		WHERE mz_cluster_replicas.name = 'replica'
		AND mz_clusters.name = 'cluster';
	`, b.Read())
}

func TestResourceClusterReplicaDrop(t *testing.T) {
	r := require.New(t)
	b := newClusterReplicaBuilder("cluster", "replica")
	r.Equal(`DROP CLUSTER REPLICA cluster.replica;`, b.Drop())
}
