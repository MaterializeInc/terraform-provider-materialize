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

func TestUserResourceRead(t *testing.T) {
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

		d := schema.TestResourceDataRaw(t, User().Schema, nil)
		d.SetId("mock-user-id")

		if err := userRead(context.TODO(), d, providerMeta); err != nil {
			t.Fatal(err)
		}

		r.Equal("test@example.com", d.Get("email"))
	})
}

func TestUserResourceDelete(t *testing.T) {
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

		d := schema.TestResourceDataRaw(t, User().Schema, nil)
		d.SetId("mock-user-id")

		if err := userDelete(context.TODO(), d, providerMeta); err != nil {
			t.Fatal(err)
		}

		// Assert the user is removed from the state
		r.Empty(d.Id())
	})
}

func TestUserResourceUpdate(t *testing.T) {
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

		d := schema.TestResourceDataRaw(t, User().Schema, map[string]interface{}{
			"email": "test@example.com",
			"roles": []interface{}{"Member"},
		})
		d.SetId("mock-user-id")

		d.Set("roles", []interface{}{"Admin", "Member"})

		if err := userUpdate(context.TODO(), d, providerMeta); err != nil {
			t.Fatal(err)
		}

		if err := userRead(context.TODO(), d, providerMeta); err != nil {
			t.Fatal(err)
		}

		roles := d.Get("roles").([]interface{})
		r.Equal(2, len(roles), "Expected 2 roles after update")
		r.Contains(roles, "Admin", "Expected 'Admin' role after update")
		r.Contains(roles, "Member", "Expected 'Member' role after update")
	})
}
