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
			FronteggRoles: map[string]string{
				"Admin":  "1",
				"Member": "2",
			},
		}

		// Set the expected values for sso_config_id, group, and roles
		d := schema.TestResourceDataRaw(t, SSORoleGroupMappingSchema, map[string]interface{}{
			"sso_config_id": "expected-sso-config-id",
			"group":         "expected-group",
		})
		d.Set("roles", schema.NewSet(schema.HashString, []interface{}{"Member", "Admin"}))
		d.SetId("mock-group-id")

		diags := ssoGroupMappingCreate(context.TODO(), d, providerMeta)
		r.Nil(diags)

		// Aassertions to check the state after creation
		r.Equal("expected-group", d.Get("group"))

		// Assert on the roles set
		_, ok := d.Get("roles").(*schema.Set)
		r.True(ok)
		r.Len(d.Get("roles").(*schema.Set).List(), 2)
		r.Equal("Member", d.Get("roles").(*schema.Set).List()[1].(string))
		r.Equal("Admin", d.Get("roles").(*schema.Set).List()[0].(string))
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
			FronteggRoles: map[string]string{
				"Admin":  "1",
				"Member": "2",
			},
		}

		// Set the initial state with "group" and "roles"
		d := schema.TestResourceDataRaw(t, SSORoleGroupMappingSchema, map[string]interface{}{
			"group": "initial-group",
		})
		d.Set("roles", schema.NewSet(schema.HashString, []interface{}{"Member"}))
		d.SetId("mock-group-id")

		diags := ssoGroupMappingRead(context.TODO(), d, providerMeta)
		r.Nil(diags)

		// Assertions to check the state after read
		r.Equal("initial-group", d.Get("group"))
		// Assert on the roles set
		_, ok := d.Get("roles").(*schema.Set)
		r.True(ok)
		r.Len(d.Get("roles").(*schema.Set).List(), 1)
		r.Equal("Member", d.Get("roles").(*schema.Set).List()[0].(string))
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
			FronteggRoles: map[string]string{
				"Admin":  "1",
				"Member": "2",
			},
		}

		// Set the initial state with "group" and "roles"
		d := schema.TestResourceDataRaw(t, SSORoleGroupMappingSchema, map[string]interface{}{
			"group": "initial-group",
		})
		d.Set("roles", schema.NewSet(schema.HashString, []interface{}{"Admin"}))
		d.SetId("mock-group-id")

		// Assert that "roles" attribute has been updated correctly
		initialRolesSet, ok := d.Get("roles").(*schema.Set)
		r.True(ok, "Expected roles to be a *schema.Set")
		r.True(initialRolesSet.Contains("Admin"), "Expected initial roles to contain 'Admin'")
		r.Equal(initialRolesSet.Len(), 1, "Expected initial roles set to have 1 item")

		// Perform the update by setting the new "roles" attribute
		updatedRoles := schema.NewSet(schema.HashString, []interface{}{"Member"})
		d.Set("roles", updatedRoles)

		diags := ssoGroupMappingUpdate(context.TODO(), d, providerMeta)
		r.Nil(diags)

		// Assertions to check the state after the update
		r.Equal("initial-group", d.Get("group"))

		// Assert that "roles" attribute has been updated correctly
		updatedRolesSet, ok := d.Get("roles").(*schema.Set)
		r.True(ok, "Expected roles to be a *schema.Set")
		r.True(updatedRolesSet.Contains("Member"), "Expected updated roles to contain 'Member'")
		r.Equal(updatedRolesSet.Len(), 1, "Expected updated roles set to have 1 item")
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
