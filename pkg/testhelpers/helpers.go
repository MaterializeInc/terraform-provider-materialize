package testhelpers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/MaterializeInc/terraform-provider-materialize/pkg/clients"
	"github.com/MaterializeInc/terraform-provider-materialize/pkg/utils"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/require"
)

type MockAppPassword struct {
	ClientID    string    `json:"clientId"`
	Description string    `json:"description"`
	Owner       string    `json:"owner"`
	CreatedAt   time.Time `json:"created_at"`
	Secret      string    `json:"secret"`
}

type FronteggRole struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type FronteggRolesResponse struct {
	Items    []FronteggRole `json:"items"`
	Metadata struct {
		TotalItems int `json:"totalItems"`
		TotalPages int `json:"totalPages"`
	} `json:"_metadata"`
}

type Domain struct {
	ID          string `json:"id"`
	Domain      string `json:"domain"`
	Validated   bool   `json:"validated"`
	SsoConfigId string `json:"sso_config_id"`
}

type GroupMapping struct {
	ID          string   `json:"id"`
	Group       string   `json:"group"`
	RoleIds     []string `json:"roleIds"`
	SsoConfigId string   `json:"ssoConfigId"`
	Enabled     bool     `json:"enabled"`
}

type DefaultRoles struct {
	RoleIds []string `json:"roleIds"`
}

type SSOConfig struct {
	Id                        string         `json:"id"`
	Enabled                   bool           `json:"enabled"`
	SsoEndpoint               string         `json:"ssoEndpoint"`
	PublicCertificate         string         `json:"publicCertificate"`
	SignRequest               bool           `json:"signRequest"`
	AcsUrl                    string         `json:"acsUrl"`
	SpEntityId                string         `json:"spEntityId"`
	Type                      string         `json:"type"`
	OidcClientId              string         `json:"oidcClientId"`
	OidcSecret                string         `json:"oidcSecret"`
	Domains                   []Domain       `json:"domains"`
	Groups                    []GroupMapping `json:"groups"`
	DefaultRoles              DefaultRoles   `json:"defaultRoles"`
	GeneratedVerification     string         `json:"generatedVerification,omitempty"`
	CreatedAt                 time.Time      `json:"createdAt,omitempty"`
	UpdatedAt                 time.Time      `json:"updatedAt,omitempty"`
	ConfigMetadata            interface{}    `json:"configMetadata,omitempty"`
	OverrideActiveTenant      bool           `json:"overrideActiveTenant,omitempty"`
	SkipEmailDomainValidation bool           `json:"skipEmailDomainValidation,omitempty"`
	SubAccountAccessLimit     int            `json:"subAccountAccessLimit,omitempty"`
	RoleIds                   []string       `json:"roleIds"`
}

func WithMockDb(t *testing.T, f func(*sqlx.DB, sqlmock.Sqlmock)) {
	// Set the region for testing
	utils.DefaultRegion = "aws/us-east-1"

	t.Helper()
	r := require.New(t)
	db, mock, err := sqlmock.New()
	dbx := sqlx.NewDb(db, "sqlmock")
	r.NoError(err)
	defer dbx.Close()

	mock.MatchExpectationsInOrder(true)

	f(dbx, mock)
}

func WithMockProviderMeta(t *testing.T, f func(*utils.ProviderMeta, sqlmock.Sqlmock)) {
	t.Helper()
	r := require.New(t)
	db, mock, err := sqlmock.New()
	r.NoError(err)
	defer db.Close()

	dbx := sqlx.NewDb(db, "sqlmock")
	dbClients := make(map[clients.Region]*clients.DBClient)
	dbClients[clients.AwsUsEast1] = &clients.DBClient{DB: dbx}
	regionsEnabled := make(map[clients.Region]bool)
	regionsEnabled[clients.AwsUsEast1] = true

	providerMeta := &utils.ProviderMeta{
		DB:             dbClients,
		RegionsEnabled: regionsEnabled,
		DefaultRegion:  clients.AwsUsEast1,
		Frontegg: &clients.FronteggClient{
			TokenExpiry: time.Date(9999, 1, 1, 0, 0, 0, 0, time.UTC),
		},
		CloudAPI: nil,
	}

	mock.MatchExpectationsInOrder(true)

	f(providerMeta, mock)
}

