package materialize

import (
	"fmt"
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

func GetSliceValueString(attrName string, v []interface{}) ([]string, error) {
	var o []string
	for _, item := range v {
		str, ok := item.(string)
		if !ok {
			return nil, fmt.Errorf("value %v of attribute %s cannot be converted to string", item, attrName)
		}
		o = append(o, str)
	}
	return o, nil
}

func GetSliceValueInt(v []interface{}) []int {
	var o []int
	for _, i := range v {
		o = append(o, i.(int))
	}
	return o
}
