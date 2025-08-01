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

func TestAppPasswordResourceCreate(t *testing.T) {
	r := require.New(t)

	in := map[string]interface{}{
		"name": "test-app-password",
	}
	d := schema.TestResourceDataRaw(t, AppPassword().Schema, in)
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

		if err := appPasswordCreate(context.TODO(), d, providerMeta); err != nil {
			t.Fatal(err)
		}

		r.Equal("test-app-password", d.Get("name"))
		r.Equal("mock-secret", d.Get("secret"))
	})
}

func TestAppPasswordResourceCreate_SelfHostedError(t *testing.T) {
	r := require.New(t)

	in := map[string]interface{}{
		"name":  "test-app-password",
		"type":  "service",
		"user":  "test_user",
		"roles": []interface{}{"Member"},
	}
	d := schema.TestResourceDataRaw(t, AppPassword().Schema, in)
	r.NotNil(d)

	// Create a provider meta configured for self-hosted mode
	providerMeta := &utils.ProviderMeta{
		Mode: utils.ModeSelfHosted,
		// No Frontegg client in self-hosted mode
		Frontegg: nil,
	}

	diags := appPasswordCreate(context.TODO(), d, providerMeta)
	r.True(diags.HasError())
	r.Contains(diags[0].Summary, "materialize_app_password is only available in Materialize Cloud (SaaS) environments")
	r.Contains(diags[0].Summary, "You are currently using self-hosted authentication mode")
}

func TestAppPasswordResourceRead(t *testing.T) {
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

		d := schema.TestResourceDataRaw(t, AppPassword().Schema, nil)
		d.SetId("mock-client-id")

		if err := appPasswordRead(context.TODO(), d, providerMeta); err != nil {
			t.Fatal(err)
		}

		// Assertions to check the state after read
		r.Equal("mock-client-id", d.Id())
		r.Equal("test-app-password", d.Get("name"))
	})
}

func TestAppPasswordResourceDelete(t *testing.T) {
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

		d := schema.TestResourceDataRaw(t, AppPassword().Schema, nil)
		d.SetId("mock-client-id")

		if err := appPasswordDelete(context.TODO(), d, providerMeta); err != nil {
			t.Fatal(err)
		}

		r.Empty(d.Id())
	})
}
