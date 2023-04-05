package materialize

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestIndexCreateQuery(t *testing.T) {
	r := require.New(t)
	b := NewIndexBuilder("index", false, IdentifierSchemaStruct{SchemaName: "schema", Name: "source", DatabaseName: "database"})
	b.ClusterName("cluster")
	b.Method("ARRANGEMENT")
	b.ColExpr([]IndexColumn{
		{
			Field: "Column",
			Val:   "LONG",
		},
	})
	r.Equal(`CREATE INDEX index IN CLUSTER cluster ON "database"."schema"."source" USING ARRANGEMENT (Column LONG);`, b.Create())
}

func TestIndexDefaultCreateQuery(t *testing.T) {
	r := require.New(t)
	b := NewIndexBuilder("", true, IdentifierSchemaStruct{SchemaName: "schema", Name: "source", DatabaseName: "database"})
	b.ClusterName("cluster")
	b.Method("ARRANGEMENT")
	r.Equal(`CREATE DEFAULT INDEX IN CLUSTER cluster ON "database"."schema"."source" USING ARRANGEMENT ();`, b.Create())
}

func TestIndexDropQuery(t *testing.T) {
	r := require.New(t)
	b := NewIndexBuilder("index", false, IdentifierSchemaStruct{SchemaName: "schema", Name: "source", DatabaseName: "database"})
	r.Equal(`DROP INDEX "database"."schema"."index" RESTRICT;`, b.Drop())
}
