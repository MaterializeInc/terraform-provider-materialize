package clients

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

type MockFronteggService struct {
	MockResponseStatus int
	LastRequest        *http.Request
}

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

func (m *MockFronteggService) RoundTrip(req *http.Request) (*http.Response, error) {
	m.LastRequest = req
	// Check the requested URL and return a response accordingly
	if strings.HasSuffix(req.URL.Path, "/api/cloud-regions") {
		// Mock response data
		data := CloudProviderResponse{
			Data: []CloudProvider{
				{ID: "aws/us-east-1", Name: "us-east-1", Url: "http://mockendpoint", CloudProvider: "aws"},
				{ID: "aws/eu-west-1", Name: "eu-west-1", Url: "http://mockendpoint", CloudProvider: "aws"},
			},
		}

		// Convert response data to JSON
		respData, _ := json.Marshal(data)

		// Create a new HTTP response with the JSON data and the specified status code
		return &http.Response{
			StatusCode: m.MockResponseStatus,
			Body:       io.NopCloser(bytes.NewReader(respData)),
			Header:     make(http.Header),
		}, nil
	} else if strings.HasSuffix(req.URL.Path, "/api/region") {
		// Return mock response for GetRegionDetails
		details := CloudRegion{
			RegionInfo: &RegionInfo{
				SqlAddress:  "sql.materialize.com",
				HttpAddress: "http.materialize.com",
				Resolvable:  true,
				EnabledAt:   "2021-01-01T00:00:00Z",
			},
		}
		respData, _ := json.Marshal(details)
		return &http.Response{
			StatusCode: m.MockResponseStatus,
			Body:       io.NopCloser(bytes.NewReader(respData)),
			Header:     make(http.Header),
		}, nil
	}
	return nil, fmt.Errorf("no mock available for the requested endpoint")
}

func TestCloudAPIClient_ListCloudProviders(t *testing.T) {
	mockService := &MockFronteggService{
		MockResponseStatus: http.StatusOK,
	}
	mockClient := &http.Client{Transport: mockService}
	mockAuthenticator := &MockAuthenticator{Token: "mock-token"}

	apiClient := &CloudAPIClient{
		HTTPClient:    mockClient,
		Authenticator: mockAuthenticator,
		Endpoint:      "http://mockendpoint.com",
	}

	providers, err := apiClient.ListCloudProviders(context.Background())
	require.NoError(t, err)
	require.Len(t, providers, 2)

	require.True(t, mockAuthenticator.NeedsRefreshCalled)
	require.Equal(t, "Bearer mock-token", mockService.LastRequest.Header.Get("Authorization"))
}

func TestCloudAPIClient_GetRegionDetails(t *testing.T) {
	mockService := &MockFronteggService{
		MockResponseStatus: http.StatusOK,
	}
	mockClient := &http.Client{Transport: mockService}
	mockAuthenticator := &MockAuthenticator{Token: "mock-token"}

	apiClient := &CloudAPIClient{
		HTTPClient:    mockClient,
		Authenticator: mockAuthenticator,
		Endpoint:      "http://mockendpoint.com",
	}

	provider := CloudProvider{
		ID:   "aws/us-east-1",
		Name: "us-east-1",
		Url:  "http://mockendpoint.com/api/region",
	}

	region, err := apiClient.GetRegionDetails(context.Background(), provider)
	require.NoError(t, err)
	require.Equal(t, "sql.materialize.com", region.RegionInfo.SqlAddress)

	require.True(t, mockAuthenticator.NeedsRefreshCalled)
	require.Equal(t, "Bearer mock-token", mockService.LastRequest.Header.Get("Authorization"))
}

func TestCloudAPIClient_GetHost(t *testing.T) {
	mockService := &MockFronteggService{
		MockResponseStatus: http.StatusOK,
	}
	mockClient := &http.Client{Transport: mockService}
	mockAuthenticator := &MockAuthenticator{Token: "mock-token"}

	apiClient := &CloudAPIClient{
		HTTPClient:    mockClient,
		Authenticator: mockAuthenticator,
		Endpoint:      "http://mockendpoint.com",
	}

	regionID := "aws/us-east-1"

	sqlAddress, err := apiClient.GetHost(context.Background(), regionID)
	require.NoError(t, err)
	require.Equal(t, "sql.materialize.com", sqlAddress)

	require.True(t, mockAuthenticator.NeedsRefreshCalled)
}

func TestCloudAPIClient_ListCloudProviders_ErrorResponse(t *testing.T) {
	mockService := &MockFronteggService{
		MockResponseStatus: http.StatusInternalServerError,
	}
	mockClient := &http.Client{Transport: mockService}
	mockAuthenticator := &MockAuthenticator{Token: "mock-token"}

	apiClient := &CloudAPIClient{
		HTTPClient:    mockClient,
		Authenticator: mockAuthenticator,
		Endpoint:      "http://mockendpoint.com",
	}

	_, err := apiClient.ListCloudProviders(context.Background())
	require.Error(t, err)
}

