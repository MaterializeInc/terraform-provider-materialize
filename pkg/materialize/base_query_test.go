package materialize

import (
	"testing"

	"github.com/stretchr/testify/require"
)

var exampleStatement = `SELECT id, name, schema_name, database, parameter FROM mz_example`

func TestQueryPredicateParams(t *testing.T) {
	r := require.New(t)
	p := map[string]string{
		"mz_example.name":   "example",
		"mz_schemas.name":   "schema",
		"mz_databases.name": "database",
	}

	b := NewBaseQuery(exampleStatement)
	r.Equal(`SELECT id, name, schema_name, database, parameter FROM mz_example WHERE mz_example.name = 'example' AND mz_schemas.name = 'schema' AND mz_databases.name = 'database';`, b.queryPredicate(p))
}

func TestQueryPredicateId(t *testing.T) {
	r := require.New(t)
	p := map[string]string{
		"mz_example.id": "u1",
	}

	b := NewBaseQuery(exampleStatement)
	r.Equal(`SELECT id, name, schema_name, database, parameter FROM mz_example WHERE mz_example.id = 'u1';`, b.queryPredicate(p))
}
