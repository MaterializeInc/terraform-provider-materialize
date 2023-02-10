package resources

import (
	"database/sql"
	"fmt"
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
		return &SQLError{Err: err}
	}

	return nil
}