func WithMockFronteggServer(t *testing.T, f func(url string)) {
	t.Helper()
	r := require.New(t)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		switch {
		case strings.HasPrefix(req.URL.Path, "/identity/resources/users/api-tokens/v1"):
			handleAppPasswords(w, req, r)
		case strings.HasPrefix(req.URL.Path, "/identity/resources/users/v1/"):
			handleUserRequests(w, req, r)
		case strings.HasPrefix(req.URL.Path, "/identity/resources/roles/v2"):
			handleRolesRequests(w, req, r)
		case req.URL.Path == "/frontegg/team/resources/sso/v1/configurations":
			switch req.Method {
			case http.MethodPost:
				handleCreateSSOConfig(w, req, r)
			case http.MethodGet:
				handleListSSOConfigs(w, req, r)
			default:
				http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
			}
		case strings.HasPrefix(req.URL.Path, "/frontegg/team/resources/sso/v1/configurations/"):
			handleSSOConfigAndDomainRequests(w, req)
		default:
			http.Error(w, "Not Found", http.StatusNotFound)
		}
	}))
	defer server.Close()

	f(server.URL)
}

func handleAppPasswords(w http.ResponseWriter, req *http.Request, r *require.Assertions) {
	switch req.Method {
	case http.MethodPost:
		var createReq struct {
			Description string `json:"description"`
		}
		err := json.NewDecoder(req.Body).Decode(&createReq)
		r.NoError(err)

		appPassword := MockAppPassword{
			ClientID:    "mock-client-id",
			Description: createReq.Description,
			Owner:       "mockOwner",
			CreatedAt:   time.Now(),
			Secret:      "mock-secret",
		}

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(appPassword)

	case http.MethodGet:
		mockAppPassword := MockAppPassword{
			ClientID:    "mock-client-id",
			Description: "test-app-password",
			Owner:       "mockOwner",
			CreatedAt:   time.Now(),
			Secret:      "mock-secret",
		}
		json.NewEncoder(w).Encode([]MockAppPassword{mockAppPassword})

	case http.MethodDelete:
		clientID := req.URL.Query().Get("clientId")
		if clientID != "" {
			w.WriteHeader(http.StatusOK)
		}

	default:
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
	}
}

func handleUserRequests(w http.ResponseWriter, req *http.Request, r *require.Assertions) {
	userID := strings.TrimPrefix(req.URL.Path, "/identity/resources/users/v1/")

	switch req.Method {
	case http.MethodGet:
		if userID != "" {
			// Mock response for a specific user
			mockUser := struct {
				ID                string `json:"id"`
				Email             string `json:"email"`
				ProfilePictureURL string `json:"profilePictureUrl"`
				Verified          bool   `json:"verified"`
				Metadata          string `json:"metadata"`
			}{
				ID:                userID,
				Email:             "test@example.com",
				ProfilePictureURL: "http://example.com/picture.jpg",
				Verified:          true,
				Metadata:          "{}",
			}
			json.NewEncoder(w).Encode(mockUser)
		} else {
			// Handle case where user ID is not provided
			http.Error(w, "User ID is required", http.StatusBadRequest)
		}

	case http.MethodPost:
		// Implement logic for creating a user
		var newUser struct {
			Email             string `json:"email"`
			ProfilePictureURL string `json:"profilePictureUrl"`
		}
		err := json.NewDecoder(req.Body).Decode(&newUser)
		r.NoError(err)

		// Create and return a mock user
		mockUser := struct {
			ID                string `json:"id"`
			Email             string `json:"email"`
			ProfilePictureURL string `json:"profilePictureUrl"`
			Verified          bool   `json:"verified"`
			Metadata          string `json:"metadata"`
		}{
			ID:                "new-mock-user-id",
			Email:             newUser.Email,
			ProfilePictureURL: newUser.ProfilePictureURL,
			Verified:          false,
			Metadata:          "{}",
		}
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(mockUser)

	case http.MethodDelete:
		if userID != "" {
			// Mock logic for deleting a user
			w.WriteHeader(http.StatusOK)
		} else {
			// Handle case where user ID is not provided
			http.Error(w, "User ID is required", http.StatusBadRequest)
		}

	default:
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
	}
}

