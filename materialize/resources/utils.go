package resources

import (
	"database/sql"
	"fmt"
	"log"
)

type SQLError struct {
	Err error
}

func (e *SQLError) Error() string {
	return fmt.Sprintf("Unable to execute SQL: %v", e.Err)
}

func ExecResource(conn *sql.DB, queryStr string) error {
	_, err := conn.Exec(queryStr)
	if err != nil {
		log.Printf("[ERROR] could not execute query: %s", queryStr)
		return &SQLError{Err: err}
	}

	return nil
}
