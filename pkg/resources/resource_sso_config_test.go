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

func TestSSOConfigResourceCreate(t *testing.T) {
	r := require.New(t)

	in := map[string]interface{}{
		"enabled":            true,
		"sso_endpoint":       "https://example.com/sso",
		"public_certificate": "public_cert",
		"sign_request":       true,
		"type":               "SAML",
		"oidc_client_id":     "client_id",
		"oidc_secret":        "secret",
	}
	d := schema.TestResourceDataRaw(t, SSOConfigSchema, in)
	r.NotNil(d)

	testhelpers.WithMockFronteggServer(t, func(serverURL string) {
		client := &clients.FronteggClient{
			Endpoint:    serverURL,
			HTTPClient:  &http.Client{},
			TokenExpiry: time.Date(9999, 1, 1, 0, 0, 0, 0, time.UTC),
		}

		mockCloudClient := &clients.CloudAPIClient{
			FronteggClient: client,
			Endpoint:       serverURL,
		}

		providerMeta := &utils.ProviderMeta{
			Frontegg: client,
			CloudAPI: mockCloudClient,
		}

		if err := ssoConfigCreate(context.TODO(), d, providerMeta); err != nil {
			t.Fatal(err)
		}

		// Assertions to check the state after create
		r.True(d.Get("enabled").(bool))
		r.Equal("https://example.com/sso", d.Get("sso_endpoint"))
		r.Equal("public_cert", d.Get("public_certificate"))
		r.True(d.Get("sign_request").(bool))
		r.Equal("SAML", d.Get("type"))
	})
}

func TestSSOConfigResourceRead(t *testing.T) {
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

		d := schema.TestResourceDataRaw(t, SSOConfigSchema, nil)
		d.SetId("mock-config-1")

		if err := ssoConfigRead(context.TODO(), d, providerMeta); err != nil {
			t.Fatal(err)
		}

		r.Equal("mock-config-1", d.Id())
		r.Equal(true, d.Get("enabled"))
		r.Equal("https://sso.example.com", d.Get("sso_endpoint"))
		r.Equal("mock-public-certificate\n", d.Get("public_certificate"))
	})

}

func TestSSOConfigResourceUpdate(t *testing.T) {
	r := require.New(t)

	testhelpers.WithMockFronteggServer(t, func(serverURL string) {
		client := &clients.FronteggClient{
			Endpoint:    serverURL,
			HTTPClient:  &http.Client{},
			TokenExpiry: time.Date(9999, 1, 1, 0, 0, 0, 0, time.UTC),
		}

		mockCloudClient := &clients.CloudAPIClient{
			FronteggClient: client,
			Endpoint:       serverURL,
		}

		providerMeta := &utils.ProviderMeta{
			Frontegg: client,
			CloudAPI: mockCloudClient,
		}

		d := schema.TestResourceDataRaw(t, SSOConfigSchema, nil)
		d.SetId("mock-config-1")

		if err := ssoConfigUpdate(context.TODO(), d, providerMeta); err != nil {
			t.Fatal(err)
		}
		r.Equal("mock-config-1", d.Id())
	})
}

func TestSSOConfigResourceDelete(t *testing.T) {
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

		d := schema.TestResourceDataRaw(t, SSOConfigSchema, nil)
		d.SetId("mock-sso-config-id")

		if err := ssoConfigDelete(context.TODO(), d, providerMeta); err != nil {
			t.Fatal(err)
		}

		r.Empty(d.Id())
	})
}
