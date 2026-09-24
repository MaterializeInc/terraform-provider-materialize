package resources

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/MaterializeInc/terraform-provider-materialize/pkg/clients"
	"github.com/MaterializeInc/terraform-provider-materialize/pkg/testhelpers"
	"github.com/MaterializeInc/terraform-provider-materialize/pkg/utils"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/require"
)

func TestScimGroupUsersCreate(t *testing.T) {
	r := require.New(t)

	in := map[string]interface{}{
		"group_id": "test-group-id",
		"users":    []interface{}{"user1", "user2"},
	}
	d := schema.TestResourceDataRaw(t, ScimGroupUsersSchema, in)
	r.NotNil(d)

	testhelpers.WithMockFronteggServer(t, func(serverURL string) {
		client := &clients.FronteggClient{
			Endpoint:   serverURL,
			HTTPClient: &http.Client{},
		}

		providerMeta := &utils.ProviderMeta{
			Frontegg: client,
		}

		if err := scimGroupUsersCreate(context.TODO(), d, providerMeta); err != nil {
			t.Fatal(err)
		}

		// Assertions to check the state after create
		r.Equal("test-group-id", d.Id())
	})
}

func TestScimGroupUsersRead(t *testing.T) {
	r := require.New(t)

	testhelpers.WithMockFronteggServer(t, func(serverURL string) {
		client := &clients.FronteggClient{
			Endpoint:   serverURL,
			HTTPClient: &http.Client{},
		}

		providerMeta := &utils.ProviderMeta{
			Frontegg: client,
		}

		d := schema.TestResourceDataRaw(t, ScimGroupUsersSchema, nil)
		d.SetId("mock-group-id")

		if err := scimGroupUsersRead(context.TODO(), d, providerMeta); err != nil {
			t.Fatal(err)
		}

		r.Equal("mock-group-id", d.Id())
	})
}

func TestScimGroupUsersUpdate(t *testing.T) {

	testhelpers.WithMockFronteggServer(t, func(serverURL string) {
		client := &clients.FronteggClient{
			Endpoint:   serverURL,
			HTTPClient: &http.Client{},
		}

		providerMeta := &utils.ProviderMeta{
			Frontegg: client,
		}

		d := schema.TestResourceDataRaw(t, ScimGroupUsersSchema, nil)
		d.SetId("mock-group-id")
		d.Set("users", []interface{}{"user1", "user2"})

		if err := scimGroupUsersUpdate(context.TODO(), d, providerMeta); err != nil {
			t.Fatal(err)
		}

		// Assertions to check the state after update
	})
}

func TestScimGroupUsersDelete(t *testing.T) {
	r := require.New(t)

	testhelpers.WithMockFronteggServer(t, func(serverURL string) {
		client := &clients.FronteggClient{
			Endpoint:   serverURL,
			HTTPClient: &http.Client{},
		}

		providerMeta := &utils.ProviderMeta{
			Frontegg: client,
		}

		d := schema.TestResourceDataRaw(t, ScimGroupUsersSchema, nil)
		d.SetId("mock-group-id")
		d.Set("users", []interface{}{"user1", "user2"})

		if err := scimGroupUsersDelete(context.TODO(), d, providerMeta); err != nil {
			t.Fatal(err)
		}

		// Assertions to check the state after delete
		r.Empty(d.Id())
	})
}

func groupUsersStatusServer(status int) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(status)
	}))
}

func scimGroupUsersMeta(srv *httptest.Server) *utils.ProviderMeta {
	return &utils.ProviderMeta{Frontegg: &clients.FronteggClient{Endpoint: srv.URL, HTTPClient: srv.Client()}}
}

func scimGroupUsersData(t *testing.T) *schema.ResourceData {
	d := schema.TestResourceDataRaw(t, ScimGroupUsersSchema, map[string]interface{}{
		"group_id": "gone-group",
		"users":    []interface{}{"u1"},
	})
	d.SetId("gone-group")
	return d
}

func TestScimGroupUsersReadGoneRemovesFromState(t *testing.T) {
	r := require.New(t)
	srv := groupUsersStatusServer(http.StatusNotFound)
	defer srv.Close()

	d := scimGroupUsersData(t)
	r.False(scimGroupUsersRead(context.TODO(), d, scimGroupUsersMeta(srv)).HasError())
	r.Empty(d.Id())
}

func TestScimGroupUsersReadErrorKeepsState(t *testing.T) {
	r := require.New(t)
	srv := groupUsersStatusServer(http.StatusInternalServerError)
	defer srv.Close()

	d := scimGroupUsersData(t)
	r.True(scimGroupUsersRead(context.TODO(), d, scimGroupUsersMeta(srv)).HasError())
	r.Equal("gone-group", d.Id())
}

func TestScimGroupUsersUpdateGoneReportsMissingGroup(t *testing.T) {
	r := require.New(t)
	srv := groupUsersStatusServer(http.StatusNotFound)
	defer srv.Close()

	d := scimGroupUsersData(t)
	diags := scimGroupUsersUpdate(context.TODO(), d, scimGroupUsersMeta(srv))
	r.True(diags.HasError())
	r.Contains(diags[0].Summary, "does not exist")
	r.Empty(d.Id())
}

func TestScimGroupUsersDeleteGoneSucceeds(t *testing.T) {
	r := require.New(t)
	srv := groupUsersStatusServer(http.StatusNotFound)
	defer srv.Close()

	d := scimGroupUsersData(t)
	r.False(scimGroupUsersDelete(context.TODO(), d, scimGroupUsersMeta(srv)).HasError())
	r.Empty(d.Id())
}
