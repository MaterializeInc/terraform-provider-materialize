package frontegg

import (
	"context"
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
