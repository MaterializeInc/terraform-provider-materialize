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

func TestSSODomainResourceCreate(t *testing.T) {
	r := require.New(t)

	in := map[string]interface{}{
		"sso_config_id": "mock-sso-config-id",
		"domain":        "example.com",
	}
	d := schema.TestResourceDataRaw(t, SSODomainSchema, in)
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

		diags := ssoDomainCreate(context.TODO(), d, providerMeta)
		r.Nil(diags)

		// Assertions to check the state after create
		r.Equal(false, d.Get("validated"))
		r.Equal("mock-sso-config-id", d.Get("sso_config_id"))
	})
}

func TestSSODomainResourceRead(t *testing.T) {
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

		// Set the expected values for sso_config_id and domain
		d := schema.TestResourceDataRaw(t, SSODomainSchema, map[string]interface{}{
			"sso_config_id": "expected-sso-config-id",
			"domain":        "expected-domain",
			"validated":     true,
		})
		d.SetId("mock-domain-id")

		diags := ssoDomainRead(context.TODO(), d, providerMeta)
		r.Nil(diags)

		// Assertions to check the state after read
		r.Equal("expected-sso-config-id", d.Get("sso_config_id"))
		r.Equal("expected-domain", d.Get("domain"))
		r.Equal(true, d.Get("validated"))
	})
}

func TestSSODomainResourceUpdate(t *testing.T) {
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

		// Set the initial state with "validated" as false
		d := schema.TestResourceDataRaw(t, SSODomainSchema, map[string]interface{}{
			"validated": false,
			"domain":    "example.com",
		})
		d.SetId("mock-domain-id")

		diags := ssoDomainUpdate(context.TODO(), d, providerMeta)
		r.Nil(diags)

		// Assertions to check the state after update
		r.Equal(false, d.Get("validated"))
		r.Equal("example.com", d.Get("domain"))
	})
}

func TestSSODomainResourceDelete(t *testing.T) {
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

		d := schema.TestResourceDataRaw(t, SSODomainSchema, nil)
		d.SetId("mock-domain-id")

		diags := ssoDomainDelete(context.TODO(), d, providerMeta)
		r.Nil(diags)

		// Assertions to check the state after delete
		r.Empty(d.Id())
	})
}
