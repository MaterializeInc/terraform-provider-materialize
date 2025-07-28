package resources

import (
	"context"
	"fmt"
	"strings"

	"github.com/MaterializeInc/terraform-provider-materialize/pkg/frontegg"
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
		Description: "The domain name for the SSO domain configuration. This domain will be used to validate the SSO configuration and needs to be unique across all SSO configurations.",
	},
	"validated": {
		Type:        schema.TypeBool,
		Computed:    true,
		Description: "Indicates whether the domain has been validated.",
	},
}

func SSODomain() *schema.Resource {
	return &schema.Resource{
		CreateContext: ssoDomainCreate,
		ReadContext:   ssoDomainRead,
		UpdateContext: ssoDomainUpdate,
		DeleteContext: ssoDomainDelete,

		Importer: &schema.ResourceImporter{
			StateContext: ssoDomainImport,
		},

		Schema: SSODomainSchema,

		Description: "The SSO domain resource allows you to set the domain for an SSO configuration. This domain will be used to validate the SSO configuration.",
	}
}

func ssoDomainCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerMeta, err := utils.GetProviderMeta(meta)
	if err != nil {
		return diag.FromErr(err)
	}

	// Validate that SSO domains are only managed in SaaS mode
	if diags := providerMeta.ValidateSaaSOnly("materialize_sso_domain"); diags.HasError() {
		return diags
	}

	client := providerMeta.Frontegg

	ssoConfigID := d.Get("sso_config_id").(string)
	domainName := d.Get("domain").(string)

	domain, err := frontegg.CreateSSODomain(ctx, client, ssoConfigID, domainName)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(domain.ID)
	return ssoDomainRead(ctx, d, meta)
}

func ssoDomainRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerMeta, err := utils.GetProviderMeta(meta)
	if err != nil {
		return diag.FromErr(err)
	}
	client := providerMeta.Frontegg

	ssoConfigID := d.Get("sso_config_id").(string)
	domainName := d.Get("domain").(string)

	domain, err := frontegg.FetchSSODomain(ctx, client, ssoConfigID, domainName)
	if err != nil {
		if err.Error() == "domain not found" {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	d.SetId(domain.ID)
	d.Set("validated", domain.Validated)
	return nil
}

func ssoDomainUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerMeta, err := utils.GetProviderMeta(meta)
	if err != nil {
		return diag.FromErr(err)
	}
	client := providerMeta.Frontegg

	ssoConfigID := d.Get("sso_config_id").(string)
	domainName := d.Get("domain").(string)

	err = frontegg.DeleteSSODomain(ctx, client, ssoConfigID, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	newDomain, err := frontegg.CreateSSODomain(ctx, client, ssoConfigID, domainName)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(newDomain.ID)
	return ssoDomainRead(ctx, d, meta)
}

func ssoDomainDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerMeta, err := utils.GetProviderMeta(meta)
	if err != nil {
		return diag.FromErr(err)
	}
	client := providerMeta.Frontegg

	ssoConfigID := d.Get("sso_config_id").(string)

	err = frontegg.DeleteSSODomain(ctx, client, ssoConfigID, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId("")
	return nil
}

func ssoDomainImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	compositeID := d.Id()
	parts := strings.Split(compositeID, ":")

	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid format of ID (%s), expected ssoConfigID:domainName", compositeID)
	}

	d.Set("sso_config_id", parts[0])
	d.Set("domain", parts[1])

	diags := ssoDomainRead(ctx, d, meta)
	if diags.HasError() {
		var err error
		for _, d := range diags {
			if d.Severity == diag.Error {
				if err == nil {
					err = fmt.Errorf(d.Summary)
				} else {
					err = fmt.Errorf("%v; %s", err, d.Summary)
				}
			}
		}
		return nil, err
	}

	// If the domain ID is not set, return an error
	if d.Id() == "" {
		return nil, fmt.Errorf("domain not found")
	}

	return []*schema.ResourceData{d}, nil
}
