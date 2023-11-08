package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"
)

type AppPassword struct {
	ClientID    string    `json:"clientId"`
	Secret      string    `json:"secret"`
	Description string    `json:"description"`
	Owner       string    `json:"owner"`
	CreatedAt   time.Time `json:"created_at"`
}

type User struct {
	ID                string `json:"id"`
	Email             string `json:"email"`
	ProfilePictureURL string `json:"profilePictureUrl"`
	Verified          bool   `json:"verified"`
	Metadata          string `json:"metadata"`
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

var (
	appPasswords = make(map[string]AppPassword)
	users        = make(map[string]User)
	mutex        = &sync.Mutex{}
)

func main() {
	http.HandleFunc("/identity/resources/auth/v1/api-token", handleTokenRequest)
	http.HandleFunc("/identity/resources/users/api-tokens/v1", handleAppPasswords)
	http.HandleFunc("/identity/resources/users/v1/", handleUserRequest)
	http.HandleFunc("/identity/resources/users/v2", handleUserRequest)
	http.HandleFunc("/identity/resources/roles/v2", handleRolesRequest)

	fmt.Println("Mock Frontegg server is running at http://localhost:3000")
	log.Fatal(http.ListenAndServe(":3000", nil))
}

func handleUserRequest(w http.ResponseWriter, r *http.Request) {
	logRequest(r)
	switch r.Method {
	case http.MethodGet:
		getUser(w, r)
	case http.MethodDelete:
		deleteUser(w, r)
	case http.MethodPost:
		createUser(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func getUser(w http.ResponseWriter, r *http.Request) {
	userID := strings.TrimPrefix(r.URL.Path, "/identity/resources/users/v1/")
	if userID == "" {
		http.Error(w, "User ID is required", http.StatusBadRequest)
		return
	}

	user, ok := users[userID]
	if !ok {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

func createUser(w http.ResponseWriter, r *http.Request) {
	var newUser User
	if err := json.NewDecoder(r.Body).Decode(&newUser); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	userID := generateUserID()
	newUser.ID = userID

	// Store the new user
	mutex.Lock()
	users[userID] = newUser
	mutex.Unlock()

	w.WriteHeader(http.StatusCreated)
	// Return the created user
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(newUser)
}

func generateUserID() string {
	return fmt.Sprintf("user-%d", time.Now().UnixNano())
}

func deleteUser(w http.ResponseWriter, r *http.Request) {
	userID := strings.TrimPrefix(r.URL.Path, "/identity/resources/users/v1/")
	if userID == "" {
		http.Error(w, "User ID is required", http.StatusBadRequest)
		return
	}

	_, ok := users[userID]
	if !ok {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	delete(users, userID)
	w.WriteHeader(http.StatusOK)
}

func handleTokenRequest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var payload struct {
		ClientId string `json:"clientId"`
		Secret   string `json:"secret"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if payload.ClientId == "1b2a3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d" && payload.Secret == "7e8f9a0b-1c2d-3e4f-5a6b-7c8d9e0f1a2b" {
		mockToken := createMockJWTToken()
		response := map[string]string{
			"accessToken": mockToken,
			"email":       "mz_system",
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	} else {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
	}
}

func createMockJWTToken() string {
	header := base64UrlEncode([]byte(`{"alg":"HS256","typ":"JWT"}`))
	payload := base64UrlEncode([]byte(`{"email":"mz_system","exp":1700000000}`))
	signature := base64UrlEncode([]byte(`signature`))
	return fmt.Sprintf("%s.%s.%s", header, payload, signature)
}

func base64UrlEncode(input []byte) string {
	encoded := base64.StdEncoding.EncodeToString(input)
	encoded = strings.ReplaceAll(encoded, "+", "-")
	encoded = strings.ReplaceAll(encoded, "/", "_")
	encoded = strings.TrimRight(encoded, "=")
	return encoded
}

func handleAppPasswords(w http.ResponseWriter, r *http.Request) {
	logRequest(r)
	switch r.Method {
	case http.MethodPost:
		createAppPassword(w, r)
	case http.MethodGet:
		listAppPasswords(w, r)
	case http.MethodDelete:
		deleteAppPassword(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func handleRolesRequest(w http.ResponseWriter, r *http.Request) {
	logRequest(r)
	roles := []FronteggRole{
		{ID: "1", Name: "Organization Admin"},
		{ID: "2", Name: "Organization Member"},
	}

	response := FronteggRolesResponse{
		Items: roles,
		Metadata: struct {
			TotalItems int `json:"totalItems"`
			TotalPages int `json:"totalPages"`
		}{
			TotalItems: len(roles),
			TotalPages: 1,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func createAppPassword(w http.ResponseWriter, r *http.Request) {
	logRequest(r)
	var req struct {
		Description string `json:"description"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Generate a new app password
	newAppPassword := AppPassword{
		ClientID:    generateClientID(),
		Secret:      generateSecret(),
		Description: req.Description,
		Owner:       "mockOwner",
		CreatedAt:   time.Now(),
	}

	// Store the new app password
	mutex.Lock()
	appPasswords[newAppPassword.ClientID] = newAppPassword
	mutex.Unlock()

	// Send the response back
	sendResponse(w, http.StatusCreated, newAppPassword)
}

func listAppPasswords(w http.ResponseWriter, r *http.Request) {
	mutex.Lock()
	passwords := make([]AppPassword, 0, len(appPasswords))
	for _, password := range appPasswords {
		passwords = append(passwords, password)
	}
	mutex.Unlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(passwords)
}

func deleteAppPassword(w http.ResponseWriter, r *http.Request) {
	clientID := r.URL.Query().Get("clientId")
	if clientID == "" {
		http.Error(w, "Client ID is required", http.StatusBadRequest)
		return
	}

	mutex.Lock()
	delete(appPasswords, clientID)
	mutex.Unlock()

	w.WriteHeader(http.StatusOK)
}

// generateClientID generates a unique client ID.
func generateClientID() string {
	return fmt.Sprintf("client-%d", time.Now().UnixNano())
}

// generateSecret generates a secret.
func generateSecret() string {
	return fmt.Sprintf("secret-%d", time.Now().UnixNano())
}

func sendResponse(w http.ResponseWriter, statusCode int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if payload != nil {
		responseBytes, _ := json.Marshal(payload)
		fmt.Printf("Response body: %s\n", string(responseBytes))
		w.Write(responseBytes)
	}
}

func logRequest(r *http.Request) {
	fmt.Printf("Received request: %s %s\n", r.Method, r.URL.Path)
	if r.Body != nil {
		bodyBytes, err := io.ReadAll(r.Body)
		if err == nil {
			fmt.Printf("Request body: %s\n", string(bodyBytes))
			// Important: Restore the body for further reading
			r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
		}
	}
}
