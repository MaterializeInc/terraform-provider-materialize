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

	// Frontegg Roles is a map that associates each Frontegg role with its corresponding ID.
	// This is used to map role names to role IDs when creating/updating users.
	FronteggRoles map[string]string
}

var DefaultRegion string

func GetProviderMeta(meta interface{}) (*ProviderMeta, error) {
	providerMeta := meta.(*ProviderMeta)

	if err := providerMeta.Frontegg.NeedsTokenRefresh(); err != nil {
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
	} else if d != nil && ExtractRegion(d.Id()) != "" {
		region = clients.Region(ExtractRegion(d.Id()))
	} else {
		region = providerMeta.DefaultRegion
	}

	if d != nil {
		d.Set("region", string(region))
	}

	// Check if the region is enabled using the stored information
	enabled, exists := providerMeta.RegionsEnabled[region]
	if !exists {
		var regions []string
		for regionKey := range providerMeta.RegionsEnabled {
			regions = append(regions, string(regionKey))
		}
		enabledRegions := strings.Join(regions, ", ")
		return nil, region, fmt.Errorf("region not found: '%s'. Currently enabled regions: %s", region, enabledRegions)
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

// Helper function to prepend region and type to the ID
func TransformIdWithTypeAndRegion(region string, idType string, value string) string {
	return fmt.Sprintf("%s:%s:%s", region, idType, value)
}

// Function to get the ID from the region + ID string
func ExtractId(fullId string) string {
	parts := strings.Split(fullId, ":")
	if len(parts) == 3 {
		// Format: region:idType:value
		return parts[2]
	} else if len(parts) == 2 {
		// Format: region:id
		return parts[1]
	}
	// Return original if not in expected format
	return fullId
}

// Function to get the region from the region + ID string
func ExtractRegion(oldID string) string {
	parts := strings.Split(oldID, ":")
	if len(parts) < 2 {
		// Return an empty string if the ID doesn't have a region
		return ""
	}
	return parts[0]
}

func ExtractIdType(fullId string) string {
	parts := strings.Split(fullId, ":")
	if len(parts) == 3 {
		return parts[1]
	}
	return "id"
}

// Function to extract the prefix and value from a prefixed ID
func ExtractPrefixedId(id string) (string, string, bool, error) {
	parts := strings.SplitN(id, ":", 2)
	if len(parts) != 2 {
		return "", "", false, fmt.Errorf("invalid ID format: %s", id)
	}

	prefix := strings.ToUpper(parts[0])
	value := parts[1]

	switch prefix {
	case "ID":
		return prefix, value, false, nil
	case "NAME":
		return prefix, value, true, nil
	default:
		return "", "", false, fmt.Errorf("invalid ID prefix: %s, allowed prefixes: ID, NAME", prefix)
	}
}
