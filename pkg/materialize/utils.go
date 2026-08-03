package materialize

import (
	"database/sql/driver"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgtype"
)

// isUndefinedObject reports whether err is Postgres undefined_table or
// undefined_column. Materialize versions that predate a catalog view the
// provider reads fail this way, which lets callers degrade gracefully without
// also swallowing genuine query failures.
func isUndefinedObject(err error) bool {
	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) {
		return false
	}
	switch pgErr.SQLState() {
	case "42P01", "42703":
		return true
	default:
		return false
	}
}

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

// StringArray is a custom type that wraps []string and provides
// compatibility with PostgreSQL text[] arrays using pgx.
// It replaces pq.StringArray for pgx v4 compatibility.
type StringArray []string

// Scan implements the sql.Scanner interface for StringArray.
// It allows scanning PostgreSQL text[] arrays into a []string slice.
func (a *StringArray) Scan(src interface{}) error {
	var textArray pgtype.TextArray
	if err := textArray.Scan(src); err != nil {
		return err
	}

	if textArray.Status != pgtype.Present {
		*a = nil
		return nil
	}

	// Convert pgtype.TextArray elements to []string
	elements := make([]string, len(textArray.Elements))
	for i, elem := range textArray.Elements {
		if elem.Status == pgtype.Present {
			elements[i] = elem.String
		}
	}

	*a = elements
	return nil
}

// Value implements the driver.Valuer interface for StringArray.
// It allows converting a []string slice into a PostgreSQL text[] array.
func (a StringArray) Value() (driver.Value, error) {
	if a == nil {
		return nil, nil
	}

	var textArray pgtype.TextArray
	if err := textArray.Set([]string(a)); err != nil {
		return nil, err
	}

	return textArray.Value()
}
