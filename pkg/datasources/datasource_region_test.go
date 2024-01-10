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

func TestRegionRead(t *testing.T) {
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
		d := schema.TestResourceDataRaw(t, Region().Schema, nil)
		d.SetId("regions")

		// Call the RegionRead function
		diags := RegionRead(context.TODO(), d, providerMeta)

		// Print error messages in diagnostics
		for _, diag := range diags {
			t.Logf("Error: %s", diag.Summary)
			t.Logf("Details: %s", diag.Detail)
		}

		// Check for errors within the diagnostics
		r.False(diags.HasError())
		r.Equal("aws/us-east-1", d.Get("regions.0.id"))
		r.Equal("us-east-1", d.Get("regions.0.name"))
		r.Equal("http://mockendpoint", d.Get("regions.0.url"))
		r.Equal("aws", d.Get("regions.0.cloud_provider"))
		r.Equal("sql.materialize.com", d.Get("regions.0.host"))
	})
}
