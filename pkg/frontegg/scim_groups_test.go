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

func TestFetchSCIMGroupsSuccess(t *testing.T) {
	assert := assert.New(t)

	// Mock server to return a sample SCIM group response
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"groups":[{"id":"group-id","name":"test-group","description":"A test group","metadata":"test-metadata","roles":[{"id":"role-id","key":"test-key","name":"test-role","description":"A test role","is_default":true}],"users":[{"id":"user-id","name":"test-user","email":"test@user.com"}],"managedBy":"manager-id"}]}`))
	}))
	defer mockServer.Close()

	client := &clients.FronteggClient{
		Endpoint:   mockServer.URL,
		HTTPClient: mockServer.Client(),
	}

	groups, err := FetchSCIMGroups(context.Background(), client)
	assert.NoError(err)
	assert.Len(groups, 1)
	assert.Equal("group-id", groups[0].ID)
	assert.Equal("test-group", groups[0].Name)
	assert.Equal("A test group", groups[0].Description)
	assert.Equal("test-metadata", groups[0].Metadata)
	assert.Len(groups[0].Roles, 1)
	assert.Len(groups[0].Users, 1)
}

func TestFetchSCIMGroupsFailure(t *testing.T) {
	assert := assert.New(t)

	// Mock server to return an error
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Internal Server Error"))
	}))
	defer mockServer.Close()

	client := &clients.FronteggClient{
		Endpoint:   mockServer.URL,
		HTTPClient: mockServer.Client(),
	}

	_, err := FetchSCIMGroups(context.Background(), client)
	assert.Error(err)
}

func TestCreateSCIMGroupSuccess(t *testing.T) {
	assert := assert.New(t)

	// Mock server to return a successful response for group creation
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal("POST", r.Method, "Expected POST request")
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(`{"id":"new-group-id","name":"New Group","description":"A new test group","metadata":"new-metadata"}`))
	}))
	defer mockServer.Close()

	client := &clients.FronteggClient{
		Endpoint:   mockServer.URL,
		HTTPClient: mockServer.Client(),
	}

	params := GroupCreateParams{
		Name:        "New Group",
		Description: "A new test group",
		Metadata:    "new-metadata",
	}

	group, err := CreateSCIMGroup(context.Background(), client, params)
	assert.NoError(err)
	assert.NotNil(group)
	assert.Equal("new-group-id", group.ID)
	assert.Equal("New Group", group.Name)
	assert.Equal("A new test group", group.Description)
	assert.Equal("new-metadata", group.Metadata)
}

func TestCreateSCIMGroupFailure(t *testing.T) {
	assert := assert.New(t)

	// Mock server to return an error response for group creation
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal("POST", r.Method, "Expected POST request")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Internal Server Error"))
	}))
	defer mockServer.Close()

	client := &clients.FronteggClient{
		Endpoint:   mockServer.URL,
		HTTPClient: mockServer.Client(),
	}

	params := GroupCreateParams{
		Name:        "New Group",
		Description: "A new test group",
		Metadata:    "new-metadata",
	}

	_, err := CreateSCIMGroup(context.Background(), client, params)
	assert.Error(err)
}

func TestUpdateSCIMGroupSuccess(t *testing.T) {
	assert := assert.New(t)

	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal("PATCH", r.Method, "Expected PATCH request")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"id":"group-id","name":"Updated Group","description":"Updated description","metadata":"updated-metadata"}`))
	}))
	defer mockServer.Close()

	client := &clients.FronteggClient{
		Endpoint:   mockServer.URL,
		HTTPClient: mockServer.Client(),
	}

	params := GroupUpdateParams{
		Name:        "Updated Group",
		Description: "Updated description",
		Metadata:    "updated-metadata",
	}

	err := UpdateSCIMGroup(context.Background(), client, "group-id", params)
	assert.NoError(err)
}

func TestUpdateSCIMGroupFailure(t *testing.T) {
	assert := assert.New(t)

	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal("PATCH", r.Method, "Expected PATCH request")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Internal Server Error"))
	}))
	defer mockServer.Close()

	client := &clients.FronteggClient{
		Endpoint:   mockServer.URL,
		HTTPClient: mockServer.Client(),
	}

	params := GroupUpdateParams{
		Name:        "Updated Group",
		Description: "Updated description",
		Metadata:    "updated-metadata",
	}

	err := UpdateSCIMGroup(context.Background(), client, "group-id", params)
	assert.Error(err)
}

func TestDeleteSCIMGroupSuccess(t *testing.T) {
	assert := assert.New(t)

	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal("DELETE", r.Method, "Expected DELETE request")
		w.WriteHeader(http.StatusOK)
	}))
	defer mockServer.Close()

	client := &clients.FronteggClient{
		Endpoint:   mockServer.URL,
		HTTPClient: mockServer.Client(),
	}

	err := DeleteSCIMGroup(context.Background(), client, "group-id")
	assert.NoError(err)
}

func TestDeleteSCIMGroupFailure(t *testing.T) {
	assert := assert.New(t)

	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal("DELETE", r.Method, "Expected DELETE request")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Internal Server Error"))
	}))
	defer mockServer.Close()

	client := &clients.FronteggClient{
		Endpoint:   mockServer.URL,
		HTTPClient: mockServer.Client(),
	}

	err := DeleteSCIMGroup(context.Background(), client, "group-id")
	assert.Error(err)
}

