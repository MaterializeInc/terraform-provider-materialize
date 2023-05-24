package materialize

import (
	"testing"

	"github.com/stretchr/testify/require"
)

var statement = "SELECT * FROM table"

func TestQueryPredicateBasic(t *testing.T) {
	r := require.New(t)
	b := NewBaseQuery(statement)

	q := b.QueryPredicate(map[string]string{})
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
	b := NewBaseQuery(statement)

	q := b.QueryPredicate(p)
	r.Equal(`SELECT * FROM table WHERE az = 'us-east-1' AND cluster = 'cluster_name' AND database = 'database_name' AND schema = 'schema_name' AND table = 'table_name';`, q)
}

func TestQueryPredicateAdditionalParams(t *testing.T) {
	r := require.New(t)
	b := NewBaseQuery(statement)
	b.CustomPredicate([]string{"salary BETWEEN 1 AND 10"})

	q := b.QueryPredicate(map[string]string{})
	r.Equal(`SELECT * FROM table WHERE salary BETWEEN 1 AND 10;`, q)
}

func TestQueryPredicateAllParams(t *testing.T) {
	r := require.New(t)
	b := NewBaseQuery(statement)
	b.CustomPredicate([]string{"salary BETWEEN 1 AND 10"})

	q := b.QueryPredicate(map[string]string{"table": "table_name"})
	r.Equal(`SELECT * FROM table WHERE salary BETWEEN 1 AND 10 AND table = 'table_name';`, q)
}
