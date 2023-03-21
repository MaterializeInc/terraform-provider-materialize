package materialize

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSecretCreateQuery(t *testing.T) {
	r := require.New(t)
	b := NewSecretBuilder("secret", "schema", "database")
	r.Equal(`CREATE SECRET "database"."schema"."secret" AS 'c2VjcmV0Cg';`, b.Create(`c2VjcmV0Cg`))
}

func TestSecretCreateEmptyValueQuery(t *testing.T) {
	r := require.New(t)
	b := NewSecretBuilder("secret", "schema", "database")
	r.Equal(`CREATE SECRET "database"."schema"."secret" AS '';`, b.Create(``))
}

func TestSecretCreateEscapedValueQuery(t *testing.T) {
	r := require.New(t)
	b := NewSecretBuilder("secret", "schema", "database")
	r.Equal(`CREATE SECRET "database"."schema"."secret" AS 'c2Vjcm''V0Cg';`, b.Create(`c2Vjcm'V0Cg`))
}

func TestSecretRenameQuery(t *testing.T) {
	r := require.New(t)
	b := NewSecretBuilder("secret", "schema", "database")
	r.Equal(`ALTER SECRET "database"."schema"."secret" RENAME TO "database"."schema"."new_secret";`, b.Rename("new_secret"))
}

func TestSecretUpdateValueQuery(t *testing.T) {
	r := require.New(t)
	b := NewSecretBuilder("secret", "schema", "database")
	r.Equal(`ALTER SECRET "database"."schema"."secret" AS 'c2VjcmV0Cgdd';`, b.UpdateValue(`c2VjcmV0Cgdd`))
}

func TestSecretUpdateEscapedValueQuery(t *testing.T) {
	r := require.New(t)
	b := NewSecretBuilder("secret", "schema", "database")
	r.Equal(`ALTER SECRET "database"."schema"."secret" AS 'c2Vjcm''V0Cgdd';`, b.UpdateValue(`c2Vjcm'V0Cgdd`))
}

func TestSecretDropQuery(t *testing.T) {
	r := require.New(t)
	b := NewSecretBuilder("secret", "schema", "database")
	r.Equal(`DROP SECRET "database"."schema"."secret";`, b.Drop())
}

func TestSecretReadIdQuery(t *testing.T) {
	r := require.New(t)
	b := NewSecretBuilder("secret", "schema", "database")
	r.Equal(`
		SELECT mz_secrets.id
		FROM mz_secrets
		JOIN mz_schemas
			ON mz_secrets.schema_id = mz_schemas.id
		JOIN mz_databases
			ON mz_schemas.database_id = mz_databases.id
		WHERE mz_secrets.name = 'secret'
		AND mz_schemas.name = 'schema'
		AND mz_databases.name = 'database';`, b.ReadId())
}

func TestSecretReadParamsQuery(t *testing.T) {
	r := require.New(t)
	b := ReadSecretParams("u1")
	r.Equal(`
		SELECT
			mz_secrets.name AS name,
			mz_schemas.name AS schema_name,
			mz_databases.name AS database_name
		FROM mz_secrets
		JOIN mz_schemas
			ON mz_secrets.schema_id = mz_schemas.id
		JOIN mz_databases
			ON mz_schemas.database_id = mz_databases.id
		WHERE mz_secrets.id = 'u1';`, b)
}
