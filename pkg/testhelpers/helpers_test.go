package testhelpers

import (
	"encoding/json"
	"net/http"
	"strings"
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

		err = mock.ExpectationsWereMet()
		assert.NoError(t, err)
	})
}

func TestWithMockFronteggServer(t *testing.T) {
	t.Run("TestWithMockFronteggServer_PostRequest", func(t *testing.T) {
		WithMockFronteggServer(t, func(url string) {
			// Perform HTTP POST request to the mock server and assert the response
			req, err := http.NewRequest(http.MethodPost, url+"/identity/resources/users/api-tokens/v1", strings.NewReader(`{"description":"test-description"}`))
			require.NoError(t, err)

			resp, err := http.DefaultClient.Do(req)
			require.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, http.StatusCreated, resp.StatusCode)

			var appPassword MockAppPassword
			err = json.NewDecoder(resp.Body).Decode(&appPassword)
			require.NoError(t, err)

			assert.Equal(t, "mock-client-id", appPassword.ClientID)
			assert.Equal(t, "test-description", appPassword.Description)
			assert.Equal(t, "mockOwner", appPassword.Owner)
			assert.NotNil(t, appPassword.CreatedAt)
			assert.Equal(t, "mock-secret", appPassword.Secret)
		})
	})

	t.Run("TestWithMockFronteggServer_GetRequest", func(t *testing.T) {
		WithMockFronteggServer(t, func(url string) {
			resp, err := http.Get(url + "/identity/resources/users/api-tokens/v1")
			require.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, http.StatusOK, resp.StatusCode)

			var appPasswords []MockAppPassword
			err = json.NewDecoder(resp.Body).Decode(&appPasswords)
			require.NoError(t, err)

			require.Len(t, appPasswords, 1)
			appPassword := appPasswords[0]
			assert.Equal(t, "mock-client-id", appPassword.ClientID)
			assert.Equal(t, "test-app-password", appPassword.Description)
			assert.Equal(t, "mockOwner", appPassword.Owner)
			assert.NotNil(t, appPassword.CreatedAt)
			assert.Equal(t, "mock-secret", appPassword.Secret)
		})
	})
}

func TestMockCloudService_RoundTrip(t *testing.T) {
	t.Run("TestMockCloudService_RoundTrip_ValidURL", func(t *testing.T) {
		mockService := &MockCloudService{}
		req, err := http.NewRequest(http.MethodGet, "http://example.com/api/cloud-regions", nil)
		assert.NoError(t, err)

		resp, err := mockService.RoundTrip(req)
		assert.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})
}

func TestCopyHeaders(t *testing.T) {
	t.Run("TestCopyHeaders", func(t *testing.T) {
		dstHeader := make(http.Header)
		srcHeader := http.Header{
			"Content-Type":  []string{"application/json"},
			"Authorization": []string{"Bearer token"},
		}

		copyHeaders(dstHeader, srcHeader)
		assert.Equal(t, "application/json", dstHeader.Get("Content-Type"))
		assert.Equal(t, "Bearer token", dstHeader.Get("Authorization"))
	})
}
