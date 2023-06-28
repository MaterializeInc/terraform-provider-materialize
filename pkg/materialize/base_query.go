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

func (b *BaseQuery) CustomPredicate(c []string) *BaseQuery {
	b.customPredicate = c
	return b
}

func (b *BaseQuery) Order(c string) *BaseQuery {
	b.order = c
	return b
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
