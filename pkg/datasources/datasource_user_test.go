package datasources

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/MaterializeInc/terraform-provider-materialize/pkg/clients"
	"github.com/MaterializeInc/terraform-provider-materialize/pkg/testhelpers"
	"github.com/MaterializeInc/terraform-provider-materialize/pkg/utils"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/require"
)

func TestUserDataSourceRead(t *testing.T) {
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
		d.Set("email", "test@example.com")

		diags := userDataSourceRead(context.TODO(), d, providerMeta)
		r.Empty(diags)

		// Validate the user data
		r.Equal("new-mock-user-id", d.Id())
		r.Equal("test@example.com", d.Get("email"))
		r.Equal(false, d.Get("verified"))
		r.Equal("{}", d.Get("metadata"))
	})
}

func TestUserDataSourceRead_UserNotFound(t *testing.T) {
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
		d.Set("email", "nonexistent@example.com")

		diags := userDataSourceRead(context.TODO(), d, providerMeta)
		r.NotEmpty(diags)
		r.Equal(diag.Error, diags[0].Severity)
		r.Contains(diags[0].Summary, "no user found with email: nonexistent@example.com")
	})
}
