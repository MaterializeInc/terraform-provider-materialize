package clients

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
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
	HTTPClient    *http.Client
	Authenticator Authenticator
	Endpoint      string
	BaseEndpoint  string
}

// NewCloudAPIClient creates a new Cloud API client
func NewCloudAPIClient(authenticator Authenticator, cloudAPIEndpoint, baseEndpoint string) *CloudAPIClient {
	return &CloudAPIClient{
		HTTPClient:    &http.Client{},
		Authenticator: authenticator,
		Endpoint:      cloudAPIEndpoint,
		BaseEndpoint:  baseEndpoint,
	}
}

// ListCloudProviders fetches the list of cloud providers and their regions
func (c *CloudAPIClient) ListCloudProviders(ctx context.Context) ([]CloudProvider, error) {
	providersEndpoint := fmt.Sprintf("%s/api/cloud-regions", c.Endpoint)

	resp, err := c.doRequest(ctx, http.MethodGet, providersEndpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("error listing cloud providers: %w", err)
	}
	defer resp.Body.Close()

	var response CloudProviderResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("error decoding response: %w", err)
	}

	return response.Data, nil
}

// GetRegionDetails fetches the details for a given region
func (c *CloudAPIClient) GetRegionDetails(ctx context.Context, provider CloudProvider) (*CloudRegion, error) {
	regionEndpoint := fmt.Sprintf("%s/api/region", provider.Url)

	resp, err := c.doRequest(ctx, http.MethodGet, regionEndpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("error retrieving region details: %w", err)
	}
	defer resp.Body.Close()

	var region CloudRegion
	if err := json.NewDecoder(resp.Body).Decode(&region); err != nil {
		return nil, fmt.Errorf("error decoding region details: %w", err)
	}

	return &region, nil
}

// EnableRegion sends a PATCH request to enable a cloud region
func (c *CloudAPIClient) EnableRegion(ctx context.Context, provider CloudProvider) (*CloudRegion, error) {
	endpoint := fmt.Sprintf("%s/api/region", provider.Url)
	emptyJSONPayload := bytes.NewBuffer([]byte("{}"))

	resp, err := c.doRequest(ctx, http.MethodPatch, endpoint, emptyJSONPayload)
	if err != nil {
		return nil, fmt.Errorf("error enabling region: %w", err)
	}
	defer resp.Body.Close()

	var region CloudRegion
	if err := json.NewDecoder(resp.Body).Decode(&region); err != nil {
		return nil, fmt.Errorf("error decoding enabled region details: %w", err)
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

func (c *CloudAPIClient) doRequest(ctx context.Context, method, url string, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	if err := c.Authenticator.NeedsTokenRefresh(); err != nil {
		if err := c.Authenticator.RefreshToken(); err != nil {
			return nil, fmt.Errorf("error refreshing token: %w", err)
		}
	}

	req.Header.Set("Authorization", "Bearer "+c.Authenticator.GetToken())
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error sending request: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		return nil, &APIError{
			StatusCode: resp.StatusCode,
			Message:    string(body),
		}
	}

	return resp, nil
}

type APIError struct {
	StatusCode int
	Message    string
}

func (e *APIError) Error() string {
	return fmt.Sprintf("API error: %d - %s", e.StatusCode, e.Message)
}
