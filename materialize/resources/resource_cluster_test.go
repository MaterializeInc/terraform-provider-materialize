package resources

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestResourceClusterCreate(t *testing.T) {
	r := require.New(t)
	b := newClusterBuilder("cluster")
	r.Equal(`CREATE CLUSTER cluster REPLICAS ();`, b.Create())
}

func TestResourceClusterRead(t *testing.T) {
	r := require.New(t)
	b := newClusterBuilder("cluster")
	r.Equal(`SELECT id, name FROM mz_clusters WHERE name = 'cluster';`, b.Read())
}

func TestResourceClusterDrop(t *testing.T) {
	r := require.New(t)
	b := newClusterBuilder("cluster")
	r.Equal(`DROP CLUSTER cluster;`, b.Drop())
}
