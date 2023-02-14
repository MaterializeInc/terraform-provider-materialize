package resources

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestResourceDatabaseReadId(t *testing.T) {
	r := require.New(t)
	b := newDatabaseBuilder("database")
	r.Equal(`SELECT id FROM mz_databases WHERE name = 'database';`, b.ReadId())
}

func TestResourceDatabaseCreate(t *testing.T) {
	r := require.New(t)
	b := newDatabaseBuilder("database")
	r.Equal(`CREATE DATABASE database;`, b.Create())
}

func TestResourceDatabaseDrop(t *testing.T) {
	r := require.New(t)
	b := newDatabaseBuilder("database")
	r.Equal(`DROP DATABASE database;`, b.Drop())
}

func TestResourceDatabaseReadParams(t *testing.T) {
	r := require.New(t)
	b := readDatabaseParams("u1")
	r.Equal(`SELECT name FROM mz_databases WHERE id = 'u1';`, b)
}
