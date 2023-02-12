package resources

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestResourceSecretReadId(t *testing.T) {
	r := require.New(t)
	b := newSecretBuilder("secret", "schema", "database")
	r.Equal(`
		SELECT mz_secrets.id
		FROM mz_secrets
		JOIN mz_schemas
			ON mz_secrets.schema_id = mz_schemas.id
		JOIN mz_databases
			ON mz_schemas.database_id = mz_databases.id
		WHERE mz_secrets.name = 'secret'
		AND mz_schemas.name = 'schema'
		AND mz_databases.name = 'database';
	`, b.ReadId())
}

func TestResourceSecretCreate(t *testing.T) {
	r := require.New(t)
	b := newSecretBuilder("secret", "schema", "database")
	r.Equal(`CREATE SECRET database.schema.secret AS decode('c2VjcmV0Cg==', 'base64');`, b.Create(`decode('c2VjcmV0Cg==', 'base64')`))
}

func TestResourceSecretRename(t *testing.T) {
	r := require.New(t)
	b := newSecretBuilder("secret", "schema", "database")
	r.Equal(`ALTER SECRET database.schema.secret RENAME TO database.schema.new_secret;`, b.Rename("new_secret"))
}

func TestResourceSecretUpdateValue(t *testing.T) {
	r := require.New(t)
	b := newSecretBuilder("secret", "schema", "database")
	r.Equal(`ALTER SECRET database.schema.secret AS decode('c2VjcmV0Cgdd', 'base64');`, b.UpdateValue(`decode('c2VjcmV0Cgdd', 'base64')`))
}

func TestResourceSecretDrop(t *testing.T) {
	r := require.New(t)
	b := newSecretBuilder("secret", "schema", "database")
	r.Equal(`DROP SECRET database.schema.secret;`, b.Drop())
}

func TestResourceSecretReadParams(t *testing.T) {
	r := require.New(t)
	b := readSecretParams("u1")
	r.Equal(`
		SELECT
			mz_secrets.name,
			mz_schemas.name,
			mz_databases.name
		FROM mz_secrets
		JOIN mz_schemas
			ON mz_secrets.schema_id = mz_schemas.id
		JOIN mz_databases
			ON mz_schemas.database_id = mz_databases.id
		WHERE mz_secrets.id = 'u1';
	`, b)
}
