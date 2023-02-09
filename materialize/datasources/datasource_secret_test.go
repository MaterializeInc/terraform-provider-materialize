package datasources

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSecretQuery(t *testing.T) {
	r := require.New(t)
	b := secretQuery("", "")
	r.Equal(`
		SELECT
			mz_secrets.id,
			mz_secrets.name,
			mz_schemas.name,
			mz_databases.name
		FROM mz_secrets
		JOIN mz_schemas
			ON mz_secrets.schema_id = mz_schemas.id
		JOIN mz_databases
			ON mz_schemas.database_id = mz_databases.id;`, b)
}

func TestSecretDatabaseQuery(t *testing.T) {
	r := require.New(t)
	b := secretQuery("database", "")
	r.Equal(`
		SELECT
			mz_secrets.id,
			mz_secrets.name,
			mz_schemas.name,
			mz_databases.name
		FROM mz_secrets
		JOIN mz_schemas
			ON mz_secrets.schema_id = mz_schemas.id
		JOIN mz_databases
			ON mz_schemas.database_id = mz_databases.id
		WHERE mz_databases.name = 'database';`, b)
}

func TestSecretSchemaDatabaseQuery(t *testing.T) {
	r := require.New(t)
	b := secretQuery("database", "schema")
	r.Equal(`
		SELECT
			mz_secrets.id,
			mz_secrets.name,
			mz_schemas.name,
			mz_databases.name
		FROM mz_secrets
		JOIN mz_schemas
			ON mz_secrets.schema_id = mz_schemas.id
		JOIN mz_databases
			ON mz_schemas.database_id = mz_databases.id
		WHERE mz_databases.name = 'database' AND mz_schemas.name = 'schema';`, b)
}