func TestGetSCIMGroupByIDSuccess(t *testing.T) {
	assert := assert.New(t)

	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal("GET", r.Method, "Expected GET request")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"id":"group-id","name":"test-group","description":"A test group","metadata":"test-metadata","roles":[],"users":[],"managedBy":"manager-id"}`))
	}))
	defer mockServer.Close()

	client := &clients.FronteggClient{
		Endpoint:   mockServer.URL,
		HTTPClient: mockServer.Client(),
	}

	group, err := GetSCIMGroupByID(context.Background(), client, "group-id")
	assert.NoError(err)
	assert.NotNil(group)
	assert.Equal("group-id", group.ID)
	assert.Equal("test-group", group.Name)
	assert.Equal("A test group", group.Description)
	assert.Equal("test-metadata", group.Metadata)
}

func TestGetSCIMGroupByIDFailure(t *testing.T) {
	assert := assert.New(t)

	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal("GET", r.Method, "Expected GET request")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Internal Server Error"))
	}))
	defer mockServer.Close()

	client := &clients.FronteggClient{
		Endpoint:   mockServer.URL,
		HTTPClient: mockServer.Client(),
	}

	_, err := GetSCIMGroupByID(context.Background(), client, "group-id")
	assert.Error(err)
}

func TestAddRolesToGroupSuccess(t *testing.T) {
	assert := assert.New(t)

	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal("POST", r.Method, "Expected POST request")
		assert.Equal("/frontegg/identity/resources/groups/v1/test-group-id/roles", r.URL.Path)
		w.WriteHeader(http.StatusCreated)
	}))
	defer mockServer.Close()

	client := &clients.FronteggClient{
		Endpoint:   mockServer.URL,
		HTTPClient: mockServer.Client(),
	}

	err := AddRolesToGroup(context.Background(), client, "test-group-id", []string{"role1", "role2"})
	assert.NoError(err)
}

func TestAddRolesToGroupFailure(t *testing.T) {
	assert := assert.New(t)

	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal("POST", r.Method, "Expected POST request")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Internal Server Error"))
	}))
	defer mockServer.Close()

	client := &clients.FronteggClient{
		Endpoint:   mockServer.URL,
		HTTPClient: mockServer.Client(),
	}

	err := AddRolesToGroup(context.Background(), client, "test-group-id", []string{"role1", "role2"})
	assert.Error(err)
}

func TestRemoveRolesFromGroupSuccess(t *testing.T) {
	assert := assert.New(t)

	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal("DELETE", r.Method, "Expected DELETE request")
		assert.Equal("/frontegg/identity/resources/groups/v1/test-group-id/roles", r.URL.Path)
		w.WriteHeader(http.StatusOK)
	}))
	defer mockServer.Close()

	client := &clients.FronteggClient{
		Endpoint:   mockServer.URL,
		HTTPClient: mockServer.Client(),
	}

	err := RemoveRolesFromGroup(context.Background(), client, "test-group-id", []string{"role1", "role2"})
	assert.NoError(err)
}

func TestRemoveRolesFromGroupFailure(t *testing.T) {
	assert := assert.New(t)

	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal("DELETE", r.Method, "Expected DELETE request")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Internal Server Error"))
	}))
	defer mockServer.Close()

	client := &clients.FronteggClient{
		Endpoint:   mockServer.URL,
		HTTPClient: mockServer.Client(),
	}

	err := RemoveRolesFromGroup(context.Background(), client, "test-group-id", []string{"role1", "role2"})
	assert.Error(err)
}

func TestAddUsersToGroupSuccess(t *testing.T) {
	assert := assert.New(t)

	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal("POST", r.Method, "Expected POST request")
		assert.Equal("/frontegg/identity/resources/groups/v1/group-id/users", r.URL.Path)

		var body struct {
			UserIds []string `json:"userIds"`
		}
		err := json.NewDecoder(r.Body).Decode(&body)
		assert.NoError(err)
		assert.ElementsMatch([]string{"user1", "user2"}, body.UserIds)

		w.WriteHeader(http.StatusCreated)
	}))
	defer mockServer.Close()

	client := &clients.FronteggClient{
		Endpoint:   mockServer.URL,
		HTTPClient: mockServer.Client(),
	}

	err := AddUsersToGroup(context.Background(), client, "group-id", []string{"user1", "user2"})
	assert.NoError(err)
}

func TestAddUsersToGroupFailure(t *testing.T) {
	assert := assert.New(t)

	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Internal Server Error"))
	}))
	defer mockServer.Close()

	client := &clients.FronteggClient{
		Endpoint:   mockServer.URL,
		HTTPClient: mockServer.Client(),
	}

	err := AddUsersToGroup(context.Background(), client, "group-id", []string{"user1", "user2"})
	assert.Error(err)
}

func TestRemoveUsersFromGroupSuccess(t *testing.T) {
	assert := assert.New(t)

	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal("DELETE", r.Method, "Expected DELETE request")
		assert.Equal("/frontegg/identity/resources/groups/v1/group-id/users", r.URL.Path)

		var body struct {
			UserIds []string `json:"userIds"`
		}
		err := json.NewDecoder(r.Body).Decode(&body)
		assert.NoError(err)
		assert.ElementsMatch([]string{"user1", "user2"}, body.UserIds)

		w.WriteHeader(http.StatusOK)
	}))
	defer mockServer.Close()

	client := &clients.FronteggClient{
		Endpoint:   mockServer.URL,
		HTTPClient: mockServer.Client(),
	}

	err := RemoveUsersFromGroup(context.Background(), client, "group-id", []string{"user1", "user2"})
	assert.NoError(err)
}

func TestRemoveUsersFromGroupFailure(t *testing.T) {
	assert := assert.New(t)

	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Internal Server Error"))
	}))
	defer mockServer.Close()

	client := &clients.FronteggClient{
		Endpoint:   mockServer.URL,
		HTTPClient: mockServer.Client(),
	}

	err := RemoveUsersFromGroup(context.Background(), client, "group-id", []string{"user1", "user2"})
	assert.Error(err)
}
