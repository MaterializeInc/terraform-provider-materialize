package materialize

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSchemaReadIdQuery(t *testing.T) {
	r := require.New(t)
	b := NewSchemaBuilder("schema", "database")
	r.Equal(`
		SELECT mz_schemas.id
		FROM mz_schemas JOIN mz_databases
			ON mz_schemas.database_id = mz_databases.id
		WHERE mz_schemas.name = 'schema'
		AND mz_databases.name = 'database';
	`, b.ReadId())
}

func TestSchemaCreateQuery(t *testing.T) {
	r := require.New(t)
	b := NewSchemaBuilder("schema", "database")
	r.Equal(`CREATE SCHEMA "database"."schema";`, b.Create())
}

func TestSchemaDropQuery(t *testing.T) {
	r := require.New(t)
	b := NewSchemaBuilder("schema", "database")
	r.Equal(`DROP SCHEMA "database"."schema";`, b.Drop())
}

func TestSchemaReadParamsQuery(t *testing.T) {
	r := require.New(t)
	b := ReadSchemaParams("u1")
	r.Equal(`
		SELECT
			mz_schemas.name,
			mz_databases.name
		FROM mz_schemas JOIN mz_databases
			ON mz_schemas.database_id = mz_databases.id
		WHERE mz_schemas.id = 'u1';`, b)
}
