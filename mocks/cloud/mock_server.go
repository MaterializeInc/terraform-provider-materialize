package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

type Region struct {
	ID            string      `json:"id"`
	Name          string      `json:"name"`
	CloudProvider string      `json:"cloudProvider"`
	URL           string      `json:"url"`
	RegionInfo    *RegionInfo `json:"regionInfo,omitempty"`
}

type RegionInfo struct {
	SqlAddress  string `json:"sqlAddress"`
	HttpAddress string `json:"httpAddress"`
	Resolvable  bool   `json:"resolvable"`
	EnabledAt   string `json:"enabledAt"`
}

type CloudRegion struct {
	RegionInfo *RegionInfo `json:"regionInfo"`
}

type CloudProviderResponse struct {
	Data       []Region `json:"data"`
	NextCursor string   `json:"nextCursor,omitempty"`
}

// Mock data
var regions = []Region{
	{
		ID:            "aws/us-east-1",
		Name:          "us-east-1",
		CloudProvider: "aws",
		URL:           "http://cloud:3001/us-east-1",
		RegionInfo: &RegionInfo{
			SqlAddress:  "materialized:6877",
			HttpAddress: "materialized:6875",
			Resolvable:  true,
			EnabledAt:   "2023-01-01T00:00:00Z",
		},
	},
	{
		ID:            "aws/us-west-2",
		Name:          "us-west-2",
		CloudProvider: "aws",
		URL:           "http://cloud:3001/us-west-2",
		RegionInfo: &RegionInfo{
			SqlAddress:  "materialized2:7877",
			HttpAddress: "materialized2:7875",
			Resolvable:  true,
			EnabledAt:   "2023-01-01T00:00:00Z",
		},
	},
}

func main() {
	// Cloud Global API endpoint
	http.HandleFunc("/api/cloud-regions", cloudRegionsHandler)

	// Cloud Region API endpoints
	http.HandleFunc("/us-east-1/api/region", regionHandler("aws/us-east-1"))
	http.HandleFunc("/us-west-2/api/region", regionHandler("aws/us-west-2"))

	fmt.Println("Mock Cloud API server is running at http://localhost:3001")
	log.Fatal(http.ListenAndServe(":3001", nil))
}

func regionHandler(regionID string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Received %s request to %s", r.Method, r.URL.Path)

		var selectedRegion *Region
		for _, region := range regions {
			if region.ID == regionID {
				selectedRegion = &region
				break
			}
		}

		if selectedRegion == nil {
			http.Error(w, "Region not found", http.StatusNotFound)
			return
		}

		switch r.Method {
		case http.MethodGet:
			mockRegion := CloudRegion{
				RegionInfo: selectedRegion.RegionInfo,
			}
			w.Header().Set("Content-Type", "application/json")
			if err := json.NewEncoder(w).Encode(mockRegion); err != nil {
				http.Error(w, "Failed to encode response", http.StatusInternalServerError)
			}

		case http.MethodPatch:
			var updatedRegion RegionInfo
			if err := json.NewDecoder(r.Body).Decode(&updatedRegion); err != nil {
				http.Error(w, "Invalid request body", http.StatusBadRequest)
				return
			}
			selectedRegion.RegionInfo = &updatedRegion
			w.WriteHeader(http.StatusOK)

		case http.MethodDelete:
			selectedRegion.RegionInfo = nil
			w.WriteHeader(http.StatusAccepted)

		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}
}

func cloudRegionsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	response := CloudProviderResponse{
		Data: regions,
	}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}
