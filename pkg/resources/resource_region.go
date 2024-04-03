package resources

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/MaterializeInc/terraform-provider-materialize/pkg/clients"
	"github.com/MaterializeInc/terraform-provider-materialize/pkg/utils"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

var regionSchema = map[string]*schema.Schema{
	"region_id": {
		Type:        schema.TypeString,
		Required:    true,
		Description: "The ID of the region to manage. Example: aws/us-west-2",
	},
	"sql_address": {
		Type:        schema.TypeString,
		Computed:    true,
		Description: "The SQL address of the region.",
	},
	"http_address": {
		Type:        schema.TypeString,
		Computed:    true,
		Description: "The HTTP address of the region.",
	},
	"resolvable": {
		Type:        schema.TypeBool,
		Computed:    true,
		Description: "Indicates if the region is resolvable.",
	},
	"enabled_at": {
		Type:        schema.TypeString,
		Computed:    true,
		Description: "The timestamp when the region was enabled.",
	},
	"region_state": {
		Type:        schema.TypeBool,
		Computed:    true,
		Description: "The state of the region. True if enabled, false otherwise.",
	},
}

func Region() *schema.Resource {
	return &schema.Resource{
		Description: "The region resource allows you to manage regions in Materialize. " +
			"When a new region is created, it automatically includes an 'xsmall' quickstart cluster as part of the initialization process. " +
			"Users are billed for this quickstart cluster from the moment the region is created. " +
			"To avoid unnecessary charges, you can connect to the new region and drop the quickstart cluster if it is not needed. " +
			"Please note that disabling a region cannot be achieved directly through this provider. " +
			"If you need to disable a region, contact Materialize support for assistance. " +
			"This process ensures that any necessary cleanup and billing adjustments are handled properly.",
		CreateContext: resourceCloudRegionCreate,
		ReadContext:   resourceCloudRegionRead,
		UpdateContext: resourceCloudRegionUpdate,
		DeleteContext: resourceCloudRegionDelete,
		Schema:        regionSchema,
	}
}

func resourceCloudRegionCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerMeta, err := utils.GetProviderMeta(meta)
	if err != nil {
		return diag.FromErr(err)
	}
	client := providerMeta.CloudAPI

	regionID := d.Get("region_id").(string)

	providers, err := client.ListCloudProviders(ctx)
	if err != nil {
		return diag.FromErr(err)
	}

	var targetProvider *clients.CloudProvider
	for _, provider := range providers {
		if provider.ID == regionID {
			targetProvider = &provider
			break
		}
	}

	if targetProvider == nil {
		return diag.Errorf("region %s not found", regionID)
	}

	// Check if the region is already enabled
	region, err := client.GetRegionDetails(ctx, *targetProvider)
	if err != nil {
		if strings.Contains(err.Error(), "non-200 status code: 204") || strings.Contains(err.Error(), "region not found") {
			// The region is not enabled, so proceed to enable it
			log.Printf("[INFO] Enabling region %s", regionID)
			_, err := client.EnableRegion(ctx, *targetProvider)
			if err != nil {
				return diag.Errorf("error enabling region %s: %s", regionID, err)
			}

			// Wait for the region to be fully enabled and resolvable
			err = waitForRegionToBeEnabled(ctx, client, *targetProvider)
			if err != nil {
				return diag.FromErr(err)
			}
		} else {
			return diag.FromErr(err)
		}
	} else if region.RegionInfo == nil || region.RegionInfo.EnabledAt == "" {
		// If the region exists but isn't enabled, enable it
		_, err := client.EnableRegion(ctx, *targetProvider)
		if err != nil {
			return diag.Errorf("error enabling region %s: %s", regionID, err)
		}

		// Wait for the region to be fully enabled and resolvable
		err = waitForRegionToBeEnabled(ctx, client, *targetProvider)
		if err != nil {
			return diag.FromErr(err)
		}
	} else {
		log.Printf("[INFO] Region %s is already enabled", regionID)
	}

	d.SetId(regionID)
	return resourceCloudRegionRead(ctx, d, meta)
}

func resourceCloudRegionRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerMeta, err := utils.GetProviderMeta(meta)
	if err != nil {
		return diag.FromErr(err)
	}
	client := providerMeta.CloudAPI

	regionID := d.Id()

	// Fetch the list of providers to confirm the region still exists
	providers, err := client.ListCloudProviders(ctx)
	if err != nil {
		return diag.FromErr(err)
	}

	var targetProvider *clients.CloudProvider
	for _, provider := range providers {
		if provider.ID == regionID {
			targetProvider = &provider
			break
		}
	}

	// If the region is not found among providers, it is considered removed or disabled
	if targetProvider == nil {
		log.Printf("[WARN] Region %s not found, removing from state", regionID)
		d.SetId("")
		return nil
	}

	// Fetch the region details
	region, err := client.GetRegionDetails(ctx, *targetProvider)
	if err != nil {
		if strings.Contains(err.Error(), "non-200 status code: 204") {
			log.Printf("[WARN] Region %s details not available, removing from state", regionID)
			d.SetId("")
			return nil
		}
		return diag.FromErr(fmt.Errorf("error fetching region %s details: %s", regionID, err))
	}

	// Update the Terraform state with the fetched region details
	if region.RegionInfo != nil {
		d.Set("region_id", regionID)
		d.Set("sql_address", region.RegionInfo.SqlAddress)
		d.Set("http_address", region.RegionInfo.HttpAddress)
		d.Set("resolvable", region.RegionInfo.Resolvable)
		d.Set("enabled_at", region.RegionInfo.EnabledAt)
		d.Set("region_state", region.RegionInfo.Resolvable)
	}

	return nil
}

func resourceCloudRegionUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	log.Printf("[WARN] Update operation is not supported for regions and will perform no action")

	return resourceCloudRegionRead(ctx, d, meta)
}

func resourceCloudRegionDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	log.Printf("[WARN] Delete operation is not supported for regions and will perform no action")

	d.SetId("")
	return nil
}

func waitForRegionToBeEnabled(ctx context.Context, client *clients.CloudAPIClient, provider clients.CloudProvider) error {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	timeout := time.After(10 * time.Minute)

	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("operation was canceled")
		case <-timeout:
			return fmt.Errorf("timeout while waiting for region to be enabled")
		case <-ticker.C:
			region, err := client.GetRegionDetails(ctx, provider)
			if err != nil {
				log.Printf("Waiting for region to be enabled, current status: %v", err)
			} else if region.RegionInfo != nil && region.RegionInfo.EnabledAt != "" && region.RegionInfo.Resolvable {
				return nil
			}
		}
	}
}
