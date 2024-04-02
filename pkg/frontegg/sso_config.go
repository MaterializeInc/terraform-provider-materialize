package frontegg

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/MaterializeInc/terraform-provider-materialize/pkg/clients"
)

const (
	SSOConfigurationsApiPathV1 = "/frontegg/team/resources/sso/v1/configurations"
)

// SSOConfig represents the structure for SSO configuration.
type SSOConfig struct {
	Id                string    `json:"id"`
	Enabled           bool      `json:"enabled"`
	SsoEndpoint       string    `json:"ssoEndpoint"`
	PublicCertificate string    `json:"publicCertificate"`
	SignRequest       bool      `json:"signRequest"`
	AcsUrl            string    `json:"acsUrl"`
	SpEntityId        string    `json:"spEntityId"`
	Type              string    `json:"type"`
	OidcClientId      string    `json:"oidcClientId"`
	OidcSecret        string    `json:"oidcSecret"`
	CreatedAt         time.Time `json:"createdAt"`
	Domains           []Domain
}

type SSOConfigurationsResponse []SSOConfig

// Helper function to flatten the SSO configurations data
func FlattenSSOConfigurations(configurations SSOConfigurationsResponse) []interface{} {
	var flattenedConfigurations []interface{}
	for _, config := range configurations {
		flattenedConfig := map[string]interface{}{
			"id":                 config.Id,
			"enabled":            config.Enabled,
			"sso_endpoint":       config.SsoEndpoint,
			"public_certificate": config.PublicCertificate,
			"sign_request":       config.SignRequest,
			"acs_url":            config.AcsUrl,
			"sp_entity_id":       config.SpEntityId,
			"type":               config.Type,
			"oidc_client_id":     config.OidcClientId,
			"oidc_secret":        config.OidcSecret,
			"created_at":         config.CreatedAt.Format(time.RFC3339),
		}
		flattenedConfigurations = append(flattenedConfigurations, flattenedConfig)
	}
	return flattenedConfigurations
}

// FetchSSOConfigurations fetches the SSO configurations
func FetchSSOConfigurations(ctx context.Context, client *clients.FronteggClient) (SSOConfigurationsResponse, error) {
	endpoint := fmt.Sprintf("%s%s", client.Endpoint, SSOConfigurationsApiPathV1)
	resp, err := doRequest(ctx, client, "GET", endpoint, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var configurations SSOConfigurationsResponse
	if err := json.NewDecoder(resp.Body).Decode(&configurations); err != nil {
		return nil, err
	}

	return configurations, nil
}

// FetchSSOConfigurationsRaw fetches the raw SSO configurations
func FetchSSOConfigurationsRaw(ctx context.Context, client *clients.FronteggClient) ([]byte, error) {
	endpoint := fmt.Sprintf("%s%s", client.Endpoint, SSOConfigurationsApiPathV1)
	resp, err := doRequest(ctx, client, "GET", endpoint, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return io.ReadAll(resp.Body)
}

// CreateSSOConfiguration creates a new SSO configuration
func CreateSSOConfiguration(ctx context.Context, client *clients.FronteggClient, ssoConfig SSOConfig) (*SSOConfig, error) {
	requestBody, err := jsonEncode(ssoConfig)
	if err != nil {
		return nil, err
	}

	endpoint := fmt.Sprintf("%s%s", client.Endpoint, SSOConfigurationsApiPathV1)
	resp, err := doRequest(ctx, client, "POST", endpoint, requestBody)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var newConfig SSOConfig
	if err := json.NewDecoder(resp.Body).Decode(&newConfig); err != nil {
		return nil, err
	}

	return &newConfig, nil
}

// UpdateSSOConfiguration updates an existing SSO configuration
func UpdateSSOConfiguration(ctx context.Context, client *clients.FronteggClient, ssoConfig SSOConfig) (*SSOConfig, error) {
	requestBody, err := jsonEncode(ssoConfig)
	if err != nil {
		return nil, err
	}

	endpoint := fmt.Sprintf("%s%s/%s", client.Endpoint, SSOConfigurationsApiPathV1, ssoConfig.Id)
	resp, err := doRequest(ctx, client, "PATCH", endpoint, requestBody)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var updatedConfig SSOConfig
	if err := json.NewDecoder(resp.Body).Decode(&updatedConfig); err != nil {
		return nil, err
	}

	return &updatedConfig, nil
}

// DeleteSSOConfiguration deletes an existing SSO configuration
func DeleteSSOConfiguration(ctx context.Context, client *clients.FronteggClient, configId string) error {
	endpoint := fmt.Sprintf("%s%s/%s", client.Endpoint, SSOConfigurationsApiPathV1, configId)
	resp, err := doRequest(ctx, client, "DELETE", endpoint, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var sb strings.Builder
		_, err = io.Copy(&sb, resp.Body)
		if err != nil {
			return err
		}
		return fmt.Errorf("error deleting SSO configuration: status %d, response: %s", resp.StatusCode, sb.String())
	}

	return nil
}
