package frontegg

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/MaterializeInc/terraform-provider-materialize/pkg/clients"
)

const (
	SSODomainsApiPathV1 = "/frontegg/team/resources/sso/v1/configurations"
)

// Domain represents the structure for SSO domain.
type Domain struct {
	ID          string `json:"id"`
	Domain      string `json:"domain"`
	Validated   bool   `json:"validated"`
	SsoConfigId string `json:"ssoConfigId"`
}

// FetchSSODomain fetches a specific SSO domain.
func FetchSSODomain(ctx context.Context, client *clients.FronteggClient, configID, domainName string) (*Domain, error) {
	endpoint := fmt.Sprintf("%s%s", client.Endpoint, SSODomainsApiPathV1)
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
		responseData, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		return nil, fmt.Errorf("error reading SSO configurations: status %d, response: %s", resp.StatusCode, string(responseData))
	}

	var configs []SSOConfig
	if err := json.NewDecoder(resp.Body).Decode(&configs); err != nil {
		return nil, err
	}

	for _, config := range configs {
		if config.Id == configID {
			for _, domain := range config.Domains {
				if domain.Domain == domainName {
					return &domain, nil
				}
			}
		}
	}

	return nil, fmt.Errorf("domain not found")
}

// CreateDomain creates a new SSO domain.
func CreateSSODomain(ctx context.Context, client *clients.FronteggClient, configID, domainName string) (*Domain, error) {
	endpoint := fmt.Sprintf("%s%s/%s/domains", client.Endpoint, SSODomainsApiPathV1, configID)
	payload := map[string]string{"domain": domainName}
	requestBody, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", endpoint, bytes.NewBuffer(requestBody))
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
		responseData, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		return nil, fmt.Errorf("error creating SSO domain: status %d, response: %s", resp.StatusCode, string(responseData))
	}

	var result Domain
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return &result, nil
}

// DeleteSSODomain deletes a specific SSO domain.
func DeleteSSODomain(ctx context.Context, client *clients.FronteggClient, configID, domainID string) error {
	endpoint := fmt.Sprintf("%s%s/%s/domains/%s", client.Endpoint, SSODomainsApiPathV1, configID, domainID)

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

	if resp.StatusCode != http.StatusOK {
		responseData, err := io.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		return fmt.Errorf("error deleting SSO domain: status %d, response: %s", resp.StatusCode, string(responseData))
	}

	return nil
}
