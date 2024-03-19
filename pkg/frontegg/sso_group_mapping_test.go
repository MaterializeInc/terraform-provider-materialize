package frontegg

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/MaterializeInc/terraform-provider-materialize/pkg/clients"
	"github.com/stretchr/testify/assert"
)

func TestCreateSSOGroupMappingSuccess(t *testing.T) {
	assert := assert.New(t)

	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal("POST", r.Method)
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(`{"id": "group-id", "group": "group-name", "enabled": true, "roleIds": ["role1", "role2"]}`))
	}))
	defer mockServer.Close()

	client := &clients.FronteggClient{
		Endpoint:   mockServer.URL,
		HTTPClient: mockServer.Client(),
	}

	groupMapping, err := CreateSSOGroupMapping(context.Background(), client, "sso-config-id", "group-name", []string{"role1", "role2"})
	assert.NoError(err)
	assert.NotNil(groupMapping)
	assert.Equal("group-id", groupMapping.ID)
	assert.Equal("group-name", groupMapping.Group)
	assert.Equal(true, groupMapping.Enabled)
	assert.ElementsMatch([]string{"role1", "role2"}, groupMapping.RoleIds)
}

func TestGetSSOGroupMappingsSuccess(t *testing.T) {
	assert := assert.New(t)

	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal("GET", r.Method)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`[{"id": "group-id", "group": "group-name", "enabled": true, "roleIds": ["role1", "role2"]}]`))
	}))
	defer mockServer.Close()

	client := &clients.FronteggClient{
		Endpoint:   mockServer.URL,
		HTTPClient: mockServer.Client(),
	}

	groups, err := GetSSOGroupMappings(context.Background(), client, "sso-config-id")
	assert.NoError(err)
	assert.Len(*groups, 1)
	assert.Equal("group-id", (*groups)[0].ID)
	assert.Equal("group-name", (*groups)[0].Group)
	assert.Equal(true, (*groups)[0].Enabled)
	assert.ElementsMatch([]string{"role1", "role2"}, (*groups)[0].RoleIds)
}

func TestFetchSSOGroupMappingSuccess(t *testing.T) {
	assert := assert.New(t)

	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal("GET", r.Method)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`[{"id": "group-id", "group": "group-name", "enabled": true, "roleIds": ["role1", "role2"]}]`))
	}))
	defer mockServer.Close()

	client := &clients.FronteggClient{
		Endpoint:   mockServer.URL,
		HTTPClient: mockServer.Client(),
	}

	groupMapping, err := FetchSSOGroupMapping(context.Background(), client, "sso-config-id", "group-id")
	assert.NoError(err)
	assert.NotNil(groupMapping)
	assert.Equal("group-id", groupMapping.ID)
	assert.Equal("group-name", groupMapping.Group)
	assert.Equal(true, groupMapping.Enabled)
	assert.ElementsMatch([]string{"role1", "role2"}, groupMapping.RoleIds)
}
