package materialize

import (
	"fmt"
	"sort"
	"strings"
)

type BaseQuery struct {
	statement       string
	customPredicate []string
	order           string
}

func NewBaseQuery(statement string) *BaseQuery {
	return &BaseQuery{statement: statement}
}

// Statement returns the query without its predicates and order, so the test
// mocks can be built from the same SQL.
func (b *BaseQuery) Statement() string {
	return b.statement
}

// CustomPredicate and Order return a copy, so the shared query variables are
// not changed by callers adding their own predicates.
func (b *BaseQuery) CustomPredicate(c []string) *BaseQuery {
	q := *b
	q.customPredicate = c
	return &q
}

func (b *BaseQuery) Order(c string) *BaseQuery {
	q := *b
	q.order = c
	return &q
}

func (b *BaseQuery) QueryPredicate(predicate map[string]string) string {
	q := strings.Builder{}
	q.WriteString(b.statement)

	var p []string

	// predicate mapping
	for k, v := range predicate {
		if v != "" {
			p = append(p, fmt.Sprintf(`%s = %s`, k, QuoteString(v)))
		}
	}

	// custom predicates
	p = append(p, b.customPredicate...)

	if len(p) > 0 {
		// sort predicateds for testing consistency
		sort.Strings(p)
		f := strings.Join(p, " AND ")
		q.WriteString(fmt.Sprintf(` WHERE %s`, f))
	}

	if b.order != "" {
		q.WriteString(fmt.Sprintf(` ORDER BY %s`, b.order))
	}

	q.WriteString(";")
	return q.String()
}
