package utils

import (
	"fmt"
	"strings"

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

var DefaultRegion string

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

func GetDBClientFromMeta(meta interface{}, d *schema.ResourceData) (*sqlx.DB, clients.Region, error) {
	providerMeta, ok := GetProviderMeta(meta)
	if !ok {
		return nil, "", fmt.Errorf("failed to get provider meta: %v", providerMeta)
	}

	// Determine the region to use, if one is not specified, use the default region
	var region clients.Region
	if d != nil && d.Get("region") != "" {
		region = clients.Region(d.Get("region").(string))
	} else {
		region = providerMeta.DefaultRegion
	}

	// Check if the region is enabled using the stored information
	enabled, exists := providerMeta.RegionsEnabled[region]
	if !exists {
		return nil, region, fmt.Errorf("no information available for region: %s", region)
	}

	if !enabled {
		return nil, region, fmt.Errorf("region '%s' is not enabled", region)
	}

	// Retrieve the appropriate DBClient for the region from the map
	dbClient, exists := providerMeta.DB[region]
	if !exists {
		return nil, region, fmt.Errorf("no database client for region: %s", region)
	}

	return dbClient.SQLX(), region, nil
}

func SetDefaultRegion(region string) error {
	DefaultRegion = region
	return nil
}

// Helper function to prepend region to the ID
func TransformIdWithRegion(region string, oldID string) string {
	// If the ID already has a region, return the original ID
	if strings.Contains(oldID, ":") {
		return oldID
	}
	return fmt.Sprintf("%s:%s", region, oldID)
}

// Function to get the ID from the region + ID string
func ExtractId(oldID string) string {
	parts := strings.Split(oldID, ":")
	if len(parts) < 2 {
		// Return the original ID if it doesn't have a region
		return oldID
	}
	return parts[1]
}
