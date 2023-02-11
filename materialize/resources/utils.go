package resources

import (
	"fmt"
	"log"
	"reflect"

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

func createResource(conn *sqlx.DB, d *schema.ResourceData, queryCreateStr, queryReadStr string) error {
	_, errr := conn.Exec(queryCreateStr)
	if errr != nil {
		log.Printf("[ERROR] could not execute query: %s", queryCreateStr)
		return &SQLError{Err: errr}
	}

	var i string
	err := conn.QueryRow(queryReadStr).Scan(&i)
	if err != nil {
		log.Printf("[ERROR] could not execute query: %s", queryReadStr)
		return &SQLError{Err: err}
	}

	d.SetId(i)
	return nil
}

func readResource(conn *sqlx.DB, d *schema.ResourceData, id, queryStr string, resourceStruct interface{}) error {
	err := conn.QueryRowx(queryStr).StructScan(resourceStruct)
	if err != nil {
		return &SQLError{Err: err}
	}

	d.SetId(id)

	values := reflect.ValueOf(resourceStruct)
	types := values.Type()
	for i := 0; i < values.NumField(); i++ {
		d.Set(types.Field(i).Name, values.Field(i))
	}

	return nil
}

func dropResource(conn *sqlx.DB, d *schema.ResourceData, queryStr string) error {
	_, errr := conn.Exec(queryStr)
	if errr != nil {
		log.Printf("[ERROR] could not execute query: %s", queryStr)
		return &SQLError{Err: errr}
	}

	d.SetId("")
	return nil
}
