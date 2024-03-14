package frontegg

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/MaterializeInc/terraform-provider-materialize/pkg/clients"
	"github.com/stretchr/testify/assert"
)

func TestFetchSSODomainSuccess(t *testing.T) {
	assert := assert.New(t)

	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		// Mimic the actual structure with configurations containing domains
		w.Write([]byte(`[
			{
				"id": "config-id",
				"domains": [
					{"id": "domain-id", "domain": "example.com", "validated": true}
				]
			}
		]`))
	}))
	defer mockServer.Close()

	client := &clients.FronteggClient{
		Endpoint:   mockServer.URL,
		HTTPClient: mockServer.Client(),
	}

	domain, err := FetchSSODomain(context.Background(), client, "config-id", "example.com")
	assert.NoError(err)
	assert.NotNil(domain)
	assert.Equal("domain-id", domain.ID)
	assert.Equal("example.com", domain.Domain)
	assert.True(domain.Validated)
}

func TestCreateSSODomainSuccess(t *testing.T) {
	assert := assert.New(t)

	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(`{"id":"new-domain-id","domain":"new-example.com","validated":false,"ssoConfigId":"config-id"}`))
	}))
	defer mockServer.Close()

	client := &clients.FronteggClient{
		Endpoint:   mockServer.URL,
		HTTPClient: mockServer.Client(),
	}

	domain, err := CreateSSODomain(context.Background(), client, "config-id", "new-example.com")
	assert.NoError(err)
	assert.NotNil(domain)
	assert.Equal("new-domain-id", domain.ID)
	assert.Equal("new-example.com", domain.Domain)
	assert.False(domain.Validated)
	assert.Equal("config-id", domain.SsoConfigId)
}

func TestDeleteSSODomainSuccess(t *testing.T) {
	assert := assert.New(t)

	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer mockServer.Close()

	client := &clients.FronteggClient{
		Endpoint:   mockServer.URL,
		HTTPClient: mockServer.Client(),
	}

	err := DeleteSSODomain(context.Background(), client, "config-id", "domain-id")
	assert.NoError(err)
}
