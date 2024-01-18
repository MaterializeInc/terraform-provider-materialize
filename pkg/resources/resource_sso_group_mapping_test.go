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

func TestSSORoleGroupMappingCreate(t *testing.T) {
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

		// Create a set of role strings
		roles := schema.NewSet(schema.HashString, []interface{}{"Admin", "Member"})

		// Set the expected values for sso_config_id, group, and roles
		d := schema.TestResourceDataRaw(t, SSORoleGroupMappingSchema, map[string]interface{}{
			"sso_config_id": "expected-sso-config-id",
			"group":         "expected-group",
			"roles":         roles.List(), // Convert set to a slice of interface{}
		})
		d.SetId("mock-group-id")

		diags := ssoGroupMappingCreate(context.TODO(), d, providerMeta)
		r.Nil(diags)

		// Add assertions to check the state after creation
		// You can add assertions based on the expected response or resource state
		r.Equal("expected-group", d.Get("group"))

		// Convert the roles back to a slice for assertion
		rolesSlice := d.Get("roles").([]interface{})
		r.Contains(rolesSlice, "Admin")
		r.Contains(rolesSlice, "Member")
	})
}

func TestSSORoleGroupMappingRead(t *testing.T) {
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

		// Set the initial state with "group" and "roles"
		d := schema.TestResourceDataRaw(t, SSORoleGroupMappingSchema, map[string]interface{}{
			"group": "initial-group",
			"roles": "Member",
		})
		d.SetId("mock-group-id")

		diags := ssoGroupMappingRead(context.TODO(), d, providerMeta)
		r.Nil(diags)

		// Add assertions to check the state after read
		r.Equal("initial-group", d.Get("group"))
		r.NotContains(d.Get("roles").([]interface{}), "Admin")
	})
}

func TestSSORoleGroupMappingUpdate(t *testing.T) {
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

		// Create a list of role strings for the updated roles
		updatedRoles := []string{"Member"}

		// Set the initial state with "group" and "roles"
		d := schema.TestResourceDataRaw(t, SSORoleGroupMappingSchema, map[string]interface{}{
			"group": "initial-group",
			"roles": []interface{}{"Member"},
		})
		d.SetId("mock-group-id")

		// Perform the update by setting the new "roles" attribute
		d.Set("roles", updatedRoles)

		diags := ssoGroupMappingUpdate(context.TODO(), d, providerMeta)
		r.Nil(diags)

		// Add assertions to check the state after the update
		r.Equal("initial-group", d.Get("group"))

		// Assert that "roles" attribute has been updated correctly
		updatedRolesSlice := d.Get("roles").([]interface{})
		r.Len(updatedRolesSlice, 1)
		r.Equal("Member", updatedRolesSlice[0].(string))

	})
}

func TestSSORoleGroupMappingDelete(t *testing.T) {
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

		// Set the initial state with "group" and "roles" as a list
		d := schema.TestResourceDataRaw(t, SSORoleGroupMappingSchema, map[string]interface{}{
			"group": "initial-group",
			"roles": []interface{}{"Member"},
		})
		d.SetId("mock-group-id")

		diags := ssoGroupMappingDelete(context.TODO(), d, providerMeta)
		r.Nil(diags)

		// Check if the resource ID is empty after deletion
		r.Equal("", d.Id())
	})
}
