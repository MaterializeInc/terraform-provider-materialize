package utils

import (
	"fmt"

	"github.com/MaterializeInc/terraform-provider-materialize/pkg/clients"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/jmoiron/sqlx"
)

type ProviderMeta struct {
	DB             map[clients.Region]*clients.DBClient
	Frontegg       *clients.FronteggClient
	CloudAPI       *clients.CloudAPIClient
	DefaultRegion  clients.Region
	RegionsEnabled map[clients.Region]bool
}

func GetProviderMeta(meta interface{}) (*ProviderMeta, bool) {
	providerMeta, ok := meta.(*ProviderMeta)
	if !ok || providerMeta == nil {
		fmt.Println("Type assertion failed: provider meta is not of type *ProviderMeta or is nil")
		return nil, false
	}
	if providerMeta.Frontegg.NeedsTokenRefresh() {
		err := providerMeta.Frontegg.RefreshToken()
		if err != nil {
			fmt.Printf("Failed to refresh token: %v\n", err)
			return nil, false
		}
	}

	return providerMeta, true
}

func GetDBClientFromMeta(meta interface{}, d *schema.ResourceData) (*sqlx.DB, error) {
	providerMeta, ok := GetProviderMeta(meta)
	if !ok {
		return nil, fmt.Errorf("failed to get provider meta: %v", providerMeta)
	}

	// Determine the region to use
	var region clients.Region
	if d != nil && d.Get("region") != "" {
		region = clients.Region(d.Get("region").(string))
	} else {
		region = providerMeta.DefaultRegion
	}

	// Check if the region is enabled using the stored information
	enabled, exists := providerMeta.RegionsEnabled[region]
	if !exists {
		return nil, fmt.Errorf("no information available for region: %s", region)
	}

	if !enabled {
		return nil, fmt.Errorf("region '%s' is not enabled", region)
	}

	// Retrieve the appropriate DBClient for the region from the map
	dbClient, exists := providerMeta.DB[region]
	if !exists {
		return nil, fmt.Errorf("no database client for region: %s", region)
	}

	return dbClient.SQLX(), nil
}
