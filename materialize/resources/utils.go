package resources

import (
	"fmt"
	"log"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/jmoiron/sqlx"
)

type SQLError struct {
	Err error
}

func (e *SQLError) Error() string {
	return fmt.Sprintf("Unable to execute SQL: %v", e.Err)
}

func ExecResource(conn *sqlx.DB, queryStr string) error {
	_, err := conn.Exec(queryStr)
	if err != nil {
		return &SQLError{Err: err}
	}

	return nil
}

func createResource(conn *sqlx.DB, d *schema.ResourceData, queryCreateStr, queryReadStr, resource string) error {
	_, errr := conn.Exec(queryCreateStr)
	if errr != nil {
		log.Printf("[ERROR] could not create %s: %s", resource, queryCreateStr)
		return &SQLError{Err: errr}
	}

	var i string
	err := conn.QueryRow(queryReadStr).Scan(&i)
	if err != nil {
		log.Printf("[ERROR] could not read %s id", resource)
		return &SQLError{Err: err}
	}

	d.SetId(i)
	return nil
}

func dropResource(conn *sqlx.DB, d *schema.ResourceData, queryStr, resource string) error {
	_, errr := conn.Exec(queryStr)
	if errr != nil {
		log.Printf("[ERROR] could not drop %s: %s", resource, queryStr)
		return &SQLError{Err: errr}
	}

	// Explicit set id to empty
	d.SetId("")
	return nil
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

func QuoteIdentifierName(input string) (output string) {
	output = `"` + strings.Replace(input, `"`, `""`, -1) + `"`
	return
}

func QualifiedName(fields ...string) string {
	var o []string
	for _, f := range fields {
		c := fmt.Sprintf(`%v`, QuoteIdentifierName(f))
		o = append(o, c)
	}

	q := strings.Join(o[:], ".")
	return q
}
