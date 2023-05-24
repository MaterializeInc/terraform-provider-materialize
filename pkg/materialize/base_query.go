package materialize

import (
	"fmt"
	"strings"
)

func queryPredicate(statement string, predicate map[string]string) string {
	q := strings.Builder{}
	q.WriteString(statement)

	var p []string
	for k, v := range predicate {
		if v != "" {
			p = append(p, fmt.Sprintf(`%s = %s`, k, QuoteString(v)))
		}
	}

	if len(p) > 0 {
		f := strings.Join(p, " AND ")
		q.WriteString(fmt.Sprintf(` WHERE %s`, f))
	}

	q.WriteString(";")
	return q.String()
}
