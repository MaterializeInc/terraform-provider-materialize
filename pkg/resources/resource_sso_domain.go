package resources

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/MaterializeInc/terraform-provider-materialize/pkg/clients"
	"github.com/MaterializeInc/terraform-provider-materialize/pkg/utils"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

var SSODomainSchema = map[string]*schema.Schema{
	"sso_config_id": {
		Type:        schema.TypeString,
		Required:    true,
		Description: "The ID of the associated SSO configuration.",
	},
	"domain": {
		Type:        schema.TypeString,
		Required:    true,
		Description: "The domain name for the SSO domain configuration.",
	},
	"validated": {
		Type:        schema.TypeBool,
		Computed:    true,
		Description: "Indicates whether the domain has been validated.",
	},
}

// Domain represents the structure for SSO domain.
type Domain struct {
	ID          string `json:"id"`
	Domain      string `json:"domain"`
	Validated   bool   `json:"validated"`
	SsoConfigId string `json:"sso_config_id"`
}

func SSODomain() *schema.Resource {
	return &schema.Resource{
		CreateContext: ssoDomainCreate,
		ReadContext:   ssoDomainRead,
		UpdateContext: ssoDomainUpdate,
		DeleteContext: ssoDomainDelete,

		Schema: SSODomainSchema,
	}
}

func ssoDomainCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerMeta, err := utils.GetProviderMeta(meta)
	if err != nil {
		return diag.FromErr(err)
	}
	client := providerMeta.Frontegg

	ssoConfigID := d.Get("sso_config_id").(string)
	domainName := d.Get("domain").(string)

	domainID, err := createDomain(ctx, client, ssoConfigID, domainName)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(domainID)
	return ssoDomainRead(ctx, d, meta)
}

func ssoDomainRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerMeta, err := utils.GetProviderMeta(meta)
	if err != nil {
		return diag.FromErr(err)
	}
	client := providerMeta.Frontegg

	// Make the request to the SSO configurations endpoint as there is no specific endpoint for domains
	req, err := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("%s/frontegg/team/resources/sso/v1/configurations", client.Endpoint), nil)
	if err != nil {
		return diag.FromErr(err)
	}
	req.Header.Add("Authorization", "Bearer "+client.Token)

	resp, err := client.HTTPClient.Do(req)
	if err != nil {
		return diag.FromErr(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		responseData, err := io.ReadAll(resp.Body)
		if err != nil {
			return diag.FromErr(err)
		}
		return diag.Errorf("error reading SSO configurations: status %d, response: %s", resp.StatusCode, string(responseData))
	}

	var configs []SSOConfig
	if err := json.NewDecoder(resp.Body).Decode(&configs); err != nil {
		return diag.FromErr(err)
	}

	// Extract sso_config_id and domain from the Terraform resource data
	ssoConfigID := d.Get("sso_config_id").(string)
	domainName := d.Get("domain").(string)

	// Iterate over the configurations to find the specific domain
	for _, config := range configs {
		if config.Id == ssoConfigID {
			for _, domain := range config.Domains {
				if domain.Domain == domainName {
					// Set the Terraform resource data from the domain
					d.Set("id", domain.ID)
					d.Set("validated", domain.Validated)
					return nil
				}
			}
		}
	}

	// If domain not found, set the resource ID to empty to indicate it doesn't exist
	d.SetId("")
	return nil
}

func ssoDomainUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerMeta, err := utils.GetProviderMeta(meta)
	if err != nil {
		return diag.FromErr(err)
	}
	client := providerMeta.Frontegg

	// Extract the sso_config_id and domain
	ssoConfigID := d.Get("sso_config_id").(string)
	domainName := d.Get("domain").(string)

	// Delete the existing domain
	err = deleteDomain(ctx, client, ssoConfigID, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	// Create the new domain with the updated details
	newDomainID, err := createDomain(ctx, client, ssoConfigID, domainName)
	if err != nil {
		return diag.FromErr(err)
	}

	// Update the Terraform resource ID to the new domain ID
	d.SetId(newDomainID)

	return ssoDomainRead(ctx, d, meta)
}

func ssoDomainDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerMeta, err := utils.GetProviderMeta(meta)
	if err != nil {
		return diag.FromErr(err)
	}
	client := providerMeta.Frontegg

	ssoConfigID := d.Get("sso_config_id").(string)
	domainName := d.Get("domain").(string)
	log.Printf("[DEBUG] Domain name: %s", domainName)

	err = deleteDomain(ctx, client, ssoConfigID, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId("")
	return nil
}

func createDomain(ctx context.Context, client *clients.FronteggClient, configID string, domainName string) (string, error) {
	endpoint := fmt.Sprintf("%s/frontegg/team/resources/sso/v1/configurations/%s/domains", client.Endpoint, configID)
	payload := map[string]string{"domain": domainName}

	requestBody, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", endpoint, bytes.NewBuffer(requestBody))
	if err != nil {
		return "", err
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", "Bearer "+client.Token)

	resp, err := client.HTTPClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		responseData, err := io.ReadAll(resp.Body)
		if err != nil {
			return "", err
		}

		// Handle the specific case where the domain already exists
		if resp.StatusCode == http.StatusConflict {
			return "", fmt.Errorf("error creating domain: domain '%s' already exists in another configuration", domainName)
		}

		return "", fmt.Errorf("error creating domain: status %d, response: %s", resp.StatusCode, string(responseData))
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	domainID, ok := result["id"].(string)
	if !ok {
		return "", fmt.Errorf("error retrieving ID from domain creation response")
	}

	log.Printf("[DEBUG] Domain create response ID: %s", domainID)

	return domainID, nil
}

func deleteDomain(ctx context.Context, client *clients.FronteggClient, configID string, domainId string) error {
	endpoint := fmt.Sprintf("%s/frontegg/team/resources/sso/v1/configurations/%s/domains/%s", client.Endpoint, configID, domainId)

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
		return fmt.Errorf("error deleting domain: status %d, response: %s", resp.StatusCode, string(responseData))
	}

	return nil
}
