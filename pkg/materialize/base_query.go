package materialize

import (
	"fmt"
	"strings"
)

type BaseQuery struct {
	statement string
}

func NewBaseQuery(statement string) *BaseQuery {
	return &BaseQuery{
		statement: statement,
	}
}

func (b *BaseQuery) queryPredicate(predicate map[string]string) string {
	var p []string

	for k, v := range predicate {
		p = append(p, fmt.Sprintf(`%s = %s`, k, QuoteString(v)))
	}

	if len(p) > 0 {
		f := strings.Join(p, " AND ")
		return fmt.Sprintf(`%s WHERE %s;`, b.statement, f)
	}

	return fmt.Sprintf(`%s;`, b.statement)
}
