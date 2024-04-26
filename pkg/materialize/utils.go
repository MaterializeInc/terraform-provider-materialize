package materialize

import (
	"strings"
)

func QuoteString(input string) string {
	return "'" + strings.Replace(input, "'", "''", -1) + "'"
}

func QuoteIdentifier(input string) string {
	return `"` + strings.Replace(input, `"`, `""`, -1) + `"`
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
		if i != nil {
			str, ok := i.(string)
			if ok {
				o = append(o, str)
			}
		}
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
