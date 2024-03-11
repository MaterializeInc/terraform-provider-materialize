package frontegg

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/MaterializeInc/terraform-provider-materialize/pkg/clients"
	"github.com/stretchr/testify/assert"
)

func TestFetchSSOConfigurationsSuccess(t *testing.T) {
	assert := assert.New(t)

	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`[{"id":"test-id","enabled":true,"ssoEndpoint":"test-endpoint","publicCertificate":"test-cert","signRequest":true,"type":"saml"}]`))
	}))
	defer mockServer.Close()

	client := &clients.FronteggClient{
		Endpoint:   mockServer.URL,
		HTTPClient: mockServer.Client(),
	}

	configurations, err := FetchSSOConfigurations(context.Background(), client)
	assert.NoError(err)
	assert.Len(configurations, 1)
	assert.Equal("test-id", configurations[0].Id)
	assert.True(configurations[0].Enabled)
	assert.Equal("test-endpoint", configurations[0].SsoEndpoint)
	assert.Equal("test-cert", configurations[0].PublicCertificate)
	assert.True(configurations[0].SignRequest)
	assert.Equal("saml", configurations[0].Type)
	assert.Empty(configurations[0].Domains)
}

func TestCreateSSOConfigurationSuccess(t *testing.T) {
	assert := assert.New(t)

	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(`{"id":"test-id","enabled":true,"ssoEndpoint":"test-endpoint","publicCertificate":"test-cert","signRequest":true,"type":"saml"}`))
	}))
	defer mockServer.Close()

	client := &clients.FronteggClient{
		Endpoint:   mockServer.URL,
		HTTPClient: mockServer.Client(),
	}

	ssoConfig := SSOConfig{
		Enabled:           true,
		SsoEndpoint:       "test-endpoint",
		PublicCertificate: "test-cert",
		SignRequest:       true,
		Type:              "saml",
	}

	newConfig, err := CreateSSOConfiguration(context.Background(), client, ssoConfig)
	assert.NoError(err)
	assert.Equal("test-id", newConfig.Id)
	assert.True(newConfig.Enabled)
	assert.Equal("test-endpoint", newConfig.SsoEndpoint)
	assert.Equal("test-cert", newConfig.PublicCertificate)
	assert.True(newConfig.SignRequest)
	assert.Equal("saml", newConfig.Type)
	assert.Empty(newConfig.Domains)
}

func TestUpdateSSOConfigurationSuccess(t *testing.T) {
	assert := assert.New(t)

	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"id":"test-id","enabled":true,"ssoEndpoint":"updated-endpoint","publicCertificate":"test-cert","signRequest":true,"type":"saml"}`))
	}))
	defer mockServer.Close()

	client := &clients.FronteggClient{
		Endpoint:   mockServer.URL,
		HTTPClient: mockServer.Client(),
	}

	ssoConfig := SSOConfig{
		Id:                "test-id",
		Enabled:           true,
		SsoEndpoint:       "updated-endpoint",
		PublicCertificate: "test-cert",
		SignRequest:       true,
		Type:              "saml",
	}

	updatedConfig, err := UpdateSSOConfiguration(context.Background(), client, ssoConfig)
	assert.NoError(err)
	assert.Equal("updated-endpoint", updatedConfig.SsoEndpoint)
}

func TestDeleteSSOConfigurationSuccess(t *testing.T) {
	assert := assert.New(t)

	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer mockServer.Close()

	client := &clients.FronteggClient{
		Endpoint:   mockServer.URL,
		HTTPClient: mockServer.Client(),
	}

	err := DeleteSSOConfiguration(context.Background(), client, "test-id")
	assert.NoError(err)
}
