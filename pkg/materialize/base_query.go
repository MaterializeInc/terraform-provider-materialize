package materialize

import (
	"fmt"
	"strings"
)

func queryPredicate(statement string, predicate map[string]string) string {
	var p []string

	for k, v := range predicate {
		if v != "" {
			p = append(p, fmt.Sprintf(`%s = %s`, k, QuoteString(v)))
		}
	}

	if len(p) > 0 {
		f := strings.Join(p, " AND ")
		return fmt.Sprintf(`%s WHERE %s;`, statement, f)
	}

	return fmt.Sprintf(`%s;`, statement)
}
