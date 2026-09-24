package frontegg

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/MaterializeInc/terraform-provider-materialize/pkg/clients"
	"github.com/stretchr/testify/assert"
)

func TestListFronteggRolesSuccess(t *testing.T) {
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

	roles, err := ListFronteggRoles(context.Background(), client)
	assert.NoError(err)
	assert.Equal(2, len(roles))
	assert.Equal([]string{"role-id-1"}, roles["Admin"])
	assert.Equal([]string{"role-id-2"}, roles["Member"])
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

func TestListFronteggRolesIncludesCustomRolesAcrossPages(t *testing.T) {
	pages := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := FronteggRolesResponse{}
		response.Metadata.TotalPages = 2
		if r.URL.Query().Get("_offset") == "0" {
			response.Items = []FronteggRole{{ID: "member", Name: "Organization Member"}}
		} else {
			assert.Equal(t, "1", r.URL.Query().Get("_offset"))
			response.Items = []FronteggRole{{ID: "custom", Name: "Organization Analytics"}}
		}
		pages++
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()
	roles, err := ListFronteggRoles(context.Background(), &clients.FronteggClient{Endpoint: server.URL, HTTPClient: server.Client()})
	assert.NoError(t, err)
	assert.Equal(t, map[string][]string{"Member": {"member"}, "Organization Analytics": {"custom"}}, roles)
	assert.Equal(t, 2, pages)
}

func TestListFronteggRolesOnlyRejectsAmbiguousLookup(t *testing.T) {
	response := FronteggRolesResponse{Items: []FronteggRole{
		{ID: "built-in-admin", Name: "Organization Admin"},
		{ID: "custom-admin", Name: "Admin"},
		{ID: "member", Name: "Organization Member"},
	}}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.NoError(t, json.NewEncoder(w).Encode(response))
	}))
	defer server.Close()

	roles, err := ListFronteggRoles(context.Background(), &clients.FronteggClient{Endpoint: server.URL, HTTPClient: server.Client()})
	assert.NoError(t, err)
	assert.Equal(t, []string{"built-in-admin", "custom-admin"}, roles["Admin"])
	memberID, err := RoleIDByName(roles, "Member")
	assert.NoError(t, err)
	assert.Equal(t, "member", memberID)
	_, err = RoleIDByName(roles, "Admin")
	assert.ErrorContains(t, err, "ambiguous organization role name: Admin")
	name, found := RoleNameByID(roles, "custom-admin")
	assert.True(t, found)
	assert.Equal(t, "Admin", name)
}

func TestRoleIDByNameNotFoundIsSentinel(t *testing.T) {
	roles := map[string][]string{"Admin": {"a"}, "Dup": {"x", "y"}}

	_, err := RoleIDByName(roles, "Missing")
	assert.True(t, errors.Is(err, ErrRoleNotFound))
	assert.ErrorContains(t, err, "role not found: Missing")

	_, err = RoleIDByName(roles, "Dup")
	assert.False(t, errors.Is(err, ErrRoleNotFound), "ambiguous must not read as not found")
}
