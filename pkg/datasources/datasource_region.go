package datasources

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/MaterializeInc/terraform-provider-materialize/pkg/clients"
	"github.com/MaterializeInc/terraform-provider-materialize/pkg/utils"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func Region() *schema.Resource {
	return &schema.Resource{
		ReadContext: RegionRead,
		Schema: map[string]*schema.Schema{
			"regions": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The name of the region.",
						},
						"url": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The URL at which the Region API can be reached.",
						},
						"cloud_provider": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The cloud provider of the region. Currently, only AWS is supported.",
						},
						"host": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The SQL host of the region. This is the hostname of the Materialize cluster in the region.",
						},
					},
				},
			},
		},
	}
}

func RegionRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerMeta, err := utils.GetProviderMeta(meta)
	if err != nil {
		return diag.FromErr(err)
	}
	client := providerMeta.CloudAPI

	providers, err := client.ListCloudProviders(ctx)
	if err != nil {
		return diag.FromErr(err)
	}

	var regions []map[string]interface{}
	for _, provider := range providers {
		host, err := client.GetHost(ctx, provider.ID)
		if err != nil {
			if strings.Contains(err.Error(), "non-200 status code: 204") {
				log.Printf("[WARN] No host available for region %s, skipping", provider.ID)
				continue
			}
			return diag.FromErr(fmt.Errorf("error fetching host for region %s: %s", provider.ID, err))
		}

		region := createRegionMap(provider, host)
		regions = append(regions, region)
	}

	if err := d.Set("regions", regions); err != nil {
		return diag.FromErr(err)
	}

	d.SetId("regions")
	return nil
}

// createRegionMap creates a map of region details
func createRegionMap(provider clients.CloudProvider, host string) map[string]interface{} {
	return map[string]interface{}{
		"id":             provider.ID,
		"name":           provider.Name,
		"url":            provider.Url,
		"cloud_provider": provider.CloudProvider,
		"host":           host,
	}
}
