package frontegg

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/MaterializeInc/terraform-provider-materialize/pkg/clients"
	"github.com/stretchr/testify/assert"
)

func TestFetchSCIM2ConfigurationsSuccess(t *testing.T) {
	assert := assert.New(t)

	// Mock server to return a sample response
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`[{"id":"test-id","source":"okta","tenantId":"test-tenant","connectionName":"test-conn","syncToUserManagement":true,"createdAt":"2022-01-01T00:00:00Z"}]`))
	}))
	defer mockServer.Close()

	client := &clients.FronteggClient{
		Endpoint:   mockServer.URL,
		HTTPClient: mockServer.Client(),
	}

	configurations, err := FetchSCIM2Configurations(context.Background(), client)
	assert.NoError(err)
	assert.Len(configurations, 1)
	assert.Equal("test-id", configurations[0].ID)
	assert.Equal("okta", configurations[0].Source)
	assert.Equal("test-tenant", configurations[0].TenantID)
}

func TestCreateSCIM2ConfigurationSuccess(t *testing.T) {
	assert := assert.New(t)

	// Mock server to return a sample response for the creation endpoint
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal("POST", r.Method)
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(`{"id":"new-test-id","source":"okta","connectionName":"test-conn","syncToUserManagement":true,"token":"test-token"}`))
	}))
	defer mockServer.Close()

	client := &clients.FronteggClient{
		Endpoint:   mockServer.URL,
		HTTPClient: mockServer.Client(),
	}

	newConfig, err := CreateSCIM2Configuration(context.Background(), client, SCIM2Configuration{
		Source:         "okta",
		ConnectionName: "test-conn",
	})
	assert.NoError(err)
	assert.NotNil(newConfig)
	assert.Equal("new-test-id", newConfig.ID)
	assert.Equal("test-token", newConfig.Token)
	assert.Equal("okta", newConfig.Source)
	assert.Equal("test-conn", newConfig.ConnectionName)
}

func TestDeleteSCIM2ConfigurationSuccess(t *testing.T) {
	assert := assert.New(t)

	// Mock server to simulate deletion
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal("DELETE", r.Method)
		w.WriteHeader(http.StatusNoContent)
	}))
	defer mockServer.Close()

	client := &clients.FronteggClient{
		Endpoint:   mockServer.URL,
		HTTPClient: mockServer.Client(),
	}

	err := DeleteSCIM2Configuration(context.Background(), client, "test-id")
	assert.NoError(err)
}
