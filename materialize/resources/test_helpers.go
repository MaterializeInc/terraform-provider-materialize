package resources

import (
	"database/sql"
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/require"
)

func WithMockDb(t *testing.T, f func(*sql.DB, sqlmock.Sqlmock)) {
	t.Helper()
	r := require.New(t)
	db, mock, err := sqlmock.New()
	r.NoError(err)
	defer db.Close()

	mock.MatchExpectationsInOrder(false)

	f(db, mock)
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("There were unfulfilled expectations: %s", err)
	}
}
