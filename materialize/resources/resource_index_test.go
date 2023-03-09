package resources

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestIndexCreateQuery(t *testing.T) {
	r := require.New(t)
	b := newIndexBuilder("index")
	b.ObjName("database.schema.source")
	b.ClusterName("cluster")
	b.Method("ARRANGEMENT")
	r.Equal(`CREATE INDEX index IN CLUSTER cluster ON database.schema.source USING ARRANGEMENT ();`, b.Create())
}

func TestIndexDropQuery(t *testing.T) {
	r := require.New(t)
	b := newIndexBuilder("index")
	r.Equal(`DROP INDEX index;`, b.Drop())
}
