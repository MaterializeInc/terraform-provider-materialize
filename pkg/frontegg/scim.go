package frontegg

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/MaterializeInc/terraform-provider-materialize/pkg/clients"
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

// CreateSCIM2Configuration creates a new SCIM 2.0 configuration
func CreateSCIM2Configuration(ctx context.Context, client *clients.FronteggClient, config SCIM2Configuration) (*SCIM2Configuration, error) {
	configData, err := json.Marshal(config)
	if err != nil {
		return nil, err
	}

	endpoint := fmt.Sprintf("%s/frontegg/directory/resources/v1/configurations/scim2", client.Endpoint)
	req, err := http.NewRequestWithContext(ctx, "POST", endpoint, bytes.NewBuffer(configData))
	if err != nil {
		return nil, err
	}

	req.Header.Add("Authorization", "Bearer "+client.Token)
	req.Header.Add("Content-Type", "application/json")

	resp, err := client.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		var sb strings.Builder
		_, err = io.Copy(&sb, resp.Body)
		if err != nil {
			return nil, err
		}
		return nil, fmt.Errorf("error creating SCIM 2.0 configuration: status %d, response: %s", resp.StatusCode, sb.String())
	}

	var newConfig SCIM2Configuration
	if err := json.NewDecoder(resp.Body).Decode(&newConfig); err != nil {
		return nil, err
	}

	return &newConfig, nil
}

// DeleteSCIM2Configuration deletes an existing SCIM 2.0 configuration
func DeleteSCIM2Configuration(ctx context.Context, client *clients.FronteggClient, id string) error {
	endpoint := fmt.Sprintf("%s/frontegg/directory/resources/v1/configurations/scim2/%s", client.Endpoint, id)
	req, err := http.NewRequestWithContext(ctx, "DELETE", endpoint, nil)
	if err != nil {
		return err
	}

	req.Header.Add("Authorization", "Bearer "+client.Token)

	resp, err := client.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("error deleting SCIM 2.0 configuration: status %d", resp.StatusCode)
	}

	return nil
}
