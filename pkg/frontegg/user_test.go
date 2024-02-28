package frontegg

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/MaterializeInc/terraform-provider-materialize/pkg/clients"
)

func setupUserMockServer() *httptest.Server {
	handler := http.NewServeMux()

	handler.HandleFunc("/identity/resources/users/v2", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {
			var userRequest UserRequest
			if err := json.NewDecoder(r.Body).Decode(&userRequest); err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			userResponse := UserResponse{
				ID:       "test-user-id",
				Email:    userRequest.Email,
				Roles:    userRequest.RoleIDs,
				Verified: true,
			}
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(userResponse)
			return
		}
		w.WriteHeader(http.StatusMethodNotAllowed)
	})

	// Endpoint for fetching a user
	handler.HandleFunc("/identity/resources/users/v2/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" {
			userID := strings.TrimPrefix(r.URL.Path, "/identity/resources/users/v2/")
			if userID == "test-user-id" {
				userResponse := UserResponse{
					ID:       userID,
					Email:    "test@example.com",
					Roles:    []string{"role1", "role2"},
					Verified: true,
				}
				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(userResponse)
				return
			}
			w.WriteHeader(http.StatusNotFound)
			return
		}
		if r.Method == "DELETE" {
			w.WriteHeader(http.StatusOK)
			return
		}
		w.WriteHeader(http.StatusMethodNotAllowed)
	})

	return httptest.NewServer(handler)
}

func TestCreateUser(t *testing.T) {
	mockServer := setupUserMockServer()
	defer mockServer.Close()

	client := &clients.FronteggClient{
		HTTPClient: &http.Client{},
		Endpoint:   mockServer.URL,
		Token:      "mock-token",
	}

	userRequest := UserRequest{
		Email:   "test@example.com",
		RoleIDs: []string{"role1", "role2"},
	}
	userResponse, err := CreateUser(context.Background(), client, userRequest)
	if err != nil {
		t.Fatalf("CreateUser returned an error: %v", err)
	}

	if userResponse.Email != userRequest.Email {
		t.Errorf("Expected email %s, got %s", userRequest.Email, userResponse.Email)
	}
}

func TestReadUser(t *testing.T) {
	mockServer := setupUserMockServer()
	defer mockServer.Close()

	client := &clients.FronteggClient{
		HTTPClient: &http.Client{},
		Endpoint:   mockServer.URL,
		Token:      "mock-token",
	}

	userID := "test-user-id"
	userResponse, err := ReadUser(context.Background(), client, userID)
	if err != nil {
		t.Fatalf("ReadUser returned an error: %v", err)
	}

	if userResponse.ID != userID {
		t.Errorf("Expected user ID %s, got %s", userID, userResponse.ID)
	}
}

func TestDeleteUser(t *testing.T) {
	mockServer := setupUserMockServer()
	defer mockServer.Close()

	client := &clients.FronteggClient{
		HTTPClient: &http.Client{},
		Endpoint:   mockServer.URL,
		Token:      "mock-token",
	}

	userID := "test-user-id"
	err := DeleteUser(context.Background(), client, userID)
	if err != nil {
		t.Fatalf("DeleteUser returned an error: %v", err)
	}
}
