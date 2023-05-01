package materialize

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRoleCreateQuery(t *testing.T) {
	r := require.New(t)
	b := NewRoleBuilder("role")
	b.Inherit()
	b.CreateRole()
	b.CreateDb()
	b.CreateCluster()
	r.Equal(`CREATE ROLE "role" INHERIT CREATEROLE CREATEDB CREATECLUSTER;`, b.Create())
}

func TestRoleAlterQuery(t *testing.T) {
	r := require.New(t)
	b := NewRoleBuilder("role")
	r.Equal(`ALTER ROLE "role" CREATEROLE;`, b.Alter("CREATEROLE"))
}

func TestRoleDropQuery(t *testing.T) {
	r := require.New(t)
	b := NewRoleBuilder("role")
	r.Equal(`DROP ROLE "role";`, b.Drop())
}
