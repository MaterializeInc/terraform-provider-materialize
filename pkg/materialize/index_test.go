package materialize

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestIndexCreateQuery(t *testing.T) {
	r := require.New(t)
	b := NewIndexBuilder("index")
	b.ObjName(IdentifierSchemaStruct{SchemaName: "schema", Name: "source", DatabaseName: "database"})
	b.ClusterName("cluster")
	b.Method("ARRANGEMENT")
	r.Equal(`CREATE INDEX index IN CLUSTER cluster ON database.schema.source USING ARRANGEMENT ();`, b.Create())
}

func TestIndexDropQuery(t *testing.T) {
	r := require.New(t)
	b := NewIndexBuilder("index")
	r.Equal(`DROP INDEX "database"."schema"."index" RESTRICT;`, b.Drop("database", "schema"))
}
