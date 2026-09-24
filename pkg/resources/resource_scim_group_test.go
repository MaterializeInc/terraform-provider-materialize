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

func TestScimGroupResourceCreate(t *testing.T) {
	r := require.New(t)

	in := map[string]interface{}{
		"name":        "Test Group",
		"description": "A test group description",
	}
	d := schema.TestResourceDataRaw(t, Scim2GroupSchema, in)
	r.NotNil(d)

	testhelpers.WithMockFronteggServer(t, func(serverURL string) {
		client := &clients.FronteggClient{
			Endpoint:   serverURL,
			HTTPClient: &http.Client{},
		}

		providerMeta := &utils.ProviderMeta{
			Frontegg: client,
		}

		if err := scim2GroupCreate(context.TODO(), d, providerMeta); err != nil {
			t.Fatal(err)
		}

		// Assertions to check the state after create
		r.Equal("Test Group", d.Get("name"))
		r.Equal("A test group description", d.Get("description"))
		r.NotEmpty(d.Id())
	})
}

func TestScimGroupResourceRead(t *testing.T) {
	r := require.New(t)

	testhelpers.WithMockFronteggServer(t, func(serverURL string) {
		client := &clients.FronteggClient{
			Endpoint:   serverURL,
			HTTPClient: &http.Client{},
		}

		providerMeta := &utils.ProviderMeta{
			Frontegg: client,
		}

		d := schema.TestResourceDataRaw(t, Scim2GroupSchema, nil)
		d.SetId("mock-group-id")

		if err := scim2GroupRead(context.TODO(), d, providerMeta); err != nil {
			t.Fatal(err)
		}

		// Assertions to check the state after read
		r.Equal("mock-group-id", d.Id())
		r.Equal("Test Group", d.Get("name"))
		r.Equal("A test group description", d.Get("description"))
	})
}

func TestScimGroupResourceDelete(t *testing.T) {
	r := require.New(t)

	testhelpers.WithMockFronteggServer(t, func(serverURL string) {
		client := &clients.FronteggClient{
			Endpoint:   serverURL,
			HTTPClient: &http.Client{},
		}

		providerMeta := &utils.ProviderMeta{
			Frontegg: client,
		}

		d := schema.TestResourceDataRaw(t, Scim2GroupSchema, nil)
		d.SetId("mock-group-id")

		if err := scim2GroupDelete(context.TODO(), d, providerMeta); err != nil {
			t.Fatal(err)
		}

		// Assertions to check the state after delete
		r.Empty(d.Id())
	})
}

// fronteggStatusServer answers every SCIM group call with status.
func fronteggStatusServer(status int) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(status)
	}))
}

func fronteggMeta(srv *httptest.Server) *utils.ProviderMeta {
	return &utils.ProviderMeta{Frontegg: &clients.FronteggClient{Endpoint: srv.URL, HTTPClient: srv.Client()}}
}

// A group removed in Frontegg should leave state on refresh, not fail the plan.
func TestScimGroupResourceReadGoneRemovesFromState(t *testing.T) {
	r := require.New(t)
	srv := fronteggStatusServer(http.StatusNotFound)
	defer srv.Close()

	d := schema.TestResourceDataRaw(t, Scim2GroupSchema, nil)
	d.SetId("gone-group")

	r.False(scim2GroupRead(context.TODO(), d, fronteggMeta(srv)).HasError())
	r.Empty(d.Id())
}

// Anything other than a 404 is a real failure and must not clear state.
func TestScimGroupResourceReadErrorKeepsState(t *testing.T) {
	r := require.New(t)
	srv := fronteggStatusServer(http.StatusInternalServerError)
	defer srv.Close()

	d := schema.TestResourceDataRaw(t, Scim2GroupSchema, nil)
	d.SetId("some-group")

	r.True(scim2GroupRead(context.TODO(), d, fronteggMeta(srv)).HasError())
	r.Equal("some-group", d.Id())
}

func TestScimGroupResourceDeleteGoneSucceeds(t *testing.T) {
	r := require.New(t)
	srv := fronteggStatusServer(http.StatusNotFound)
	defer srv.Close()

	d := schema.TestResourceDataRaw(t, Scim2GroupSchema, nil)
	d.SetId("gone-group")

	r.False(scim2GroupDelete(context.TODO(), d, fronteggMeta(srv)).HasError())
	r.Empty(d.Id())
}
