package testhelpers

import (
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/MaterializeInc/terraform-provider-materialize/pkg/utils"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/require"
)

func WithMockDb(t *testing.T, f func(*sqlx.DB, sqlmock.Sqlmock)) {
	// Set the region for testing
	utils.Region = "aws/us-east-1"

	t.Helper()
	r := require.New(t)
	db, mock, err := sqlmock.New()
	dbx := sqlx.NewDb(db, "sqlmock")
	r.NoError(err)
	defer dbx.Close()

	mock.MatchExpectationsInOrder(true)

	f(dbx, mock)
}
