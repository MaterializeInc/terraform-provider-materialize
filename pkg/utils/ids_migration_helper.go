package utils

import (
	"context"
	"fmt"
	"strings"
)

var Host string

func ExtractRegionFromHostname() string {
	defaultRegion := "aws/us-east-1"
	if Host == "localhost" || Host == "materialize" || Host == "materialized" || Host == "127.0.0.1" {
		return defaultRegion
	}

	parts := strings.Split(Host, ".")
	if len(parts) < 3 {
		return defaultRegion
	}

	region := fmt.Sprintf("aws/%s", parts[1])
	return region
}

// Helper function to prepend region to the ID
func TransformIdWithRegion(oldID string) (string, error) {
	region := ExtractRegionFromHostname()
	if region == "" {
		return "", fmt.Errorf("failed to extract region from hostname")
	}
	// If the ID already has a region, return the original ID
	if strings.Contains(oldID, ":") {
		return oldID, nil
	}
	return fmt.Sprintf("%s:%s", region, oldID), nil
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
