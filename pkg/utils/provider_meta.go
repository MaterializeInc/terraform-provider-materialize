package utils

import (
	"fmt"
	"strings"

	"github.com/MaterializeInc/terraform-provider-materialize/pkg/clients"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/jmoiron/sqlx"
)

// ProviderMeta holds essential configuration and client information
// required across various parts of the provider. It acts as a central
// repository of shared data, particularly for database connections, API clients,
// and regional settings.
type ProviderMeta struct {
	// DB is a map that associates each supported region with its corresponding
	// database client. This allows for region-specific database operations.
	DB map[clients.Region]*clients.DBClient

	// Frontegg represents the client used to interact with the Frontegg API,
	// which may involve authentication, token management, etc.
	Frontegg *clients.FronteggClient

	// CloudAPI is the client used for interactions with the cloud API
	CloudAPI *clients.CloudAPIClient

	// DefaultRegion specifies the default region to be used when no specific
	// region is provided in the resources and data sources.
	DefaultRegion clients.Region

	// RegionsEnabled is a map indicating which regions are currently enabled
	// for use. This can be used to quickly check the availability in different regions.
	RegionsEnabled map[clients.Region]bool
}

var DefaultRegion string

func GetProviderMeta(meta interface{}) (*ProviderMeta, error) {
	providerMeta, ok := meta.(*ProviderMeta)
	if !ok || providerMeta == nil {
		return nil, fmt.Errorf("type assertion failed: provider meta is not of type *ProviderMeta or is nil")
	}

	if providerMeta.Frontegg.NeedsTokenRefresh() {
		err := providerMeta.Frontegg.RefreshToken()
		if err != nil {
			return nil, fmt.Errorf("failed to refresh token: %v", err)
		}
	}

	return providerMeta, nil
}

func GetDBClientFromMeta(meta interface{}, d *schema.ResourceData) (*sqlx.DB, clients.Region, error) {
	providerMeta, err := GetProviderMeta(meta)
	if err != nil {
		return nil, "", err
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
