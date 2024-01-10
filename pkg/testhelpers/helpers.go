package testhelpers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
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

	// Create a mock HTTP server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		switch req.URL.Path {
		case "/identity/resources/users/api-tokens/v1":
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
				w.WriteHeader(http.StatusNoContent)
			default:
				http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
			}
		case "/identity/resources/users/v1/mock-user-id":
			switch req.Method {
			case http.MethodGet:
				if strings.HasSuffix(req.URL.Path, "/mock-user-id") {
					mockUser := struct {
						ID                string `json:"id"`
						Email             string `json:"email"`
						ProfilePictureURL string `json:"profilePictureUrl"`
						Verified          bool   `json:"verified"`
						Metadata          string `json:"metadata"`
					}{
						ID:                "mock-user-id",
						Email:             "test@example.com",
						ProfilePictureURL: "http://example.com/picture.jpg",
						Verified:          true,
						Metadata:          "{}",
					}
					json.NewEncoder(w).Encode(mockUser)
				} else {
					w.WriteHeader(http.StatusNotFound)
				}
			case http.MethodDelete:
				if strings.HasSuffix(req.URL.Path, "/mock-user-id") {
					w.WriteHeader(http.StatusOK)
				} else {
					w.WriteHeader(http.StatusNotFound)
				}
			}

		default:
			http.Error(w, "Not Found", http.StatusNotFound)
		}
	}))
	defer server.Close()

	f(server.URL)
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
