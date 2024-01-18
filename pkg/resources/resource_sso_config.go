package resources

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/MaterializeInc/terraform-provider-materialize/pkg/utils"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

var SSOConfigSchema = map[string]*schema.Schema{
	"enabled": {
		Type:        schema.TypeBool,
		Required:    true,
		Description: "Whether SSO is enabled or not. If enabled, users will be redirected to the SSO endpoint for authentication. The configuration needs to be valid for SSO to work.",
	},
	"sso_endpoint": {
		Type:        schema.TypeString,
		Required:    true,
		Description: " The URL endpoint for the SSO service. This is the URL that users will be redirected to for authentication. The URL must be accessible from the browser.",
	},
	"public_certificate": {
		Type:        schema.TypeString,
		Required:    true,
		Description: "The public certificate of the SSO service. This is used to verify the SSO response. The certificate must be in PEM format. The certificate must be accessible from the browser. If the certificate is not accessible from the browser, you can use the public certificate of the Identity Provider (IdP) instead.",
	},
	"sign_request": {
		Type:        schema.TypeBool,
		Required:    true,
		Description: "Indicates whether the SSO request needs to be digitally signed.",
	},
	"type": {
		Type:        schema.TypeString,
		Required:    true,
		Description: "Defines the type of SSO protocol being used (e.g., SAML, OIDC).",
	},
	"oidc_client_id": {
		Type:        schema.TypeString,
		Optional:    true,
		Description: "The client ID of the OIDC application. This is used to identify the application to the OIDC service. This is required if the type is OIDC.",
	},
	"oidc_secret": {
		Type:        schema.TypeString,
		Optional:    true,
		Description: "The client secret of the OIDC application. This is used to authenticate the application to the OIDC service. This is required if the type is OIDC.",
	},
}

// SSOConfig represents the structure for SSO configuration.
type SSOConfig struct {
	Id                string `json:"id"`
	Enabled           bool   `json:"enabled"`
	SsoEndpoint       string `json:"ssoEndpoint"`
	PublicCertificate string `json:"publicCertificate"`
	SignRequest       bool   `json:"signRequest"`
	AcsUrl            string `json:"acsUrl"`
	SpEntityId        string `json:"spEntityId"`
	Type              string `json:"type"`
	OidcClientId      string `json:"oidcClientId"`
	OidcSecret        string `json:"oidcSecret"`
	Domains           []Domain
}

func SSOConfiguration() *schema.Resource {
	return &schema.Resource{
		CreateContext: ssoConfigCreate,
		ReadContext:   ssoConfigRead,
		UpdateContext: ssoConfigUpdate,
		DeleteContext: ssoConfigDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: SSOConfigSchema,
	}
}

func ssoConfigCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerMeta, err := utils.GetProviderMeta(meta)
	if err != nil {
		return diag.FromErr(err)
	}
	client := providerMeta.Frontegg

	// Create the SSO configuration first
	ssoConfig := SSOConfig{
		Enabled:           d.Get("enabled").(bool),
		SsoEndpoint:       d.Get("sso_endpoint").(string),
		PublicCertificate: d.Get("public_certificate").(string),
		SignRequest:       d.Get("sign_request").(bool),
		Type:              d.Get("type").(string),
		OidcClientId:      d.Get("oidc_client_id").(string),
		OidcSecret:        d.Get("oidc_secret").(string),
		AcsUrl:            fmt.Sprintf("%s/auth/saml/callback", client.Endpoint),
		SpEntityId:        fmt.Sprintf("%s/auth/saml/metadata", client.Endpoint),
	}

	requestBody, err := json.Marshal(ssoConfig)
	if err != nil {
		return diag.FromErr(err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", fmt.Sprintf("%s/frontegg/team/resources/sso/v1/configurations", client.Endpoint), bytes.NewBuffer(requestBody))
	if err != nil {
		return diag.FromErr(err)
	}

	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", "Bearer "+client.Token)

	resp, err := client.HTTPClient.Do(req)
	if err != nil {
		return diag.FromErr(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		responseData, err := io.ReadAll(resp.Body)
		if err != nil {
			return diag.FromErr(err)
		}
		return diag.Errorf("error creating SSO configuration: status %d, response: %s", resp.StatusCode, string(responseData))
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return diag.FromErr(err)
	}

	ssoConfigID, ok := result["id"].(string)
	if !ok {
		return diag.Errorf("error retrieving ID from SSO configuration creation response")
	}

	log.Printf("[DEBUG] SSO configuration create response ID: %s", ssoConfigID)
	d.SetId(ssoConfigID)

	return ssoConfigRead(ctx, d, meta)
}

func ssoConfigRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerMeta, err := utils.GetProviderMeta(meta)
	if err != nil {
		return diag.FromErr(err)
	}
	client := providerMeta.Frontegg

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

	// Find the specific configuration
	var foundConfig *SSOConfig
	for _, config := range configs {
		if config.Id == d.Id() {
			foundConfig = &config
			break
		}
	}

	if foundConfig == nil {
		d.SetId("")
		log.Printf("[DEBUG] SSO configuration read response ID: %s", d.Id())
		return nil
	}

	decodedPublicCertificate, err := base64.StdEncoding.DecodeString(foundConfig.PublicCertificate)
	if err != nil {
		return diag.FromErr(err)
	}

	// Set the Terraform resource data
	d.Set("enabled", foundConfig.Enabled)
	d.Set("sso_endpoint", foundConfig.SsoEndpoint)
	d.Set("public_certificate", string(decodedPublicCertificate))
	d.Set("sign_request", foundConfig.SignRequest)
	d.Set("type", foundConfig.Type)
	d.Set("oidc_client_id", foundConfig.OidcClientId)
	d.Set("oidc_secret", foundConfig.OidcSecret)

	return nil
}

func ssoConfigUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerMeta, err := utils.GetProviderMeta(meta)
	if err != nil {
		return diag.FromErr(err)
	}
	client := providerMeta.Frontegg

	// Prepare the request body with the updated fields
	ssoConfig := SSOConfig{
		Enabled:           d.Get("enabled").(bool),
		SsoEndpoint:       d.Get("sso_endpoint").(string),
		PublicCertificate: d.Get("public_certificate").(string),
		SignRequest:       d.Get("sign_request").(bool),
		Type:              d.Get("type").(string),
		OidcClientId:      d.Get("oidc_client_id").(string),
		OidcSecret:        d.Get("oidc_secret").(string),
		AcsUrl:            fmt.Sprintf("%s/auth/saml/callback", client.Endpoint),
		SpEntityId:        fmt.Sprintf("%s/auth/saml/metadata", client.Endpoint),
	}

	requestBody, err := json.Marshal(ssoConfig)
	if err != nil {
		return diag.FromErr(err)
	}

	// Use PATCH method for the update
	req, err := http.NewRequestWithContext(ctx, "PATCH", fmt.Sprintf("%s/frontegg/team/resources/sso/v1/configurations/%s", client.Endpoint, d.Id()), bytes.NewBuffer(requestBody))
	if err != nil {
		return diag.FromErr(err)
	}

	req.Header.Add("Content-Type", "application/json")
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
		return diag.Errorf("error updating SSO configuration: status %d, response: %s", resp.StatusCode, string(responseData))
	}

	return ssoConfigRead(ctx, d, meta)
}

func ssoConfigDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerMeta, err := utils.GetProviderMeta(meta)
	if err != nil {
		return diag.FromErr(err)
	}
	client := providerMeta.Frontegg

	req, err := http.NewRequestWithContext(ctx, "DELETE", fmt.Sprintf("%s/frontegg/team/resources/sso/v1/configurations/%s", client.Endpoint, d.Id()), nil)
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
		return diag.Errorf("error deleting SSO configuration: status %d, response: %s", resp.StatusCode, string(responseData))
	}

	d.SetId("")
	return nil
}
