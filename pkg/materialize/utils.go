package materialize

import (
	"strings"
)

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

func GetSliceValueString(v []interface{}) []string {
	var o []string
	for _, i := range v {
		o = append(o, i.(string))
	}
	return o
}

func GetSliceValueInt(v []interface{}) []int {
	var o []int
	for _, i := range v {
		o = append(o, i.(int))
	}
	return o
}
