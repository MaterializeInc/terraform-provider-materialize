package resources

import (
	"context"
	"net/http"
	"sort"
	"testing"
	"time"

	"github.com/MaterializeInc/terraform-provider-materialize/pkg/clients"
	"github.com/MaterializeInc/terraform-provider-materialize/pkg/testhelpers"
	"github.com/MaterializeInc/terraform-provider-materialize/pkg/utils"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/require"
)

func TestSSODefaultRolesCreateOrUpdate(t *testing.T) {
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

		// Create a new ResourceData object
		d := schema.TestResourceDataRaw(t, SSODefaultRolesSchema, map[string]interface{}{
			"sso_config_id": "config-id",
		})

		// Set the roles using the Set method
		err := d.Set("roles", schema.NewSet(schema.HashString, []interface{}{"Member", "Admin"}))
		r.NoError(err)
		d.SetId("mock-config-id")

		diags := ssoDefaultRolesCreateOrUpdate(context.TODO(), d, providerMeta)
		r.Nil(diags)

		// Add assertions to check the state after creation or update
		r.Equal("config-id", d.Get("sso_config_id"))

		// Assert that "roles" attribute has been updated correctly
		updatedRolesSet := d.Get("roles").(*schema.Set)
		r.Len(updatedRolesSet.List(), 2)

		// Convert set to a slice for easier assertions
		updatedRolesSlice := convertToStringSlice(updatedRolesSet.List())
		// Sort the slice as the order is not guaranteed in sets
		sort.Strings(updatedRolesSlice)

		r.Equal("Admin", updatedRolesSlice[0])
		r.Equal("Member", updatedRolesSlice[1])
	})
}

func TestSSODefaultRolesRead(t *testing.T) {
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

		// Set the initial state with "sso_config_id" and "roles" as a list of strings
		d := schema.TestResourceDataRaw(t, SSODefaultRolesSchema, map[string]interface{}{
			"sso_config_id": "mock-config-id",
		})
		d.Set("roles", schema.NewSet(schema.HashString, []interface{}{"Member", "Admin"}))
		d.SetId("mock-config-id")

		diags := ssoDefaultRolesRead(context.TODO(), d, providerMeta)
		r.Nil(diags)

		// Add assertions to check the state after reading
		r.Equal("mock-config-id", d.Id())

		// Assert that "roles" attribute has been updated correctly
		updatedRolesSet := d.Get("roles").(*schema.Set)
		r.Len(updatedRolesSet.List(), 2)

		// Convert set to a slice for easier assertions
		updatedRolesSlice := convertToStringSlice(updatedRolesSet.List())
		// Sort the slice as the order is not guaranteed in sets
		sort.Strings(updatedRolesSlice)

		r.Equal("Admin", updatedRolesSlice[0])
		r.Equal("Member", updatedRolesSlice[1])
	})
}

func TestSSODefaultRolesDelete(t *testing.T) {
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

		// Set the initial state with "sso_config_id" and "roles" as a list of strings
		d := schema.TestResourceDataRaw(t, SSODefaultRolesSchema, map[string]interface{}{
			"sso_config_id": "config-id",
		})
		d.Set("roles", schema.NewSet(schema.HashString, []interface{}{"Member", "Admin"}))
		d.SetId("mock-config-id")

		diags := ssoDefaultRolesDelete(context.TODO(), d, providerMeta)
		r.Nil(diags)

		// Add assertions to check the state after deletion
		r.Equal("", d.Id())
	})
}
