package testhelpers

import (
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/MaterializeInc/terraform-provider-materialize/pkg/clients"
	"github.com/MaterializeInc/terraform-provider-materialize/pkg/utils"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWithMockDb(t *testing.T) {
	WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		rows := sqlmock.NewRows([]string{"number"}).AddRow(1)
		mock.ExpectQuery("^SELECT 1$").WillReturnRows(rows)
		var result int
		err := db.Get(&result, "SELECT 1")
		assert.NoError(t, err)
		assert.Equal(t, 1, result)
		err = mock.ExpectationsWereMet()
		assert.NoError(t, err)
	})
}

func TestWithMockProviderMeta(t *testing.T) {
	WithMockProviderMeta(t, func(providerMeta *utils.ProviderMeta, mock sqlmock.Sqlmock) {
		// Assert that the providerMeta is not nil
		assert.NotNil(t, providerMeta)

		// Assert that the providerMeta has the expected default region
		assert.Equal(t, clients.AwsUsEast1, providerMeta.DefaultRegion)

		// Assert that the providerMeta has the expected regions enabled
		assert.True(t, providerMeta.RegionsEnabled[clients.AwsUsEast1])

		// Optionally, set up mock expectations if you are going to perform any database operations
		mock.ExpectQuery("SELECT VERSION()").WillReturnRows(sqlmock.NewRows([]string{"version"}).AddRow("test-version"))

		// Perform a database operation using providerMeta.DB[clients.AwsUsEast1]
		var version string
		err := providerMeta.DB[clients.AwsUsEast1].DB.Get(&version, "SELECT VERSION()")
		require.NoError(t, err)
		assert.Equal(t, "test-version", version)

		// Ensure all expectations are met
		err = mock.ExpectationsWereMet()
		assert.NoError(t, err)
	})
}
