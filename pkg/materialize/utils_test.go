package materialize

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestQualifiedName(t *testing.T) {
	r := require.New(t)
	q := QualifiedName("database", "schema", "resource")
	r.Equal(q, `"database"."schema"."resource"`)

	rs := require.New(t)
	qs := QualifiedName("database", "schema")
	rs.Equal(qs, `"database"."schema"`)
}
