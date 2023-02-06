package resources

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestResourceSchemaRead(t *testing.T) {
	r := require.New(t)
	b := newSchemaBuilder("schema", "database")
	r.Equal(`
		SELECT mz_schemas.id, mz_schemas.name, mz_databases.name
		FROM mz_schemas JOIN mz_databases
			ON mz_schemas.database_id = mz_databases.id
		WHERE mz_schemas.name = 'schema'
		AND mz_databases.name = 'database';	
	`, b.Read())
}

func TestResourceSchemaCreate(t *testing.T) {
	r := require.New(t)
	b := newSchemaBuilder("schema", "database")
	r.Equal(`CREATE SCHEMA database.schema;`, b.Create())
}

func TestResourceSchemaDrop(t *testing.T) {
	r := require.New(t)
	b := newSchemaBuilder("schema", "database")
	r.Equal(`DROP SCHEMA database.schema;`, b.Drop())
}
