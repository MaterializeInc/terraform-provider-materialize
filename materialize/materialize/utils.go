package materialize

import (
	"fmt"
	"strings"
)

func QuoteString(input string) (output string) {
	output = "'" + strings.Replace(input, "'", "''", -1) + "'"
	return
}

func QuoteIdentifier(input string) (output string) {
	output = `"` + strings.Replace(input, `"`, `""`, -1) + `"`
	return
}

func QualifiedName(fields ...string) string {
	var o []string
	for _, f := range fields {
		c := fmt.Sprintf(`"%v"`, f)
		o = append(o, c)
	}

	q := strings.Join(o[:], ".")
	return q
}
