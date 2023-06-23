package materialize

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestObjectName(t *testing.T) {
	r := require.New(t)

	on := ObjectSchemaStruct{Name: "name"}
	r.Equal(on.QualifiedName(), `"name"`)

	ond := ObjectSchemaStruct{Name: "name", DatabaseName: "database"}
	r.Equal(ond.QualifiedName(), `"database"."name"`)

	onsd := ObjectSchemaStruct{Name: "name", SchemaName: "schema", DatabaseName: "database"}
	r.Equal(onsd.QualifiedName(), `"database"."schema"."name"`)

	onc := ObjectSchemaStruct{Name: "name", ClusterName: "cluster"}
	r.Equal(onc.QualifiedName(), `"cluster"."name"`)
}
