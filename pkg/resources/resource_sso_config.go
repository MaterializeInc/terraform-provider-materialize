package resources

import (
	"context"
	"encoding/base64"
	"fmt"
	"log"

	"github.com/MaterializeInc/terraform-provider-materialize/pkg/frontegg"
	"github.com/MaterializeInc/terraform-provider-materialize/pkg/utils"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
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
		Description: "The URL endpoint for the SSO service. This is the URL that users will be redirected to for authentication. The URL must be accessible from the browser.",
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
		Type:         schema.TypeString,
		Required:     true,
		Description:  "Defines the type of SSO protocol being used (e.g., saml, oidc).",
		ValidateFunc: validation.StringInSlice(ssoConfigTypes, true),
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

		Description: "The SSO configuration resource allows you to create, read, update, and delete SSO configurations.",
	}
}

func ssoConfigCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerMeta, err := utils.GetProviderMeta(meta)
	if err != nil {
		return diag.FromErr(err)
	}
	client := providerMeta.Frontegg
	baseEndpoint := providerMeta.CloudAPI.BaseEndpoint

	// Create the SSO configuration using the frontegg package
	ssoConfig := frontegg.SSOConfig{
		Enabled:           d.Get("enabled").(bool),
		SsoEndpoint:       d.Get("sso_endpoint").(string),
		PublicCertificate: d.Get("public_certificate").(string),
		SignRequest:       d.Get("sign_request").(bool),
		Type:              d.Get("type").(string),
		OidcClientId:      d.Get("oidc_client_id").(string),
		OidcSecret:        d.Get("oidc_secret").(string),
		AcsUrl:            fmt.Sprintf("%s/auth/saml/callback", client.Endpoint),
		SpEntityId:        fmt.Sprintf("%s/auth/saml/metadata", baseEndpoint),
	}

	newConfig, err := frontegg.CreateSSOConfiguration(ctx, client, ssoConfig)
	if err != nil {
		return diag.FromErr(err)
	}

	log.Printf("[DEBUG] SSO configuration create response ID: %s", newConfig.Id)
	d.SetId(newConfig.Id)

	return ssoConfigRead(ctx, d, meta)
}

func ssoConfigRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerMeta, err := utils.GetProviderMeta(meta)
	if err != nil {
		return diag.FromErr(err)
	}
	client := providerMeta.Frontegg

	// Fetch SSO configurations
	configurations, err := frontegg.FetchSSOConfigurations(ctx, client)
	if err != nil {
		return diag.FromErr(err)
	}

	// Find the matching configuration
	var foundConfig *frontegg.SSOConfig
	for _, config := range configurations {
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
	baseEndpoint := providerMeta.CloudAPI.BaseEndpoint

	// Prepare the updated SSO configuration
	ssoConfig := frontegg.SSOConfig{
		Id:                d.Id(),
		Enabled:           d.Get("enabled").(bool),
		SsoEndpoint:       d.Get("sso_endpoint").(string),
		PublicCertificate: d.Get("public_certificate").(string),
		SignRequest:       d.Get("sign_request").(bool),
		Type:              d.Get("type").(string),
		OidcClientId:      d.Get("oidc_client_id").(string),
		OidcSecret:        d.Get("oidc_secret").(string),
		AcsUrl:            fmt.Sprintf("%s/auth/saml/callback", client.Endpoint),
		SpEntityId:        fmt.Sprintf("%s/auth/saml/metadata", baseEndpoint),
	}

	updatedConfig, err := frontegg.UpdateSSOConfiguration(ctx, client, ssoConfig)
	if err != nil {
		return diag.FromErr(err)
	}

	log.Printf("[DEBUG] SSO configuration updated, ID: %s", updatedConfig.Id)
	return ssoConfigRead(ctx, d, meta)
}

func ssoConfigDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerMeta, err := utils.GetProviderMeta(meta)
	if err != nil {
		return diag.FromErr(err)
	}
	client := providerMeta.Frontegg

	err = frontegg.DeleteSSOConfiguration(ctx, client, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	log.Printf("[DEBUG] SSO configuration deleted, ID: %s", d.Id())
	d.SetId("")
	return nil
}
