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
		URL:           "http://cloud:3001",
		RegionInfo: &RegionInfo{
			SqlAddress:  "materialized:6877",
			HttpAddress: "materialized:6875",
			Resolvable:  true,
			EnabledAt:   "2023-01-01T00:00:00Z",
		},
	},
	// Add more mock regions if needed later
}

func main() {
	http.HandleFunc("/api/region", regionHandler)
	http.HandleFunc("/api/cloud-regions", cloudRegionsHandler)

	fmt.Println("Mock Cloud API server is running at http://localhost:3001")
	log.Fatal(http.ListenAndServe(":3001", nil))
}

func regionHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		mockRegion := CloudRegion{
			RegionInfo: &RegionInfo{
				SqlAddress:  "materialized:6877",
				HttpAddress: "materialized:6875",
				Resolvable:  true,
				EnabledAt:   "2023-01-01T00:00:00Z",
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(mockRegion)
	case http.MethodPatch:
		enabledRegion := CloudRegion{RegionInfo: nil}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(enabledRegion)
	case http.MethodDelete:
		w.WriteHeader(http.StatusAccepted)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
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
	json.NewEncoder(w).Encode(response)
}
