package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
)

type RegionConfig struct {
	ID        string
	Name      string
	Hostname  string
	SqlPort   string
	HttpPort  string
	EnabledAt time.Time
}

type Config struct {
	CloudHostname string
	CloudPort     string
	Regions       []RegionConfig
}

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

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}

func loadConfig() Config {
	return Config{
		CloudHostname: getEnv("CLOUD_HOSTNAME", "localhost"),
		CloudPort:     getEnv("CLOUD_PORT", "3001"),
		Regions: []RegionConfig{
			{
				ID:        "aws/us-east-1",
				Name:      "us-east-1",
				Hostname:  getEnv("US_EAST_1_HOSTNAME", "materialized"),
				SqlPort:   getEnv("US_EAST_1_SQL_PORT", "6877"),
				HttpPort:  getEnv("US_EAST_1_HTTP_PORT", "6875"),
				EnabledAt: time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
			},
			{
				ID:        "aws/us-west-2",
				Name:      "us-west-2",
				Hostname:  getEnv("US_WEST_2_HOSTNAME", "materialized2"),
				SqlPort:   getEnv("US_WEST_2_SQL_PORT", "7877"),
				HttpPort:  getEnv("US_WEST_2_HTTP_PORT", "7875"),
				EnabledAt: time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
			},
		},
	}
}

func createRegions(config Config) []Region {
	regions := make([]Region, len(config.Regions))
	for i, r := range config.Regions {
		regions[i] = Region{
			ID:            r.ID,
			Name:          r.Name,
			CloudProvider: "aws",
			URL:           fmt.Sprintf("http://%s:%s/%s", config.CloudHostname, config.CloudPort, r.Name),
			RegionInfo: &RegionInfo{
				SqlAddress:  fmt.Sprintf("%s:%s", r.Hostname, r.SqlPort),
				HttpAddress: fmt.Sprintf("%s:%s", r.Hostname, r.HttpPort),
				Resolvable:  true,
				EnabledAt:   r.EnabledAt.Format(time.RFC3339),
			},
		}
	}
	return regions
}

func main() {
	config := loadConfig()
	regions := createRegions(config)

	http.HandleFunc("/api/cloud-regions", cloudRegionsHandler(regions))
	for _, region := range regions {
		http.HandleFunc(fmt.Sprintf("/%s/api/region", region.Name), regionHandler(region.ID, regions))
	}

	fmt.Printf("Mock Cloud API server is running at http://%s:%s\n", config.CloudHostname, config.CloudPort)
	log.Fatal(http.ListenAndServe(":"+config.CloudPort, nil))
}

func regionHandler(regionID string, regions []Region) http.HandlerFunc {
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

func cloudRegionsHandler(regions []Region) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
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
}
