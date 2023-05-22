package materialize

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestClusterCreateQuery(t *testing.T) {
	r := require.New(t)
	b := NewClusterBuilder("cluster")
	r.Equal(`CREATE CLUSTER "cluster" REPLICAS ();`, b.Create())
}

func TestClusterDropQuery(t *testing.T) {
	r := require.New(t)
	b := NewClusterBuilder("cluster")
	r.Equal(`DROP CLUSTER "cluster";`, b.Drop())
}

func TestClusterReadIdQuery(t *testing.T) {
	r := require.New(t)
	b := NewClusterBuilder("cluster")
	r.Equal(`SELECT id FROM mz_clusters WHERE name = 'cluster';`, b.ReadId())
}

func TestClusterReadParamsQuery(t *testing.T) {
	r := require.New(t)
	b := ReadClusterParams("u1")
	r.Equal(`SELECT name AS cluster_name FROM mz_clusters WHERE id = 'u1';`, b)
}
