package utils

import (
	"context"
	"fmt"
	"strings"
)

var Region string

func SetRegionFromHostname(host string) error {
	defaultRegion := "aws/us-east-1"
	if host == "localhost" || host == "materialize" || host == "materialized" || host == "127.0.0.1" {
		Region = defaultRegion
		return nil
	}

	parts := strings.Split(host, ".")
	if len(parts) < 3 {
		Region = defaultRegion
		return nil
	}

	Region = fmt.Sprintf("aws/%s", parts[1])
	return nil
}

// Helper function to prepend region to the ID
func TransformIdWithRegion(oldID string) (string, error) {
	if Region == "" {
		return "", fmt.Errorf("failed to extract region from hostname")
	}
	// If the ID already has a region, return the original ID
	if strings.Contains(oldID, ":") {
		return oldID, nil
	}
	return fmt.Sprintf("%s:%s", Region, oldID), nil
}

// Function to get the ID from the region + ID string
func ExtractId(oldID string) (string, error) {
	parts := strings.Split(oldID, ":")
	if len(parts) < 2 {
		// Return the original ID if it doesn't have a region
		return oldID, nil
	}
	return parts[1], nil
}

func IdStateUpgradeV0(ctx context.Context, rawState map[string]interface{}, meta interface{}) (map[string]interface{}, error) {
	oldID, ok := rawState["id"].(string)
	if !ok {
		return nil, fmt.Errorf("unexpected type for ID")
	}

	newID, err := TransformIdWithRegion(oldID)
	if err != nil {
		return nil, err
	}
	rawState["id"] = newID

	return rawState, nil
}
