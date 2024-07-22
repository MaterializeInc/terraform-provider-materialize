package resources

import (
	"context"
	"net/http"
	"testing"

	"github.com/MaterializeInc/terraform-provider-materialize/pkg/clients"
	"github.com/MaterializeInc/terraform-provider-materialize/pkg/testhelpers"
	"github.com/MaterializeInc/terraform-provider-materialize/pkg/utils"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/require"
)

type MockAuthenticator struct {
	Token              string
	RefreshCalled      bool
	NeedsRefreshCalled bool
}

func (m *MockAuthenticator) GetToken() string {
	return m.Token
}

func (m *MockAuthenticator) RefreshToken() error {
	m.RefreshCalled = true
	return nil
}

func (m *MockAuthenticator) NeedsTokenRefresh() error {
	m.NeedsRefreshCalled = true
	return nil
}

func TestResourceCloudRegionCreate(t *testing.T) {
	r := require.New(t)

	testhelpers.WithMockCloudServer(t, func(serverURL string) {
		mockClient := &http.Client{
			Transport: &testhelpers.MockCloudService{},
		}

		mockAuthenticator := &MockAuthenticator{Token: "mock-token"}

		mockCloudClient := &clients.CloudAPIClient{
			HTTPClient:    mockClient,
			Authenticator: mockAuthenticator,
			Endpoint:      serverURL,
		}

		providerMeta := &utils.ProviderMeta{
			CloudAPI:      mockCloudClient,
			Authenticator: mockAuthenticator,
		}

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
		r.True(mockAuthenticator.NeedsRefreshCalled)
	})
}

func TestResourceCloudRegionRead(t *testing.T) {
	r := require.New(t)

	testhelpers.WithMockCloudServer(t, func(serverURL string) {
		mockClient := &http.Client{
			Transport: &testhelpers.MockCloudService{},
		}

		mockAuthenticator := &MockAuthenticator{Token: "mock-token"}

		mockCloudClient := &clients.CloudAPIClient{
			HTTPClient:    mockClient,
			Authenticator: mockAuthenticator,
			Endpoint:      serverURL,
		}

		providerMeta := &utils.ProviderMeta{
			CloudAPI:      mockCloudClient,
			Authenticator: mockAuthenticator,
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
		r.True(mockAuthenticator.NeedsRefreshCalled)
	})
}

func TestResourceCloudRegionDelete(t *testing.T) {
	r := require.New(t)

	ctx := context.Background()
	mockAuthenticator := &MockAuthenticator{Token: "mock-token"}
	mockCloudClient := &clients.CloudAPIClient{
		HTTPClient:    &http.Client{},
		Authenticator: mockAuthenticator,
		Endpoint:      "http://mockendpoint.com",
	}
	providerMeta := &utils.ProviderMeta{
		CloudAPI:      mockCloudClient,
		Authenticator: mockAuthenticator,
	}

	d := schema.TestResourceDataRaw(t, regionSchema, map[string]interface{}{"region_id": "aws/us-east-1"})
	d.SetId("aws/us-east-1")

	diags := resourceCloudRegionDelete(ctx, d, providerMeta)

	r.False(diags.HasError())
	r.Empty(d.Id())
}
