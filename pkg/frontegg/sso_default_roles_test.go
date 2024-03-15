package frontegg

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/MaterializeInc/terraform-provider-materialize/pkg/clients"
	"github.com/stretchr/testify/assert"
)

func TestListSSORolesSuccess(t *testing.T) {
	assert := assert.New(t)

	rolesResponse := FronteggRolesResponse{
		Items: []FronteggRole{
			{ID: "role-id-1", Name: "Organization Admin"},
			{ID: "role-id-2", Name: "Organization Member"},
		},
	}

	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(http.MethodGet, r.Method)
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(rolesResponse)
	}))
	defer mockServer.Close()

	client := &clients.FronteggClient{
		Endpoint:   mockServer.URL,
		HTTPClient: mockServer.Client(),
	}

	roles, err := ListSSORoles(context.Background(), client)
	assert.NoError(err)
	assert.Equal(2, len(roles))
	assert.Equal("role-id-1", roles["Admin"])
	assert.Equal("role-id-2", roles["Member"])
}

func TestSetSSODefaultRolesSuccess(t *testing.T) {
	assert := assert.New(t)

	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(http.MethodPut, r.Method)
		w.WriteHeader(http.StatusCreated)
	}))
	defer mockServer.Close()

	client := &clients.FronteggClient{
		Endpoint:   mockServer.URL,
		HTTPClient: mockServer.Client(),
	}

	err := SetSSODefaultRoles(context.Background(), client, "config-id", []string{"role-id-1", "role-id-2"})
	assert.NoError(err)
}

func TestGetSSODefaultRolesSuccess(t *testing.T) {
	assert := assert.New(t)

	rolesResponse := RoleIDs{
		RoleIds: []string{"role-id-1", "role-id-2"},
	}

	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(http.MethodGet, r.Method)
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(rolesResponse)
	}))
	defer mockServer.Close()

	client := &clients.FronteggClient{
		Endpoint:   mockServer.URL,
		HTTPClient: mockServer.Client(),
	}

	roleIDs, err := GetSSODefaultRoles(context.Background(), client, "config-id")
	assert.NoError(err)
	assert.Equal(2, len(roleIDs))
	assert.Contains(roleIDs, "role-id-1")
	assert.Contains(roleIDs, "role-id-2")
}

func TestClearSSODefaultRolesSuccess(t *testing.T) {
	assert := assert.New(t)

	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(http.MethodPut, r.Method)

		var payload RoleIDs
		err := json.NewDecoder(r.Body).Decode(&payload)
		assert.NoError(err)
		assert.Empty(payload.RoleIds)

		w.WriteHeader(http.StatusCreated)
	}))
	defer mockServer.Close()

	client := &clients.FronteggClient{
		Endpoint:   mockServer.URL,
		HTTPClient: mockServer.Client(),
	}

	err := ClearSSODefaultRoles(context.Background(), client, "config-id")
	assert.NoError(err)
}
