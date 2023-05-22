package materialize

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDatabaseReadIdQuery(t *testing.T) {
	r := require.New(t)
	b := NewDatabaseBuilder("database")
	r.Equal(`SELECT id FROM mz_databases WHERE name = 'database';`, b.ReadId())
}

func TestDatabaseCreateQuery(t *testing.T) {
	r := require.New(t)
	b := NewDatabaseBuilder("database")
	r.Equal(`CREATE DATABASE "database";`, b.Create())
}

func TestDatabaseDropQuery(t *testing.T) {
	r := require.New(t)
	b := NewDatabaseBuilder("database")
	r.Equal(`DROP DATABASE "database";`, b.Drop())
}

func TestDatabaseReadParamsQuery(t *testing.T) {
	r := require.New(t)
	b := ReadDatabaseParams("u1")
	r.Equal(`SELECT name AS database_name FROM mz_databases WHERE id = 'u1';`, b)
}
