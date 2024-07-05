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
				Verified: true,
			}
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(userResponse)
			return
		}
		w.WriteHeader(http.StatusMethodNotAllowed)
	})

	// Endpoint for fetching a user
	handler.HandleFunc("/identity/resources/users/v1/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" {
			userID := strings.TrimPrefix(r.URL.Path, "/identity/resources/users/v1/")
			if userID == "test-user-id" {
				userResponse := UserResponse{
					ID:       userID,
					Email:    "test@example.com",
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

	handler.HandleFunc("/identity/resources/users/v3", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" {
			email := r.URL.Query().Get("_email")
			if email == "test@example.com" {
				response := struct {
					Items    []UserResponse `json:"items"`
					Metadata struct {
						TotalItems int `json:"totalItems"`
					} `json:"_metadata"`
				}{
					Items: []UserResponse{
						{
							ID:       "test-user-id",
							Email:    "test@example.com",
							Verified: true,
							Provider: "email",
						},
					},
					Metadata: struct {
						TotalItems int `json:"totalItems"`
					}{
						TotalItems: 1,
					},
				}
				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(response)
				return
			}
			// No user found
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(struct {
				Items    []UserResponse `json:"items"`
				Metadata struct {
					TotalItems int `json:"totalItems"`
				} `json:"_metadata"`
			}{
				Items: []UserResponse{},
				Metadata: struct {
					TotalItems int `json:"totalItems"`
				}{
					TotalItems: 0,
				},
			})
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

func TestGetUsers(t *testing.T) {
	mockServer := setupUserMockServer()
	defer mockServer.Close()

	client := &clients.FronteggClient{
		HTTPClient: &http.Client{},
		Endpoint:   mockServer.URL,
		Token:      "mock-token",
	}

	// Test case 1: User found
	params := QueryUsersParams{
		Email: "test@example.com",
		Limit: 1,
	}
	users, err := GetUsers(context.Background(), client, params)
	if err != nil {
		t.Fatalf("GetUsers returned an error: %v", err)
	}
	if len(users) != 1 {
		t.Fatalf("Expected 1 user, got %d", len(users))
	}
	if users[0].Email != "test@example.com" {
		t.Errorf("Expected email test@example.com, got %s", users[0].Email)
	}

	// Test case 2: User not found
	params.Email = "nonexistent@example.com"
	_, err = GetUsers(context.Background(), client, params)
	if err == nil {
		t.Fatalf("Expected an error for non-existent user, got nil")
	}
	if !strings.Contains(err.Error(), "no user found with email") {
		t.Errorf("Expected 'no user found' error, got: %v", err)
	}
}
