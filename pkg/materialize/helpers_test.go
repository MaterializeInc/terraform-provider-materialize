package materialize

import (
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/require"
)

// withMockDb is testhelpers.WithMockDb for this package's own tests, which
// cannot import testhelpers since it imports this package.
func withMockDb(t *testing.T, f func(*sqlx.DB, sqlmock.Sqlmock)) {
	t.Helper()
	r := require.New(t)
	db, mock, err := sqlmock.New()
	r.NoError(err)
	dbx := sqlx.NewDb(db, "sqlmock")
	defer dbx.Close()

	mock.MatchExpectationsInOrder(true)

	f(dbx, mock)
}
