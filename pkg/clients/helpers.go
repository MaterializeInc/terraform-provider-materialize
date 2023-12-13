package clients

import (
	"fmt"
	"strings"
	"time"
)

type AppPassword struct {
	ClientID    string    `json:"clientId"`
	Secret      string    `json:"secret"`
	Description string    `json:"description"`
	Owner       string    `json:"owner"`
	CreatedAt   time.Time `json:"created_at"`
}

type Region string

// Role represents the Frontegg role structure.
type Role struct {
	ID            string    `json:"id"`
	VendorID      string    `json:"vendorId"`
	TenantID      *string   `json:"tenantId,omitempty"`
	Key           string    `json:"key"`
	Name          string    `json:"name"`
	Description   string    `json:"description"`
	IsDefault     bool      `json:"isDefault"`
	FirstUserRole bool      `json:"firstUserRole"`
	CreatedAt     time.Time `json:"createdAt"`
	UpdatedAt     time.Time `json:"updatedAt"`
	Permissions   []string  `json:"permissions"`
	Level         int       `json:"level"`
}

// Helper function to construct app password from clientId and secret.
func ConstructAppPassword(clientID, secret string) string {
	// Remove dashes and concatenate with "mzp_" prefix.
	clientIDClean := strings.ReplaceAll(clientID, "-", "")
	secretClean := strings.ReplaceAll(secret, "-", "")
	return fmt.Sprintf("mzp_%s%s", clientIDClean, secretClean)
}

const (
	AwsUsEast1 Region = "aws/us-east-1"
	AwsUsWest2 Region = "aws/us-west-2"
	AwsEuWest1 Region = "aws/eu-west-1"
)
