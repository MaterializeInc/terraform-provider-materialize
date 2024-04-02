package frontegg

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/MaterializeInc/terraform-provider-materialize/pkg/clients"
)

const (
	SCIM2ConfigurationsApiPathV1 = "/frontegg/directory/resources/v1/configurations/scim2"
)

// SCIM 2.0 Configurations API response
type SCIM2Configuration struct {
	ID                   string    `json:"id"`
	Source               string    `json:"source"`
	TenantID             string    `json:"tenantId"`
	ConnectionName       string    `json:"connectionName"`
	SyncToUserManagement bool      `json:"syncToUserManagement"`
	CreatedAt            time.Time `json:"createdAt"`
	Token                string    `json:"token"`
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
			"created_at":              config.CreatedAt.Format(time.RFC3339),
		}
		flattenedConfigurations = append(flattenedConfigurations, flattenedConfig)
	}
	return flattenedConfigurations
}

// FetchSCIM2Configurations fetches the SCIM 2.0 configurations
func FetchSCIM2Configurations(ctx context.Context, client *clients.FronteggClient) (SCIM2ConfigurationsResponse, error) {
	endpoint := fmt.Sprintf("%s%s", client.Endpoint, SCIM2ConfigurationsApiPathV1)
	resp, err := doRequest(ctx, client, "GET", endpoint, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var configurations SCIM2ConfigurationsResponse
	if err := json.NewDecoder(resp.Body).Decode(&configurations); err != nil {
		return nil, fmt.Errorf("error decoding SCIM 2.0 configurations: %v", err)
	}

	return configurations, nil
}

// CreateSCIM2Configuration creates a new SCIM 2.0 configuration
func CreateSCIM2Configuration(ctx context.Context, client *clients.FronteggClient, config SCIM2Configuration) (*SCIM2Configuration, error) {
	configData, err := jsonEncode(config)
	if err != nil {
		return nil, err
	}

	endpoint := fmt.Sprintf("%s%s", client.Endpoint, SCIM2ConfigurationsApiPathV1)
	resp, err := doRequest(ctx, client, "POST", endpoint, configData)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var newConfig SCIM2Configuration
	if err := json.NewDecoder(resp.Body).Decode(&newConfig); err != nil {
		return nil, fmt.Errorf("error decoding new SCIM 2.0 configuration: %v", err)
	}

	return &newConfig, nil
}

// DeleteSCIM2Configuration deletes an existing SCIM 2.0 configuration
func DeleteSCIM2Configuration(ctx context.Context, client *clients.FronteggClient, id string) error {
	endpoint := fmt.Sprintf("%s%s/%s", client.Endpoint, SCIM2ConfigurationsApiPathV1, id)
	resp, err := doRequest(ctx, client, "DELETE", endpoint, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("error deleting SCIM 2.0 configuration: status %d", resp.StatusCode)
	}

	return nil
}
