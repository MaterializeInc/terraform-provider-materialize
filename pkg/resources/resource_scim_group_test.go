package resources

import (
	"context"
	"net/http"
	"testing"
	"time"

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
			Endpoint:    serverURL,
			HTTPClient:  &http.Client{},
			TokenExpiry: time.Date(9999, 1, 1, 0, 0, 0, 0, time.UTC),
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
			Endpoint:    serverURL,
			HTTPClient:  &http.Client{},
			TokenExpiry: time.Date(9999, 1, 1, 0, 0, 0, 0, time.UTC),
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
			Endpoint:    serverURL,
			HTTPClient:  &http.Client{},
			TokenExpiry: time.Date(9999, 1, 1, 0, 0, 0, 0, time.UTC),
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
