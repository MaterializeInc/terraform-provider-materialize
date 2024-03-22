package resources

import (
	"context"
	"fmt"
	"log"

	"github.com/MaterializeInc/terraform-provider-materialize/pkg/frontegg"
	"github.com/MaterializeInc/terraform-provider-materialize/pkg/utils"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

var resourceSCIM2ConfigurationsSchema = map[string]*schema.Schema{
	"source": {
		Type:         schema.TypeString,
		Required:     true,
		Description:  "The source of the SCIM 2.0 configuration. Supported values are `okta`, `azure-ad`, and `other`.",
		ForceNew:     true,
		ValidateFunc: validation.StringInSlice(scim2ConfigSources, true),
	},
	"connection_name": {
		Type:        schema.TypeString,
		Required:    true,
		Description: "The name of the SCIM 2.0 connection. It must be unique.",
		ForceNew:    true,
	},
	"tenant_id": {
		Type:     schema.TypeString,
		Computed: true,
	},
	"sync_to_user_management": {
		Type:        schema.TypeBool,
		Computed:    true,
		Description: "Indicates whether automatic synchronization of data with the IdP is enabled, ensuring that changes in details or status in the IdP are updated accordingly.",
	},
	"token": {
		Type:        schema.TypeString,
		Computed:    true,
		Description: "The token of the SCIM 2.0 configuration.",
		Sensitive:   true,
	},
	"provisioning_url": {
		Type:        schema.TypeString,
		Computed:    true,
		Description: "The provisioning URL of the SCIM 2.0 configuration.",
		Sensitive:   true,
	},
	"created_at": {
		Type:        schema.TypeString,
		Computed:    true,
		Description: "The creation timestamp of the SCIM 2.0 configuration.",
	},
}

func SCIM2Configuration() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceSCIM2ConfigurationsCreate,
		ReadContext:   resourceSCIM2ConfigurationsRead,
		DeleteContext: resourceSCIM2ConfigurationsDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: resourceSCIM2ConfigurationsSchema,

		Description: "The SCIM 2.0 configurations resource allows you to create, read, and delete the SCIM 2.0 configurations.",
	}
}

func resourceSCIM2ConfigurationsCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerMeta, err := utils.GetProviderMeta(meta)
	if err != nil {
		return diag.FromErr(err)
	}
	client := providerMeta.Frontegg

	config := frontegg.SCIM2Configuration{
		Source:               d.Get("source").(string),
		ConnectionName:       d.Get("connection_name").(string),
		SyncToUserManagement: true,
	}

	newConfig, err := frontegg.CreateSCIM2Configuration(ctx, client, config)
	if err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("token", newConfig.Token); err != nil {
		return diag.FromErr(err)
	}

	// Construct and set the Provisioning URL
	provisioningURL := fmt.Sprintf("%s/frontegg/directory/resources/scim/v2.0/%s", client.Endpoint, newConfig.ID)
	if err := d.Set("provisioning_url", provisioningURL); err != nil {
		return diag.FromErr(err)
	}

	d.SetId(newConfig.ID)
	return resourceSCIM2ConfigurationsRead(ctx, d, meta)
}

// Delete an existing SCIM 2.0 configuration
func resourceSCIM2ConfigurationsDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerMeta, err := utils.GetProviderMeta(meta)
	if err != nil {
		return diag.FromErr(err)
	}
	client := providerMeta.Frontegg

	err = frontegg.DeleteSCIM2Configuration(ctx, client, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId("")
	return nil
}

func resourceSCIM2ConfigurationsRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerMeta, err := utils.GetProviderMeta(meta)
	if err != nil {
		return diag.FromErr(err)
	}
	client := providerMeta.Frontegg

	configurations, err := frontegg.FetchSCIM2Configurations(ctx, client)
	if err != nil {
		return diag.FromErr(err)
	}

	// Find the configuration with the matching ID
	var config frontegg.SCIM2Configuration
	for _, configuration := range configurations {
		if configuration.ID == d.Id() {
			config = configuration
			break
		}
	}

	if config.ID == "" {
		d.SetId("")
		log.Printf("[WARN] SCIM 2.0 configuration with ID %s not found", d.Id())
		return nil
	}

	if err := d.Set("source", config.Source); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("connection_name", config.ConnectionName); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("tenant_id", config.TenantID); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("sync_to_user_management", config.SyncToUserManagement); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("created_at", config.CreatedAt.String()); err != nil {
		return diag.FromErr(err)
	}

	// Log the response
	log.Printf("[DEBUG] SCIM 2.0 configuration: %+v", config)

	return nil
}
