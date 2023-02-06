package resources

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestResourceSecretRead(t *testing.T) {
	r := require.New(t)
	b := newSecretBuilder("secret", "schema")
	r.Equal(`
		SELECT mz_secrets.id, mz_secrets.name, mz_schemas.name
		FROM mz_secrets JOIN mz_schemas
			ON mz_secrets.schema_id = mz_schemas.id
		WHERE mz_secrets.name = 'secret'
		AND mz_schemas.name = 'schema';
	`, b.Read())
}

func TestResourceSecretCreate(t *testing.T) {
	r := require.New(t)
	b := newSecretBuilder("secret", "schema")
	r.Equal(`CREATE SECRET schema.secret AS decode('c2VjcmV0Cg==', 'base64');`, b.Create(`decode('c2VjcmV0Cg==', 'base64')`))
}

func TestResourceSecretRename(t *testing.T) {
	r := require.New(t)
	b := newSecretBuilder("secret", "schema")
	r.Equal(`ALTER SECRET schema.secret RENAME TO schema.new_secret;`, b.Rename("new_secret"))
}

func TestResourceSecretUpdateValue(t *testing.T) {
	r := require.New(t)
	b := newSecretBuilder("secret", "schema")
	r.Equal(`ALTER SECRET schema.secret AS decode('c2VjcmV0Cgdd', 'base64');`, b.UpdateValue(`decode('c2VjcmV0Cgdd', 'base64')`))
}

func TestResourceSecretDrop(t *testing.T) {
	r := require.New(t)
	b := newSecretBuilder("secret", "schema")
	r.Equal(`DROP SECRET schema.secret;`, b.Drop())
}
