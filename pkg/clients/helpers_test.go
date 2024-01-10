package clients

import (
	"strings"
	"testing"
)

// TestConstructAppPassword tests the ConstructAppPassword function.
func TestConstructAppPassword(t *testing.T) {
	tests := []struct {
		name       string
		clientID   string
		secret     string
		wantPrefix string
	}{
		{
			name:       "normal IDs without dashes",
			clientID:   "1b2a3c",
			secret:     "4d5e6f",
			wantPrefix: "mzp_1b2a3c4d5e6f",
		},
		{
			name:       "IDs with dashes",
			clientID:   "1b2a-3c4d-5e6f",
			secret:     "7a8b-9c0d-1e2f",
			wantPrefix: "mzp_1b2a3c4d5e6f7a8b9c0d1e2f",
		},
		{
			name:       "long IDs",
			clientID:   "1b2a3c4d5e6f7a8b9c0d1e2f3a4b5c6d",
			secret:     "7e8f9a0b1c2d3e4f5a6b7c8d9e0f1a2b",
			wantPrefix: "mzp_1b2a3c4d5e6f7a8b9c0d1e2f3a4b5c6d7e8f9a0b1c2d3e4f5a6b7c8d9e0f1a2b",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ConstructAppPassword(tt.clientID, tt.secret)
			if !strings.HasPrefix(got, tt.wantPrefix) {
				t.Errorf("ConstructAppPassword() = %v, want prefix %v", got, tt.wantPrefix)
			}
		})
	}
}
