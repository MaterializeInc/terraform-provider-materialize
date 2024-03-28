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

func TestScimGroupUsersCreate(t *testing.T) {
	r := require.New(t)

	in := map[string]interface{}{
		"group_id": "test-group-id",
		"users":    []interface{}{"user1", "user2"},
	}
	d := schema.TestResourceDataRaw(t, ScimGroupUsersSchema, in)
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

		if err := scimGroupUsersCreate(context.TODO(), d, providerMeta); err != nil {
			t.Fatal(err)
		}

		// Assertions to check the state after create
		r.Equal("test-group-id", d.Id())
	})
}

func TestScimGroupUsersRead(t *testing.T) {
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

		d := schema.TestResourceDataRaw(t, ScimGroupUsersSchema, nil)
		d.SetId("mock-group-id")

		if err := scimGroupUsersRead(context.TODO(), d, providerMeta); err != nil {
			t.Fatal(err)
		}

		r.Equal("mock-group-id", d.Id())
	})
}

func TestScimGroupUsersUpdate(t *testing.T) {

	testhelpers.WithMockFronteggServer(t, func(serverURL string) {
		client := &clients.FronteggClient{
			Endpoint:    serverURL,
			HTTPClient:  &http.Client{},
			TokenExpiry: time.Date(9999, 1, 1, 0, 0, 0, 0, time.UTC),
		}

		providerMeta := &utils.ProviderMeta{
			Frontegg: client,
		}

		d := schema.TestResourceDataRaw(t, ScimGroupUsersSchema, nil)
		d.SetId("mock-group-id")
		d.Set("users", []interface{}{"user1", "user2"})

		if err := scimGroupUsersUpdate(context.TODO(), d, providerMeta); err != nil {
			t.Fatal(err)
		}

		// Assertions to check the state after update
	})
}

func TestScimGroupUsersDelete(t *testing.T) {
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

		d := schema.TestResourceDataRaw(t, ScimGroupUsersSchema, nil)
		d.SetId("mock-group-id")
		d.Set("users", []interface{}{"user1", "user2"})

		if err := scimGroupUsersDelete(context.TODO(), d, providerMeta); err != nil {
			t.Fatal(err)
		}

		// Assertions to check the state after delete
		r.Empty(d.Id())
	})
}