func handleRolesRequests(w http.ResponseWriter, req *http.Request, r *require.Assertions) {
	if req.Method != http.MethodGet {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	// Mocked roles data
	mockRoles := []FronteggRole{
		{ID: "1", Name: "Organization Admin"},
		{ID: "2", Name: "Organization Member"},
	}

	// Mocked response for roles request
	response := FronteggRolesResponse{
		Items: mockRoles,
		Metadata: struct {
			TotalItems int `json:"totalItems"`
			TotalPages int `json:"totalPages"`
		}{
			TotalItems: len(mockRoles),
			TotalPages: 1,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func handleCreateSSOConfig(w http.ResponseWriter, req *http.Request, r *require.Assertions) {
	if req.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	var newConfig SSOConfig
	err := json.NewDecoder(req.Body).Decode(&newConfig)
	r.NoError(err)

	// Generating a mock ID for the new SSO configuration
	newConfig.Id = "mock-config-" + time.Now().Format("20060102150405")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(newConfig)
}

func handleListSSOConfigs(w http.ResponseWriter, req *http.Request, r *require.Assertions) {
	if req.Method != http.MethodGet {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	// Mocked SSO configurations
	mockConfigs := []SSOConfig{
		{
			Id:                "mock-config-1",
			Enabled:           true,
			SsoEndpoint:       "https://sso.example.com",
			PublicCertificate: "bW9jay1wdWJsaWMtY2VydGlmaWNhdGUK",
			SignRequest:       true,
			Type:              "SAML",
			OidcClientId:      "mock-oidc-client-id",
			OidcSecret:        "mock-oidc-secret",
			AcsUrl:            "https://acs.example.com/callback",
			SpEntityId:        "https://sp.example.com/metadata",
			Domains: []Domain{
				{
					ID:        "domain-1",
					Domain:    "example.com",
					Validated: true,
				},
			},
			Groups:       []GroupMapping{{ID: "group-1", Group: "admins", RoleIds: []string{"role-1"}}},
			DefaultRoles: DefaultRoles{RoleIds: []string{"role-1"}},
			RoleIds:      []string{"role-1"},
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(mockConfigs)
}

func handleSSOConfigAndDomainRequests(w http.ResponseWriter, req *http.Request) {
	parts := strings.Split(req.URL.Path, "/")
	if len(parts) < 8 {
		http.Error(w, "Invalid URL", http.StatusBadRequest)
		return
	}

	ssoConfigID := parts[7]

	if len(parts) >= 9 {
		resource := parts[8]

		switch resource {
		case "domains":
			handleDomainRequests(w, req, ssoConfigID, parts)
		case "groups":
			handleGroupMappingRequests(w, req, ssoConfigID, parts)
		case "roles":
			handleDefaultRolesRequests(w, req, ssoConfigID)
		default:
			http.Error(w, "Invalid request", http.StatusBadRequest)
		}
	} else {
		// Handle the PATCH request for updating an SSO configuration here.
		if req.Method == http.MethodPatch {
			// Parse the request body to get the updated SSO configuration data
			var updatedSSOConfig SSOConfig
			err := json.NewDecoder(req.Body).Decode(&updatedSSOConfig)
			if err != nil {
				http.Error(w, "Failed to decode request body", http.StatusBadRequest)
				return
			}
			w.WriteHeader(http.StatusOK)
		}
		if req.Method == http.MethodDelete {
			w.WriteHeader(http.StatusOK)
		} else {
			http.Error(w, "Invalid request", http.StatusBadRequest)
		}

	}
}

func handleDomainRequests(w http.ResponseWriter, req *http.Request, ssoConfigID string, parts []string) {
	domainID := ""
	if len(parts) > 9 {
		domainID = parts[9]
	}

	switch req.Method {
	case http.MethodPost:
		// Handle creating a new domain
		var newDomain Domain
		err := json.NewDecoder(req.Body).Decode(&newDomain)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		newDomain.ID = "mock-domain-id"
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(newDomain)

	case http.MethodGet:
		if domainID == "" {
			// Handle listing all domains for the SSO configuration
			mockDomains := []Domain{
				{
					ID:          "domain-1",
					Domain:      "example.com",
					Validated:   true,
					SsoConfigId: ssoConfigID,
				},
			}
			json.NewEncoder(w).Encode(mockDomains)
		}

	case http.MethodPatch:
		// Handle updating a specific domain
		var updatedDomain Domain
		err := json.NewDecoder(req.Body).Decode(&updatedDomain)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		updatedDomain.ID = domainID
		json.NewEncoder(w).Encode(updatedDomain)

	case http.MethodDelete:
		// Handle deleting a specific domain
		if domainID != "" {
			w.WriteHeader(http.StatusOK)
		} else {
			http.Error(w, "Domain ID is required", http.StatusBadRequest)
		}

	default:
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
	}
}

func handleGroupMappingRequests(w http.ResponseWriter, req *http.Request, ssoConfigID string, parts []string) {
	groupMappingID := ""
	if len(parts) > 9 {
		groupMappingID = parts[9]
	}

	switch req.Method {
	case http.MethodPost:
		// Handle creating a new group mapping
		var newGroupMapping GroupMapping
		err := json.NewDecoder(req.Body).Decode(&newGroupMapping)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		newGroupMapping.ID = "mock-groupmapping-id"
		newGroupMapping.SsoConfigId = ssoConfigID
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(newGroupMapping)

	case http.MethodGet:
		if groupMappingID == "" {
			// Handle listing all group mappings for the SSO configuration
			mockGroupMappings := []GroupMapping{
				{
					ID:          "groupmapping-1",
					Group:       "admins",
					RoleIds:     []string{"role-1"},
					SsoConfigId: ssoConfigID,
					Enabled:     true,
				},
			}
			json.NewEncoder(w).Encode(mockGroupMappings)
		} else {
			// Handle getting a specific group mapping
			mockGroupMapping := GroupMapping{
				ID:          groupMappingID,
				Group:       "specific-group",
				RoleIds:     []string{"role-2"},
				SsoConfigId: ssoConfigID,
				Enabled:     true,
			}
			json.NewEncoder(w).Encode(mockGroupMapping)
		}

	case http.MethodPatch:
		// Handle updating a specific group mapping
		var updatedGroupMapping GroupMapping
		err := json.NewDecoder(req.Body).Decode(&updatedGroupMapping)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		updatedGroupMapping.ID = groupMappingID
		json.NewEncoder(w).Encode(updatedGroupMapping)

	case http.MethodDelete:
		// Handle deleting a specific group mapping
		if groupMappingID != "" {
			w.WriteHeader(http.StatusOK)
		} else {
			http.Error(w, "Group Mapping ID is required", http.StatusBadRequest)
		}

	default:
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
	}
}

func handleDefaultRolesRequests(w http.ResponseWriter, req *http.Request, ssoConfigID string) {
	switch req.Method {
	case http.MethodPut:
		// Handle setting default roles
		var roles DefaultRoles
		err := json.NewDecoder(req.Body).Decode(&roles)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(roles)

	case http.MethodGet:
		// Handle getting default roles
		mockRoles := DefaultRoles{
			RoleIds: []string{"1", "2"},
		}
		log.Printf("mockRoles: %+v", mockRoles)
		json.NewEncoder(w).Encode(mockRoles)

	case http.MethodDelete:
		// Handle clearing default roles
		w.WriteHeader(http.StatusOK)

	default:
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
	}
}

// MockCloudService is a mock implementation of the http.RoundTripper interface for cloud-related requests
type MockCloudService struct{}

func (m *MockCloudService) RoundTrip(req *http.Request) (*http.Response, error) {
	// Check the requested URL and return a response accordingly
	if strings.HasSuffix(req.URL.Path, "/api/cloud-regions") {
		// Mock response data
		data := clients.CloudProviderResponse{
			Data: []clients.CloudProvider{
				{ID: "aws/us-east-1", Name: "us-east-1", Url: "http://mockendpoint", CloudProvider: "aws"},
				{ID: "aws/eu-west-1", Name: "eu-west-1", Url: "http://mockendpoint", CloudProvider: "aws"},
			},
		}

		// Convert response data to JSON
		respData, _ := json.Marshal(data)

		// Create a new HTTP response with the JSON data
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(bytes.NewReader(respData)),
			Header:     make(http.Header),
		}, nil
	} else if strings.HasSuffix(req.URL.Path, "/api/region") {
		// Return mock response for GetRegionDetails
		details := clients.CloudRegion{
			RegionInfo: &clients.RegionInfo{
				SqlAddress:  "sql.materialize.com",
				HttpAddress: "http.materialize.com",
				Resolvable:  true,
				EnabledAt:   "2021-01-01T00:00:00Z",
			},
		}
		respData, _ := json.Marshal(details)
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(bytes.NewReader(respData)),
			Header:     make(http.Header),
		}, nil
	}
	return nil, fmt.Errorf("no mock available for the requested endpoint")
}

// WithMockCloudServer sets up a mock HTTP server for cloud-related requests and calls the provided function with the server URL.
func WithMockCloudServer(t *testing.T, f func(url string)) {
	t.Helper()

	// Create a mock HTTP server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		// Use the MockCloudService for handling requests
		m := &MockCloudService{}
		resp, err := m.RoundTrip(req)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Copy the response to the server's response writer
		copyHeaders(w.Header(), resp.Header)
		w.WriteHeader(resp.StatusCode)
		_, _ = io.Copy(w, resp.Body)
	}))

	defer server.Close()

	f(server.URL)
}

// Helper function to copy headers from the response to the writer
func copyHeaders(dst, src http.Header) {
	for key, values := range src {
		for _, value := range values {
			dst.Add(key, value)
		}
	}
}
