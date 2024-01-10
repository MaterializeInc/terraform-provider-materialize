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