func TestCloudAPIClient_GetRegionDetails_ErrorResponse(t *testing.T) {
	mockService := &MockFronteggService{
		MockResponseStatus: http.StatusInternalServerError,
	}
	mockClient := &http.Client{Transport: mockService}
	mockAuthenticator := &MockAuthenticator{Token: "mock-token"}

	apiClient := &CloudAPIClient{
		HTTPClient:    mockClient,
		Authenticator: mockAuthenticator,
		Endpoint:      "http://mockendpoint.com",
	}

	provider := CloudProvider{
		ID:   "aws/us-east-1",
		Name: "us-east-1",
		Url:  "http://mockendpoint.com/api/region",
	}

	_, err := apiClient.GetRegionDetails(context.Background(), provider)
	require.Error(t, err)
}

func TestCloudAPIClient_GetHost_RegionNotFound(t *testing.T) {
	mockService := &MockFronteggService{
		MockResponseStatus: http.StatusOK,
	}
	mockClient := &http.Client{Transport: mockService}
	mockAuthenticator := &MockAuthenticator{Token: "mock-token"}

	apiClient := &CloudAPIClient{
		HTTPClient:    mockClient,
		Authenticator: mockAuthenticator,
		Endpoint:      "http://mockendpoint.com",
	}
	regionID := "non-existent-region"

	// Call the method to test
	_, err := apiClient.GetHost(context.Background(), regionID)

	// Verify that an error is returned when the region is not found
	require.Error(t, err)
	require.Contains(t, err.Error(), "provider for region 'non-existent-region' not found")

	// Verify that the authentication method was called
	require.True(t, mockAuthenticator.NeedsRefreshCalled)
}

func TestNewCloudAPIClient(t *testing.T) {
	mockAuthenticator := &MockAuthenticator{Token: "mock-token"}

	customEndpoint := "http://custom-endpoint.com/api"
	baseEndpoint := "http://cloud.frontegg.com"
	cloudAPIClient := NewCloudAPIClient(mockAuthenticator, customEndpoint, baseEndpoint)

	require.NotNil(t, cloudAPIClient)
	require.Equal(t, mockAuthenticator, cloudAPIClient.Authenticator)
	require.NotNil(t, cloudAPIClient.HTTPClient)
	require.Equal(t, customEndpoint, cloudAPIClient.Endpoint)
	require.Equal(t, baseEndpoint, cloudAPIClient.BaseEndpoint)
}

func TestCloudAPIClient_EnableRegion_Success(t *testing.T) {
	mockService := &MockFronteggService{
		MockResponseStatus: http.StatusOK,
	}
	mockClient := &http.Client{Transport: mockService}
	mockAuthenticator := &MockAuthenticator{Token: "mock-token"}

	apiClient := &CloudAPIClient{
		HTTPClient:    mockClient,
		Authenticator: mockAuthenticator,
		Endpoint:      "http://mockendpoint.com",
	}

	provider := CloudProvider{
		ID:   "aws/us-east-1",
		Name: "us-east-1",
		Url:  "http://mockendpoint.com/api/region",
	}

	region, err := apiClient.EnableRegion(context.Background(), provider)
	require.NoError(t, err)
	require.NotNil(t, region)
	require.Equal(t, "sql.materialize.com", region.RegionInfo.SqlAddress)
	require.Equal(t, "http.materialize.com", region.RegionInfo.HttpAddress)
	require.True(t, region.RegionInfo.Resolvable)
	require.Equal(t, "2021-01-01T00:00:00Z", region.RegionInfo.EnabledAt)

	require.True(t, mockAuthenticator.NeedsRefreshCalled)
	require.Equal(t, "Bearer mock-token", mockService.LastRequest.Header.Get("Authorization"))
}

func TestCloudAPIClient_EnableRegion_Error(t *testing.T) {
	mockService := &MockFronteggService{
		MockResponseStatus: http.StatusInternalServerError,
	}
	mockClient := &http.Client{Transport: mockService}
	mockAuthenticator := &MockAuthenticator{Token: "mock-token"}

	apiClient := &CloudAPIClient{
		HTTPClient:    mockClient,
		Authenticator: mockAuthenticator,
		Endpoint:      "http://mockendpoint.com",
	}

	provider := CloudProvider{
		ID:   "aws/us-east-1",
		Name: "us-east-1",
		Url:  "http://mockendpoint.com/api/region",
	}

	_, err := apiClient.EnableRegion(context.Background(), provider)
	require.Error(t, err)
	require.Contains(t, err.Error(), "cloud API returned non-200/201 status code:")
}
