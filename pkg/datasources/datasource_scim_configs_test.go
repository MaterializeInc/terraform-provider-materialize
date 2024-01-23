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

func TestDataSourceSCIM2ConfigurationsRead_Success(t *testing.T) {
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

		d := schema.TestResourceDataRaw(t, dataSourceSCIM2ConfigurationsSchema, nil)
		d.SetId("scim2_configs")

		if err := dataSourceSCIM2ConfigurationsRead(context.TODO(), d, providerMeta); err != nil {
			t.Fatal(err)
		}

		// Validate the ID of the data source
		r.Equal("scim2_configs", d.Id())

		// Validate the SCIM 2.0 configurations
		scim2Configs := d.Get("configurations").([]interface{})
		r.NotEmpty(scim2Configs)
		r.Len(scim2Configs, 2)

		// Validate the first configuration
		configMap := scim2Configs[0].(map[string]interface{})
		r.Equal("65a55dc187ee9cddee3aa8aa", configMap["id"].(string))
		r.Equal("okta", configMap["source"].(string))
		r.Equal("15b545d4-9d14-4725-8476-295073a3fb04", configMap["tenant_id"].(string))
		r.Equal("SCIM", configMap["connection_name"].(string))
		r.Equal(true, configMap["sync_to_user_management"].(bool))
		r.Equal("2024-01-15T16:30:57.000Z", configMap["created_at"].(string))

		// Validate the second configuration (and so on for other configurations)
		configMap2 := scim2Configs[1].(map[string]interface{})
		r.Equal("65afa26a0d806f407e78dfa0", configMap2["id"].(string))
		r.Equal("okta", configMap2["source"].(string))
		r.Equal("15b545d4-9d14-4725-8476-295073a3fb04", configMap2["tenant_id"].(string))
		r.Equal("test2", configMap2["connection_name"].(string))
		r.Equal(true, configMap2["sync_to_user_management"].(bool))
		r.Equal("2024-01-23T11:26:34.000Z", configMap2["created_at"].(string))
	})
}
