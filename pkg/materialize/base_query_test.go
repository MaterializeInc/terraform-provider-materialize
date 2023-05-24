package materialize

import (
	"testing"

	"github.com/stretchr/testify/require"
)

var statement = "SELECT * FROM table"

func TestQueryPredicateBasic(t *testing.T) {
	r := require.New(t)
	q := queryPredicate(statement, map[string]string{})
	r.Equal(`SELECT * FROM table;`, q)
}

func TestQueryPredicateParams(t *testing.T) {
	r := require.New(t)
	p := map[string]string{
		"table":    "table_name",
		"schema":   "schema_name",
		"database": "database_name",
		"cluster":  "cluster_name",
		"az":       "us-east-1",
	}
	q := queryPredicate(statement, p)
	r.Equal(`SELECT * FROM table WHERE az = 'us-east-1' AND cluster = 'cluster_name' AND database = 'database_name' AND schema = 'schema_name' AND table = 'table_name';`, q)
}
