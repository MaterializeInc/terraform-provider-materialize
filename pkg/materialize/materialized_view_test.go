package materialize

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMaterializedViewCreateQuery(t *testing.T) {
	r := require.New(t)
	b := NewMaterializedViewBuilder("materialized_view", "schema", "database")
	b.SelectStmt("SELECT 1 FROM t1")
	r.Equal(`CREATE MATERIALIZED VIEW "database"."schema"."materialized_view" AS SELECT 1 FROM t1;`, b.Create())
}

func TestMaterializedViewCreateQueryIfNotExist(t *testing.T) {
	r := require.New(t)
	b := NewMaterializedViewBuilder("materialized_view", "schema", "database")
	b.SelectStmt("SELECT 1 FROM t1")
	r.Equal(`CREATE MATERIALIZED VIEW "database"."schema"."materialized_view" AS SELECT 1 FROM t1;`, b.Create())
}

func TestMaterializedViewRenameQuery(t *testing.T) {
	r := require.New(t)
	b := NewMaterializedViewBuilder("materialized_view", "schema", "database")
	r.Equal(`ALTER MATERIALIZED VIEW "database"."schema"."materialized_view" RENAME TO "database"."schema"."new_view";`, b.Rename("new_view"))
}

func TestMaterializedViewDropQuery(t *testing.T) {
	r := require.New(t)
	b := NewMaterializedViewBuilder("materialized_view", "schema", "database")
	r.Equal(`DROP MATERIALIZED VIEW "database"."schema"."materialized_view";`, b.Drop())
}
