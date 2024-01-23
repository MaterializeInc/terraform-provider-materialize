package datasources

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

func TestDataSourceSSOConfigRead_Success(t *testing.T) {
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

		d := schema.TestResourceDataRaw(t, dataSourceSSOConfigSchema, nil)
		d.SetId("mock-config-1")

		if err := dataSourceSSOConfigRead(context.TODO(), d, providerMeta); err != nil {
			t.Fatal(err)
		}

		// Validate the ID of the configuration
		r.Equal("mock-config-1", d.Id())

		// Validate the SSO configurations
		ssoConfigs := d.Get("sso_configs").([]interface{})
		r.NotEmpty(ssoConfigs)

		for _, ssoConfig := range ssoConfigs {
			configMap := ssoConfig.(map[string]interface{})

			// Validate each field within a SSO configuration
			r.Equal("mock-config-1", configMap["id"].(string))
			r.Equal(true, configMap["enabled"].(bool))
			r.Equal("https://sso.example.com", configMap["sso_endpoint"].(string))
			r.Equal("bW9jay1wdWJsaWMtY2VydGlmaWNhdGUK", configMap["public_certificate"].(string))
			r.Equal(true, configMap["sign_request"].(bool))
			r.Equal("SAML", configMap["type"].(string))
			r.Equal("mock-oidc-client-id", configMap["oidc_client_id"].(string))
			r.Equal("mock-oidc-secret", configMap["oidc_secret"].(string))

			// Validate domains
			domains := configMap["domains"].([]interface{})
			for _, domain := range domains {
				domainMap := domain.(map[string]interface{})
				r.Equal("domain-1", domainMap["id"].(string))
				r.Equal("example.com", domainMap["domain"].(string))
			}

			// Validate groups
			groups := configMap["groups"].([]interface{})
			for _, group := range groups {
				groupMap := group.(map[string]interface{})
				r.Equal("group-1", groupMap["id"].(string))
				r.Equal("admins", groupMap["group"].(string))
				r.Equal(false, groupMap["enabled"].(bool))

				// Validate role IDs in groups
				roleIDs := groupMap["role_ids"].([]interface{})
				for _, roleID := range roleIDs {
					r.Equal("role-1", roleID.(string))
				}
			}
		}
	})
}
