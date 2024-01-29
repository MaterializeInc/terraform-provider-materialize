package frontegg

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/MaterializeInc/terraform-provider-materialize/pkg/clients"
)

// SCIM 2.0 Configurations API response
type SCIM2Configuration struct {
	ID                   string `json:"id"`
	Source               string `json:"source"`
	TenantID             string `json:"tenantId"`
	ConnectionName       string `json:"connectionName"`
	SyncToUserManagement bool   `json:"syncToUserManagement"`
	CreatedAt            string `json:"createdAt"`
	Token                string `json:"token"`
}

type SCIM2ConfigurationsResponse []SCIM2Configuration

// Helper function to flatten the SCIM 2.0 configurations data
func FlattenSCIM2Configurations(configurations SCIM2ConfigurationsResponse) []interface{} {
	var flattenedConfigurations []interface{}
	for _, config := range configurations {
		flattenedConfig := map[string]interface{}{
			"id":                      config.ID,
			"source":                  config.Source,
			"tenant_id":               config.TenantID,
			"connection_name":         config.ConnectionName,
			"sync_to_user_management": config.SyncToUserManagement,
			"created_at":              config.CreatedAt,
		}
		flattenedConfigurations = append(flattenedConfigurations, flattenedConfig)
	}
	return flattenedConfigurations
}

// FetchSCIM2Configurations fetches the SCIM 2.0 configurations
func FetchSCIM2Configurations(ctx context.Context, client *clients.FronteggClient) (SCIM2ConfigurationsResponse, error) {
	endpoint := fmt.Sprintf("%s/frontegg/directory/resources/v1/configurations/scim2", client.Endpoint)
	req, err := http.NewRequestWithContext(ctx, "GET", endpoint, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Authorization", "Bearer "+client.Token)

	resp, err := client.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var sb strings.Builder
		_, err = io.Copy(&sb, resp.Body)
		if err != nil {
			return nil, err
		}
		return nil, fmt.Errorf("error reading SCIM 2.0 configurations: status %d, response: %s", resp.StatusCode, sb.String())
	}

	var configurations SCIM2ConfigurationsResponse
	if err := json.NewDecoder(resp.Body).Decode(&configurations); err != nil {
		return nil, err
	}

	return configurations, nil
}
