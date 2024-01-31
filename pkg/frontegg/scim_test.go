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
