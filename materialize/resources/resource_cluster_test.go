package resources

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestResourceClusterReadId(t *testing.T) {
	r := require.New(t)
	b := newClusterBuilder("cluster")
	r.Equal(`SELECT id FROM mz_clusters WHERE name = 'cluster';`, b.ReadId())
}

func TestResourceClusterCreate(t *testing.T) {
	r := require.New(t)
	b := newClusterBuilder("cluster")
	r.Equal(`CREATE CLUSTER cluster REPLICAS ();`, b.Create())
}

func TestResourceClusterDrop(t *testing.T) {
	r := require.New(t)
	b := newClusterBuilder("cluster")
	r.Equal(`DROP CLUSTER cluster;`, b.Drop())
}

func TestResourceClusterReadParams(t *testing.T) {
	r := require.New(t)
	b := readClusterParams("u1")
	r.Equal(`SELECT name FROM mz_clusters WHERE id = 'u1';`, b)
}
