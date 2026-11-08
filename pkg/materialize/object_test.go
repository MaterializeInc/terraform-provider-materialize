package materialize

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestObjectName(t *testing.T) {
	r := require.New(t)

	on := MaterializeObject{Name: "name"}
	r.Equal(on.QualifiedName(), `"name"`)

	ond := MaterializeObject{Name: "name", DatabaseName: "database"}
	r.Equal(ond.QualifiedName(), `"database"."name"`)

	onsd := MaterializeObject{Name: "name", SchemaName: "schema", DatabaseName: "database"}
	r.Equal(onsd.QualifiedName(), `"database"."schema"."name"`)

	onc := MaterializeObject{Name: "name", ClusterName: "cluster"}
	r.Equal(onc.QualifiedName(), `"cluster"."name"`)
}
