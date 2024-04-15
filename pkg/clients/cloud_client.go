package clients

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
)

// RegionInfo holds the detailed information about a region from the Cloud API
type RegionInfo struct {
	SqlAddress  string `json:"sqlAddress"`
	HttpAddress string `json:"httpAddress"`
	Resolvable  bool   `json:"resolvable"`
	EnabledAt   string `json:"enabledAt"`
}

// Region holds the connection details for an active region
type CloudRegion struct {
	RegionInfo *RegionInfo `json:"regionInfo"`
}

// CloudProvider contains the information about a cloud provider and its region
type CloudProvider struct {
	ID            string `json:"id"`
	Name          string `json:"name"`
	Url           string `json:"url"`
	CloudProvider string `json:"cloudProvider"`
}

// CloudProviderResponse represents the response for listing cloud providers
type CloudProviderResponse struct {
	Data       []CloudProvider `json:"data"`
	NextCursor string          `json:"nextCursor,omitempty"`
}

// CloudAPIClient is a client for interacting with the Materialize Cloud API
type CloudAPIClient struct {
	HTTPClient     *http.Client
	FronteggClient *FronteggClient
	Endpoint       string
	BaseEndpoint   string
}

// NewCloudAPIClient creates a new Cloud API client
func NewCloudAPIClient(fronteggClient *FronteggClient, cloudAPIEndpoint, baseEndpoint string) *CloudAPIClient {
	return &CloudAPIClient{
		HTTPClient:     &http.Client{},
		FronteggClient: fronteggClient,
		Endpoint:       cloudAPIEndpoint,
		BaseEndpoint:   baseEndpoint,
	}
}

// ListCloudProviders fetches the list of cloud providers and their regions
func (c *CloudAPIClient) ListCloudProviders(ctx context.Context) ([]CloudProvider, error) {
	providersEndpoint := fmt.Sprintf("%s/api/cloud-regions", c.Endpoint)

	// Reuse the FronteggClient's HTTPClient which already includes the Authorization token.
	resp, err := c.FronteggClient.HTTPClient.Get(providersEndpoint)
	if err != nil {
		return nil, fmt.Errorf("error listing cloud providers: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to read response body: %v", err)
		}
		return nil, fmt.Errorf("cloud API returned non-200 status code: %d, body: %s", resp.StatusCode, string(body))
	}

	var response CloudProviderResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, err
	}

	log.Printf("[DEBUG] Cloud providers response body: %+v\n", response)

	return response.Data, nil
}

// GetRegionDetails fetches the details for a given region
func (c *CloudAPIClient) GetRegionDetails(ctx context.Context, provider CloudProvider) (*CloudRegion, error) {
	regionEndpoint := fmt.Sprintf("%s/api/region", provider.Url)

	resp, err := c.FronteggClient.HTTPClient.Get(regionEndpoint)
	if err != nil {
		return nil, fmt.Errorf("error retrieving region details: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to read response body: %v", err)
		}
		return nil, fmt.Errorf("cloud API returned non-200 status code: %d, body: %s", resp.StatusCode, string(body))
	}

	log.Printf("[DEBUG] Region details response body: %+v\n", resp.Body)

	var region CloudRegion
	if err := json.NewDecoder(resp.Body).Decode(&region); err != nil {
		return nil, err
	}

	log.Printf("[DEBUG] Region details response body: %+v\n", region)

	return &region, nil
}

// EnableRegion sends a PATCH request to enable a cloud region
func (c *CloudAPIClient) EnableRegion(ctx context.Context, provider CloudProvider) (*CloudRegion, error) {
	endpoint := fmt.Sprintf("%s/api/region", provider.Url)
	emptyJSONPayload := bytes.NewBuffer([]byte("{}"))
	req, err := http.NewRequestWithContext(ctx, http.MethodPatch, endpoint, emptyJSONPayload)
	if err != nil {
		return nil, fmt.Errorf("error creating request to enable region: %v", err)
	}

	req.Header.Add("Content-Type", "application/json")

	resp, err := c.FronteggClient.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error sending request to enable region: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to read response body: %v", err)
		}
		return nil, fmt.Errorf("cloud API returned non-200/201 status code: %d, body: %s", resp.StatusCode, string(body))
	}

	var region CloudRegion
	if err := json.NewDecoder(resp.Body).Decode(&region); err != nil {
		return nil, err
	}

	return &region, nil
}

// GetHost retrieves the SQL address for a specified region
func (c *CloudAPIClient) GetHost(ctx context.Context, regionID string) (string, error) {
	providers, err := c.ListCloudProviders(ctx)
	if err != nil {
		return "", err
	}

	var provider *CloudProvider
	for _, p := range providers {
		if p.ID == regionID {
			provider = &p
			break
		}
	}

	if provider == nil {
		return "", fmt.Errorf("provider for region '%s' not found", regionID)
	}

	region, err := c.GetRegionDetails(ctx, *provider)
	if err != nil {
		return "", err
	}

	if region.RegionInfo == nil || !region.RegionInfo.Resolvable {
		return "", fmt.Errorf("region '%s' is not enabled", regionID)
	}

	return region.RegionInfo.SqlAddress, nil
}

func SplitHostPort(hostPortStr string) (host string, port int, err error) {
	parts := strings.Split(hostPortStr, ":")
	switch len(parts) {
	case 1:
		// Only host is provided, return the default port
		return parts[0], 6875, nil
	case 2:
		// Both host and port are provided, return both
		port, err := strconv.Atoi(parts[1])
		if err != nil {
			return "", 0, fmt.Errorf("invalid port: %v", err)
		}
		return parts[0], port, nil
	default:
		// Invalid format
		return "", 0, fmt.Errorf("invalid host:port format")
	}
}
