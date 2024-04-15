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

func TestResourceCloudRegionCreate(t *testing.T) {
	r := require.New(t)

	// Set up the mock cloud server
	testhelpers.WithMockCloudServer(t, func(serverURL string) {
		// Create an http.Client that uses the mock transport
		mockClient := &http.Client{
			Transport: &testhelpers.MockCloudService{},
		}

		fronteggClient := &clients.FronteggClient{
			Endpoint:    serverURL,
			HTTPClient:  mockClient,
			TokenExpiry: time.Date(9999, 1, 1, 0, 0, 0, 0, time.UTC),
		}
		// Create a mock cloud client
		mockCloudClient := &clients.CloudAPIClient{
			FronteggClient: fronteggClient,
			Endpoint:       serverURL,
		}

		// Create a provider meta with the mock cloud client
		providerMeta := &utils.ProviderMeta{
			CloudAPI: mockCloudClient,
			Frontegg: fronteggClient,
		}

		// Create a test resource data with the Region schema
		d := schema.TestResourceDataRaw(t, regionSchema, map[string]interface{}{"region_id": "aws/us-east-1"})

		diags := resourceCloudRegionCreate(context.Background(), d, providerMeta)

		for _, diag := range diags {
			t.Logf("Error: %s", diag.Summary)
			t.Logf("Details: %s", diag.Detail)
		}

		r.False(diags.HasError())
		r.Equal("aws/us-east-1", d.Get("region_id"))
		r.Equal("http.materialize.com", d.Get("http_address"))
		r.Equal("sql.materialize.com", d.Get("sql_address"))
		r.True(d.Get("resolvable").(bool))
		r.True(d.Get("region_state").(bool))
	})
}

func TestResourceCloudRegionRead(t *testing.T) {
	r := require.New(t)

	testhelpers.WithMockCloudServer(t, func(serverURL string) {
		mockClient := &http.Client{
			Transport: &testhelpers.MockCloudService{},
		}

		fronteggClient := &clients.FronteggClient{
			Endpoint:    serverURL,
			HTTPClient:  mockClient,
			TokenExpiry: time.Date(9999, 1, 1, 0, 0, 0, 0, time.UTC),
		}

		mockCloudClient := &clients.CloudAPIClient{
			FronteggClient: fronteggClient,
			Endpoint:       serverURL,
		}

		providerMeta := &utils.ProviderMeta{
			CloudAPI: mockCloudClient,
			Frontegg: fronteggClient,
		}

		d := schema.TestResourceDataRaw(t, regionSchema, map[string]interface{}{"region_id": "aws/us-east-1"})
		d.SetId("aws/us-east-1")

		diags := resourceCloudRegionRead(context.Background(), d, providerMeta)

		for _, diag := range diags {
			t.Logf("Error: %s", diag.Summary)
			t.Logf("Details: %s", diag.Detail)
		}

		r.False(diags.HasError())
		r.Equal("aws/us-east-1", d.Get("region_id"))
		r.Equal("http.materialize.com", d.Get("http_address"))
		r.Equal("sql.materialize.com", d.Get("sql_address"))
		r.True(d.Get("resolvable").(bool))
		r.True(d.Get("region_state").(bool))
	})
}

func TestResourceCloudRegionDelete(t *testing.T) {
	r := require.New(t)

	ctx := context.Background()
	providerMeta := &utils.ProviderMeta{}

	d := schema.TestResourceDataRaw(t, regionSchema, map[string]interface{}{"region_id": "aws/us-east-1"})
	d.SetId("aws/us-east-1")

	diags := resourceCloudRegionDelete(ctx, d, providerMeta)

	r.False(diags.HasError())
	r.Empty(d.Id())
}
