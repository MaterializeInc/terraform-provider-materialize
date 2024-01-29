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

func TestSCIM2ConfigurationResourceCreate(t *testing.T) {
	r := require.New(t)

	in := map[string]interface{}{
		"source":          "okta",
		"connection_name": "test-connection",
	}
	d := schema.TestResourceDataRaw(t, resourceSCIM2ConfigurationsSchema, in)
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

		if err := resourceSCIM2ConfigurationsCreate(context.TODO(), d, providerMeta); err != nil {
			t.Fatal(err)
		}

		// Assertions to check the state after create
		r.Equal("okta", d.Get("source"))
		r.Equal("SCIM", d.Get("connection_name"))
		r.Equal(true, d.Get("sync_to_user_management"))
		r.Equal("mock-token", d.Get("token"))
		r.NotEmpty(d.Get("created_at"))
	})
}

func TestSCIM2ConfigurationResourceRead(t *testing.T) {
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

		d := schema.TestResourceDataRaw(t, resourceSCIM2ConfigurationsSchema, nil)
		d.SetId("65a55dc187ee9cddee3aa8aa")

		if err := resourceSCIM2ConfigurationsRead(context.TODO(), d, providerMeta); err != nil {
			t.Fatal(err)
		}

		// Assertions to check the state after read
		r.Equal("65a55dc187ee9cddee3aa8aa", d.Id())
		r.Equal("okta", d.Get("source"))
		r.Equal("SCIM", d.Get("connection_name"))
		r.Equal(true, d.Get("sync_to_user_management"))
		r.NotEmpty(d.Get("created_at"))
	})
}

func TestSCIM2ConfigurationResourceDelete(t *testing.T) {
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

		d := schema.TestResourceDataRaw(t, resourceSCIM2ConfigurationsSchema, nil)
		d.SetId("mock-scim-config-id")

		if err := resourceSCIM2ConfigurationsDelete(context.TODO(), d, providerMeta); err != nil {
			t.Fatal(err)
		}

		// Assertions to check the state after delete
		r.Equal("", d.Id())
	})
}
