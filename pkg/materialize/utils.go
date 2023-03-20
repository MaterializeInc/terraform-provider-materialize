package materialize

import (
	"fmt"
	"strings"
)

type SQLError struct {
	Err error
}

func (e *SQLError) Error() string {
	return fmt.Sprintf("Unable to execute SQL: %v", e.Err)
}

func QuoteString(input string) (output string) {
	output = "'" + strings.Replace(input, "'", "''", -1) + "'"
	return
}

func QuoteIdentifier(input string) (output string) {
	parts := strings.Split(input, ".")
	for i, p := range parts {
		parts[i] = `"` + strings.Replace(p, `"`, `""`, -1) + `"`
	}
	output = strings.Join(parts, ".")
	return
}

func QualifiedName(fields ...string) string {
	var o []string
	for _, f := range fields {
		c := QuoteIdentifier(f)
		o = append(o, c)
	}

	q := strings.Join(o[:], ".")
	return q
}
