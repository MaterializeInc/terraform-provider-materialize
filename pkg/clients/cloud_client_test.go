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
}

func (m *MockFronteggService) RoundTrip(req *http.Request) (*http.Response, error) {
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
	apiClient := &CloudAPIClient{
		FronteggClient: &FronteggClient{HTTPClient: mockClient},
		Endpoint:       "http://mockendpoint.com",
	}

	// Call the method to test
	providers, err := apiClient.ListCloudProviders(context.Background())
	if err != nil {
		t.Fatalf("ListCloudProviders() error: %v", err)
	}

	// Verify the results
	wantProviderCount := 2
	if len(providers) != wantProviderCount {
		t.Errorf("ListCloudProviders() got %v providers, want %v", len(providers), wantProviderCount)
	}
}

func TestCloudAPIClient_GetRegionDetails(t *testing.T) {
	mockService := &MockFronteggService{
		MockResponseStatus: http.StatusOK,
	}
	mockClient := &http.Client{Transport: mockService}
	apiClient := &CloudAPIClient{
		FronteggClient: &FronteggClient{HTTPClient: mockClient},
		Endpoint:       "http://mockendpoint.com",
	}

	provider := CloudProvider{
		ID:   "aws/us-east-1",
		Name: "us-east-1",
		Url:  "http://mockendpoint.com/api/region",
	}

	// Call the method to test
	region, err := apiClient.GetRegionDetails(context.Background(), provider)
	if err != nil {
		t.Fatalf("GetRegionDetails() error: %v", err)
	}

	// Verify the results
	wantSqlAddress := "sql.materialize.com"
	if region.RegionInfo.SqlAddress != wantSqlAddress {
		t.Errorf("GetRegionDetails() got SqlAddress = %v, want %v", region.RegionInfo.SqlAddress, wantSqlAddress)
	}
}

func TestCloudAPIClient_GetHost(t *testing.T) {
	mockService := &MockFronteggService{
		MockResponseStatus: http.StatusOK,
	}
	mockClient := &http.Client{Transport: mockService}
	apiClient := &CloudAPIClient{
		FronteggClient: &FronteggClient{HTTPClient: mockClient},
		Endpoint:       "http://mockendpoint.com",
	}

	regionID := "aws/us-east-1"

	sqlAddress, err := apiClient.GetHost(context.Background(), regionID)
	if err != nil {
		t.Fatalf("GetHost() error: %v", err)
	}

	// Verify the results
	wantSqlAddress := "sql.materialize.com"
	if sqlAddress != wantSqlAddress {
		t.Errorf("GetHost() got SqlAddress = %v, want %v", sqlAddress, wantSqlAddress)
	}
}

func TestCloudAPIClient_ListCloudProviders_ErrorResponse(t *testing.T) {
	mockService := &MockFronteggService{
		// Mock the HTTP response to return an error status code:
		MockResponseStatus: http.StatusInternalServerError,
	}
	mockClient := &http.Client{Transport: mockService}
	apiClient := &CloudAPIClient{
		FronteggClient: &FronteggClient{HTTPClient: mockClient},
		Endpoint:       "http://mockendpoint.com",
	}

	// Call the method to test
	_, err := apiClient.ListCloudProviders(context.Background())

	// Verify that an error is returned when the server responds with an error status code
	require.Error(t, err)
}

func TestCloudAPIClient_GetRegionDetails_ErrorResponse(t *testing.T) {
	mockService := &MockFronteggService{
		// Mock the HTTP response to return an error status code
		MockResponseStatus: http.StatusInternalServerError,
	}
	mockClient := &http.Client{Transport: mockService}
	apiClient := &CloudAPIClient{
		FronteggClient: &FronteggClient{HTTPClient: mockClient},
		Endpoint:       "http://mockendpoint.com",
	}
	provider := CloudProvider{
		ID:   "aws/us-east-1",
		Name: "us-east-1",
		Url:  "http://mockendpoint.com/api/region",
	}

	// Call the method to test
	_, err := apiClient.GetRegionDetails(context.Background(), provider)

	// Verify that an error is returned when the server responds with an error status code
	require.Error(t, err)
}

func TestCloudAPIClient_GetHost_RegionNotFound(t *testing.T) {
	mockService := &MockFronteggService{
		MockResponseStatus: http.StatusOK,
	}
	mockClient := &http.Client{Transport: mockService}
	apiClient := &CloudAPIClient{
		FronteggClient: &FronteggClient{HTTPClient: mockClient},
		Endpoint:       "http://mockendpoint.com",
	}
	regionID := "non-existent-region"

	// Call the method to test
	_, err := apiClient.GetHost(context.Background(), regionID)

	// Verify that an error is returned when the region is not found
	require.Error(t, err)
	require.Contains(t, err.Error(), "provider for region 'non-existent-region' not found")
}

func TestNewCloudAPIClient(t *testing.T) {
	// Create a FronteggClient instance for testing
	fronteggClient := &FronteggClient{}

	// Call the NewCloudAPIClient function with a custom API endpoint
	customEndpoint := "http://custom-endpoint.com/api"
	baseEndpoint := "http://cloud.frontegg.com"
	cloudAPIClient := NewCloudAPIClient(fronteggClient, customEndpoint, baseEndpoint)

	// Assert that the returned CloudAPIClient has the expected properties
	require.NotNil(t, cloudAPIClient)
	require.Equal(t, fronteggClient, cloudAPIClient.FronteggClient)
	require.NotNil(t, cloudAPIClient.HTTPClient)
	require.Equal(t, customEndpoint, cloudAPIClient.Endpoint)

	// Call the NewCloudAPIClient function with a different custom API endpoint
	anotherCustomEndpoint := "http://another-custom-endpoint.com/api"
	cloudAPIClient = NewCloudAPIClient(fronteggClient, anotherCustomEndpoint, baseEndpoint)

	// Assert that the returned CloudAPIClient has the updated custom endpoint
	require.NotNil(t, cloudAPIClient)
	require.Equal(t, fronteggClient, cloudAPIClient.FronteggClient)
	require.NotNil(t, cloudAPIClient.HTTPClient)
	require.Equal(t, anotherCustomEndpoint, cloudAPIClient.Endpoint)
}
