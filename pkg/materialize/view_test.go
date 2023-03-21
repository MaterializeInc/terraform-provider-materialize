package materialize

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestViewCreateQuery(t *testing.T) {
	r := require.New(t)
	b := NewViewBuilder("view", "schema", "database")
	b.SelectStmt("SELECT 1 FROM t1")
	r.Equal(`CREATE VIEW "database"."schema"."view" AS SELECT 1 FROM t1;`, b.Create())
}

func TestViewRenameQuery(t *testing.T) {
	r := require.New(t)
	b := NewViewBuilder("view", "schema", "database")
	r.Equal(`ALTER VIEW "database"."schema"."view" RENAME TO "database"."schema"."new_view";`, b.Rename("new_view"))
}

func TestViewDropQuery(t *testing.T) {
	r := require.New(t)
	b := NewViewBuilder("view", "schema", "database")
	r.Equal(`DROP VIEW "database"."schema"."view";`, b.Drop())
}
