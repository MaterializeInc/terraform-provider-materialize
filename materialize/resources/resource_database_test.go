package resources

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestResourceDatabaseCreate(t *testing.T) {
	r := require.New(t)
	b := newDatabaseBuilder("database")
	r.Equal(`CREATE DATABASE database;`, b.Create())
}

func TestResourceDatabaseRead(t *testing.T) {
	r := require.New(t)
	b := newDatabaseBuilder("database")
	r.Equal(`SELECT id, name FROM mz_databases WHERE name = 'database';`, b.Read())
}

func TestResourceDatabaseDrop(t *testing.T) {
	r := require.New(t)
	b := newDatabaseBuilder("database")
	r.Equal(`DROP DATABASE database;`, b.Drop())
}
