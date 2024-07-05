package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
)

type AppPassword struct {
	ClientID    string    `json:"clientId"`
	Secret      string    `json:"secret"`
	Description string    `json:"description"`
	Owner       string    `json:"owner"`
	CreatedAt   time.Time `json:"created_at"`
}

type User struct {
	ID                string         `json:"id"`
	Email             string         `json:"email"`
	ProfilePictureURL string         `json:"profilePictureUrl"`
	Verified          bool           `json:"verified"`
	Metadata          string         `json:"metadata"`
	Roles             []FronteggRole `json:"roles"`
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
	SsoConfigId string `json:"ssoConfigId"`
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

// SCIM 2.0 Configurations API response
type SCIM2Configuration struct {
	ID                   string    `json:"id"`
	Source               string    `json:"source"`
	TenantID             string    `json:"tenantId"`
	ConnectionName       string    `json:"connectionName"`
	SyncToUserManagement bool      `json:"syncToUserManagement"`
	CreatedAt            time.Time `json:"createdAt"`
	Token                string    `json:"token"`
}

type SCIM2ConfigurationsResponse []SCIM2Configuration

// GroupCreateParams represents the parameters for creating a new group.
type GroupCreateParams struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Color       string `json:"color,omitempty"`
	Metadata    string `json:"metadata,omitempty"`
}

// GroupUpdateParams represents the parameters for updating an existing group.
type GroupUpdateParams struct {
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
	Color       string `json:"color,omitempty"`
	Metadata    string `json:"metadata,omitempty"`
}

// ScimGroup represents the structure of a group in the response.
type ScimGroup struct {
	ID          string     `json:"id"`
	Name        string     `json:"name"`
	Description string     `json:"description"`
	Metadata    string     `json:"metadata"`
	Roles       []ScimRole `json:"roles"`
	Users       []ScimUser `json:"users"`
	ManagedBy   string     `json:"managedBy"`
	Color       string     `json:"color"`
}

// ScimRole represents the structure of a role within a group.
type ScimRole struct {
	ID          string `json:"id"`
	Key         string `json:"key"`
	Name        string `json:"name"`
	Description string `json:"description"`
	IsDefault   bool   `json:"is_default"`
}

// ScimUser represents the structure of a user within a group.
type ScimUser struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

// SCIMGroupsResponse represents the overall structure of the response from the SCIM groups API.
type SCIMGroupsResponse struct {
	Groups []ScimGroup `json:"groups"`
}

// AddRolesToGroupParams represents the parameters for adding roles to a group.
type AddRolesToGroupParams struct {
	RoleIds []string `json:"roleIds"`
}

// TenantApiTokenRequest represents the structure of a request to create a tenant API token.
type TenantApiTokenRequest struct {
	Description string            `json:"description"`
	Metadata    map[string]string `json:"metadata"`
	RoleIDs     []string          `json:"roleIds"`
}

// TenantApiTokenResponse represents the structure of a response from creating a tenant API token.
type TenantApiTokenResponse struct {
	ClientID        string            `json:"clientId"`
	Description     string            `json:"description"`
	Secret          string            `json:"secret"`
	CreatedByUserId string            `json:"createdByUserId"`
	Metadata        map[string]string `json:"metadata"`
	CreatedAt       time.Time         `json:"createdAt"`
	RoleIDs         []string          `json:"roleIds"`
}

var (
	appPasswords       = make(map[string]AppPassword)
	tenantAppPasswords = make(map[string]TenantApiTokenResponse)
	users              = make(map[string]User)
	ssoConfigs         = make(map[string]SSOConfig)
	scimConfigurations = make(map[string]SCIM2Configuration)
	groups             = make(map[string]ScimGroup)
	mutex              = &sync.Mutex{}
)

func main() {
	http.HandleFunc("/identity/resources/auth/v1/api-token", handleTokenRequest)
	http.HandleFunc("/identity/resources/users/api-tokens/v1", handleAppPasswords)
	http.HandleFunc("/identity/resources/users/api-tokens/v1/", handleAppPasswordsDelete)
	http.HandleFunc("/identity/resources/tenants/api-tokens/v1", handleTenantAppPasswords)
	http.HandleFunc("/identity/resources/tenants/api-tokens/v1/", handleTenantAppPasswordsDelete)
	http.HandleFunc("/identity/resources/users/v1/", handleUserRequest)
	http.HandleFunc("/identity/resources/users/v2", handleUserRequest)
	http.HandleFunc("/identity/resources/roles/v2", handleRolesRequest)
	http.HandleFunc("/identity/resources/users/v3", handleUserV3Request)
	http.HandleFunc("/frontegg/team/resources/sso/v1/configurations", handleSSOConfigRequest)
	http.HandleFunc("/frontegg/team/resources/sso/v1/configurations/", handleSSOConfigAndDomainRequest)
	http.HandleFunc("/frontegg/identity/resources/groups/v1", handleSCIMGroupsRequest)
	http.HandleFunc("/frontegg/identity/resources/groups/v1/", handleSCIMGroupsParamRequest)
	http.HandleFunc("/frontegg/directory/resources/v1/configurations/scim2", handleSCIM2ConfigurationsRequest)
	http.HandleFunc("/frontegg/directory/resources/v1/configurations/scim2/", handleSCIMConfigurationByID)

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

// Add this function to handle the new endpoint
func handleUserV3Request(w http.ResponseWriter, r *http.Request) {
	logRequest(r)
	switch r.Method {
	case http.MethodGet:
		getUsersV3(w, r)
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

func getUsersV3(w http.ResponseWriter, r *http.Request) {
	logRequest(r)

	// Parse query parameters
	query := r.URL.Query()
	email := query.Get("_email")
	limit := query.Get("_limit")
	offset := query.Get("_offset")
	ids := query.Get("ids")
	sortBy := query.Get("_sortBy")
	order := query.Get("_order")

	// Convert limit and offset to integers
	limitInt, _ := strconv.Atoi(limit)
	offsetInt, _ := strconv.Atoi(offset)

	// Filter and collect users
	var filteredUsers []User
	mutex.Lock()
	for _, user := range users {
		if (email == "" || user.Email == email) &&
			(ids == "" || strings.Contains(ids, user.ID)) {
			filteredUsers = append(filteredUsers, user)
		}
	}
	mutex.Unlock()

	// Sort users if sortBy is provided
	if sortBy != "" {
		sort.Slice(filteredUsers, func(i, j int) bool {
			var less bool
			switch sortBy {
			case "email":
				less = filteredUsers[i].Email < filteredUsers[j].Email
			case "id":
				less = filteredUsers[i].ID < filteredUsers[j].ID
			// Add more cases for other sortable fields
			default:
				return false
			}

			if order == "desc" {
				return !less
			}
			return less
		})
	}

	// Apply pagination
	totalItems := len(filteredUsers)
	if offsetInt >= totalItems {
		filteredUsers = []User{}
	} else {
		end := offsetInt + limitInt
		if end > totalItems {
			end = totalItems
		}
		filteredUsers = filteredUsers[offsetInt:end]
	}

	// Prepare response
	response := struct {
		Items    []User `json:"items"`
		Metadata struct {
			TotalItems int `json:"totalItems"`
		} `json:"_metadata"`
	}{
		Items: filteredUsers,
		Metadata: struct {
			TotalItems int `json:"totalItems"`
		}{
			TotalItems: totalItems,
		},
	}

	// Send response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func createUser(w http.ResponseWriter, r *http.Request) {
	var newUser struct {
		User
		RoleIDs         []string `json:"roleIds"`
		SkipInviteEmail bool     `json:"skipInviteEmail"`
	}

	if err := json.NewDecoder(r.Body).Decode(&newUser); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	userID := generateUserID()
	newUser.ID = userID

	// Map role IDs to role names and update the newUser.Roles slice
	for _, roleID := range newUser.RoleIDs {
		var roleName string
		switch roleID {
		case "1":
			roleName = "Organization Admin"
		case "2":
			roleName = "Organization Member"
		}

		if roleName != "" {
			newUser.Roles = append(newUser.Roles, FronteggRole{ID: roleID, Name: roleName})
		}
	}

	mutex.Lock()
	users[userID] = newUser.User
	mutex.Unlock()

	w.WriteHeader(http.StatusCreated)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(newUser.User)
}

func generateUserID() string {
	return fmt.Sprintf("user-%d", time.Now().UnixNano())
}

func generateConfigID() string {
	return fmt.Sprintf("config-%d", time.Now().UnixNano())
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

func handleAppPasswordsDelete(w http.ResponseWriter, r *http.Request) {
	logRequest(r)
	switch r.Method {
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
	clientID := strings.TrimPrefix(r.URL.Path, "/identity/resources/users/api-tokens/v1/")
	if clientID == "" {
		http.Error(w, "Client ID is required", http.StatusBadRequest)
		return
	}

	mutex.Lock()
	delete(appPasswords, clientID)
	mutex.Unlock()

	w.WriteHeader(http.StatusOK)
}

// HandleTenantAppPasswords provides a single entry point for POST and GET methods.
func handleTenantAppPasswords(w http.ResponseWriter, r *http.Request) {
	logRequest(r)
	switch r.Method {
	case http.MethodPost:
		createTenantAppPassword(w, r)
	case http.MethodGet:
		listTenantAppPasswords(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// HandleTenantAppPasswordsDelete handles the DELETE method.
func handleTenantAppPasswordsDelete(w http.ResponseWriter, r *http.Request) {
	logRequest(r)
	if r.Method == http.MethodDelete {
		deleteTenantAppPassword(w, r)
	} else {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// Create a new tenant app password
func createTenantAppPassword(w http.ResponseWriter, r *http.Request) {
	logRequest(r)
	var req TenantApiTokenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Simulate token creation logic
	newToken := TenantApiTokenResponse{
		ClientID:        generateClientID(),
		Secret:          generateSecret(),
		Description:     req.Description,
		CreatedByUserId: "mockUser",
		CreatedAt:       time.Now(),
		Metadata:        req.Metadata,
		RoleIDs:         req.RoleIDs,
	}

	mutex.Lock()
	tenantAppPasswords[newToken.ClientID] = newToken
	mutex.Unlock()

	sendResponse(w, http.StatusCreated, newToken)
}

// List all tenant app passwords
func listTenantAppPasswords(w http.ResponseWriter, r *http.Request) {
	mutex.Lock()
	passwords := make([]TenantApiTokenResponse, 0, len(tenantAppPasswords))
	for _, password := range tenantAppPasswords {
		passwords = append(passwords, password)
	}
	mutex.Unlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(passwords)
}

// Delete a tenant app password
func deleteTenantAppPassword(w http.ResponseWriter, r *http.Request) {
	clientID := strings.TrimPrefix(r.URL.Path, "/identity/resources/tenants/api-tokens/v1"+"/")
	if clientID == "" {
		http.Error(w, "Client ID is required", http.StatusBadRequest)
		return
	}

	mutex.Lock()
	delete(tenantAppPasswords, clientID)
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

func handleSSOConfigRequest(w http.ResponseWriter, r *http.Request) {
	logRequest(r)
	switch r.Method {
	case http.MethodPost:
		createSSOConfig(w, r)
	case http.MethodGet:
		listSSOConfigs(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func handleSSOConfigAndDomainRequest(w http.ResponseWriter, r *http.Request) {
	logRequest(r)
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 8 {
		http.Error(w, "Invalid URL", http.StatusBadRequest)
		return
	}

	ssoConfigID := parts[7]

	if len(parts) > 8 {
		switch parts[8] {
		case "domains":
			handleDomainRequests(w, r, ssoConfigID, parts)
		case "groups":
			handleGroupMappingRequests(w, r, ssoConfigID, parts)
		case "roles":
			handleDefaultRolesRequests(w, r, ssoConfigID)
		default:
			http.Error(w, "Invalid request", http.StatusBadRequest)
		}
	} else {
		switch r.Method {
		case http.MethodGet:
			getSSOConfig(w, r, ssoConfigID)
		case http.MethodPatch:
			updateSSOConfig(w, r, ssoConfigID)
		case http.MethodDelete:
			deleteSSOConfig(w, r, ssoConfigID)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}
}

func handleDomainRequests(w http.ResponseWriter, r *http.Request, ssoConfigID string, parts []string) {
	domainID := ""
	if len(parts) > 9 {
		domainID = parts[9]
	}

	switch r.Method {
	case http.MethodPost:
		createDomain(w, r, ssoConfigID)
	case http.MethodGet:
		if domainID == "" {
			listDomains(w, ssoConfigID)
		} else {
			getDomain(w, ssoConfigID, domainID)
		}
	case http.MethodPatch:
		updateDomain(w, r, ssoConfigID, domainID)
	case http.MethodDelete:
		deleteDomain(w, ssoConfigID, domainID)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func handleGroupMappingRequests(w http.ResponseWriter, r *http.Request, ssoConfigID string, parts []string) {
	groupMappingID := ""
	if len(parts) > 9 {
		groupMappingID = parts[9]
	}

	switch r.Method {
	case http.MethodPost:
		createGroupMapping(w, r, ssoConfigID)
	case http.MethodGet:
		if groupMappingID == "" {
			listGroupMappings(w, ssoConfigID)
		} else {
			getGroupMapping(w, ssoConfigID, groupMappingID)
		}
	case http.MethodPatch:
		updateGroupMapping(w, r, ssoConfigID, groupMappingID)
	case http.MethodDelete:
		deleteGroupMapping(w, ssoConfigID, groupMappingID)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func handleSCIMGroupsParamRequest(w http.ResponseWriter, r *http.Request) {
	logRequest(r)
	// Extract the group ID and potential action from the URL path
	trimmedPath := strings.TrimPrefix(r.URL.Path, "/frontegg/identity/resources/groups/v1/")
	trimmedPath = strings.Split(trimmedPath, "?")[0]
	parts := strings.Split(trimmedPath, "/")
	groupID := ""
	if len(parts) > 0 {
		groupID = parts[0]
	}

	switch r.Method {
	case http.MethodPost:
		if strings.Contains(trimmedPath, "/roles") && groupID != "" {
			// Add roles to a group
			handleAddRolesToGroup(w, r, groupID)
		} else if strings.Contains(trimmedPath, "/users") && groupID != "" {
			// Add users to a group
			handleAddUsersToGroup(w, r, groupID)
		} else {
			http.Error(w, "Invalid request for POST method", http.StatusBadRequest)
		}
	case http.MethodPatch:
		if groupID != "" {
			// Update a group
			handleUpdateScimGroup(w, r, groupID)
		} else {
			http.Error(w, "Group ID is required for PATCH method", http.StatusBadRequest)
		}
	case http.MethodDelete:
		if strings.Contains(trimmedPath, "/roles") && groupID != "" {
			// Remove roles from a group
			handleRemoveRolesFromGroup(w, r, groupID)
		} else if strings.Contains(trimmedPath, "/users") && groupID != "" {
			// Remove users from a group
			handleRemoveUsersFromGroup(w, r, groupID)
		} else if groupID != "" {
			// Delete a group
			handleDeleteScimGroup(w, r, groupID)
		} else {
			http.Error(w, "Group ID is required for DELETE method", http.StatusBadRequest)
		}
	case http.MethodGet:
		if groupID != "" {
			// Get a specific group by ID
			handleGetScimGroupByID(w, r, groupID)
		} else {
			// List all groups
			listSCIMGroups(w, r)
		}
	default:
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
	}
}

// Handle scim groups create and list
func handleSCIMGroupsRequest(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		listSCIMGroups(w, r)
	case http.MethodPost:
		handleCreateScimGroup(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func handleSCIM2ConfigurationsRequest(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		listSCIMConfigurations(w)
	case http.MethodPost:
		createSCIMConfiguration(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func listSCIMGroups(w http.ResponseWriter, r *http.Request) {
	mutex.Lock()
	defer mutex.Unlock()

	allGroups := make([]ScimGroup, 0, len(groups))
	for _, group := range groups {
		allGroups = append(allGroups, group)
	}

	// Respond with all the groups encoded as JSON
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(SCIMGroupsResponse{Groups: allGroups})
}

func handleCreateScimGroup(w http.ResponseWriter, r *http.Request) {
	var params GroupCreateParams
	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Create a new group with the provided parameters
	newGroup := ScimGroup{
		ID:          uuid.New().String(),
		Name:        params.Name,
		Description: params.Description,
		Metadata:    params.Metadata,
		Roles:       []ScimRole{},
		Users:       []ScimUser{},
	}

	// Store the new group in the mock data store
	mutex.Lock()
	groups[newGroup.ID] = newGroup
	mutex.Unlock()

	// Respond with the newly created group
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(newGroup)
}

func handleUpdateScimGroup(w http.ResponseWriter, r *http.Request, groupID string) {
	var params GroupUpdateParams
	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Lock the mutex before accessing the shared resource
	mutex.Lock()
	group, exists := groups[groupID]
	mutex.Unlock()

	if !exists {
		http.Error(w, "Group not found", http.StatusNotFound)
		return
	}

	// Update the group's attributes if they are provided in the request
	if params.Name != "" {
		group.Name = params.Name
	}
	if params.Description != "" {
		group.Description = params.Description
	}
	if params.Color != "" {
		group.Color = params.Color
	}
	if params.Metadata != "" {
		group.Metadata = params.Metadata
	}

	// Update the group in the mock data store
	mutex.Lock()
	groups[groupID] = group
	mutex.Unlock()

	// Respond with the updated group data
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(group)
}

func handleDeleteScimGroup(w http.ResponseWriter, r *http.Request, groupID string) {
	// Lock the mutex before accessing the shared resource
	mutex.Lock()
	defer mutex.Unlock()

	// Check if the group exists
	_, exists := groups[groupID]
	if !exists {
		http.Error(w, "Group not found", http.StatusNotFound)
		return
	}

	// Delete the group from the mock data store
	delete(groups, groupID)

	// Respond with a 200 OK status to indicate successful deletion
	w.WriteHeader(http.StatusOK)
}

func handleGetScimGroupByID(w http.ResponseWriter, r *http.Request, groupID string) {
	// Lock the mutex before accessing the shared resource
	mutex.Lock()
	group, exists := groups[groupID]
	mutex.Unlock()

	if !exists {
		// If the group does not exist, return a 404 Not Found status
		http.Error(w, "Group not found", http.StatusNotFound)
		return
	}

	// If the group exists, encode it to JSON and return it with a 200 OK status
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(group)
}

func handleSCIMConfigurationByID(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodDelete:
		deleteSCIMConfiguration(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func handleAddRolesToGroup(w http.ResponseWriter, r *http.Request, groupID string) {
	logRequest(r)
	var params AddRolesToGroupParams
	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Lock the mutex before accessing the shared resource
	mutex.Lock()
	defer mutex.Unlock()

	group, exists := groups[groupID]
	if !exists {
		// If the group does not exist, return a 404 Not Found status
		http.Error(w, "Group not found", http.StatusNotFound)
		return
	}

	for _, roleID := range params.RoleIds {
		// Check if the role is already in the group to prevent duplicates
		found := false
		for _, role := range group.Roles {
			if role.ID == roleID {
				found = true
				break
			}
		}

		// If the role is not found, add it to the group with a specific name based on its ID
		if !found {
			roleName := ""
			switch roleID {
			case "1":
				roleName = "Organization Admin"
			case "2":
				roleName = "Organization Member"
			}
			group.Roles = append(group.Roles, ScimRole{ID: roleID, Name: roleName})
		}
	}

	// Update the group in the mock data store
	groups[groupID] = group

	// Respond with a 201 Created status
	w.WriteHeader(http.StatusCreated)
}

func handleRemoveRolesFromGroup(w http.ResponseWriter, r *http.Request, groupID string) {
	var params AddRolesToGroupParams
	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Lock the mutex before accessing the shared resource
	mutex.Lock()
	defer mutex.Unlock()

	group, exists := groups[groupID]
	if !exists {
		// If the group does not exist, return a 404 Not Found status
		http.Error(w, "Group not found", http.StatusNotFound)
		return
	}

	// Remove the specified roles from the group
	for _, roleID := range params.RoleIds {
		for i, role := range group.Roles {
			if role.ID == roleID {
				// Remove the role from the slice
				group.Roles = append(group.Roles[:i], group.Roles[i+1:]...)
				break
			}
		}
	}

	// Update the group in the mock data store
	groups[groupID] = group

	// Respond with a 200 OK status to indicate successful removal
	w.WriteHeader(http.StatusOK)
}

func handleAddUsersToGroup(w http.ResponseWriter, r *http.Request, groupID string) {
	var params struct {
		UserIds []string `json:"userIds"`
	}
	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Lock the mutex before accessing the shared resource
	mutex.Lock()
	defer mutex.Unlock()

	group, exists := groups[groupID]
	if !exists {
		// If the group does not exist, return a 404 Not Found status
		http.Error(w, "Group not found", http.StatusNotFound)
		return
	}

	// Add the specified users to the group, avoiding duplicates
	for _, userID := range params.UserIds {
		// Check if the user is already in the group
		found := false
		for _, user := range group.Users {
			if user.ID == userID {
				found = true
				break
			}
		}

		// If the user is not found, add them to the group
		if !found {
			group.Users = append(group.Users, ScimUser{ID: userID})
		}
	}

	// Update the group in the mock data store
	groups[groupID] = group

	// Respond with a 201 Created status
	w.WriteHeader(http.StatusCreated)
}

func handleRemoveUsersFromGroup(w http.ResponseWriter, r *http.Request, groupID string) {
	var params struct {
		UserIds []string `json:"userIds"`
	}
	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Lock the mutex before accessing the shared resource
	mutex.Lock()
	defer mutex.Unlock()

	group, exists := groups[groupID]
	if !exists {
		// If the group does not exist, return a 404 Not Found status
		http.Error(w, "Group not found", http.StatusNotFound)
		return
	}

	// Remove the specified users from the group
	for _, userID := range params.UserIds {
		for i, user := range group.Users {
			if user.ID == userID {
				// Remove the user from the group
				group.Users = append(group.Users[:i], group.Users[i+1:]...)
				break
			}
		}
	}

	// Update the group in the mock data store
	groups[groupID] = group

	// Respond with a 200 OK status to indicate successful removal
	w.WriteHeader(http.StatusOK)
}

func listSCIMConfigurations(w http.ResponseWriter) {
	mutex.Lock()
	configs := make([]SCIM2Configuration, 0, len(scimConfigurations))
	for _, config := range scimConfigurations {
		configs = append(configs, config)
	}
	for i, config := range configs {
		if config.CreatedAt.IsZero() {
			configs[i].CreatedAt = time.Now()
		}
	}
	mutex.Unlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(configs)
}

func createSCIMConfiguration(w http.ResponseWriter, r *http.Request) {
	var newConfig SCIM2Configuration
	if err := json.NewDecoder(r.Body).Decode(&newConfig); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	newConfig.ID = generateMockUUID()
	newConfig.Token = generateMockUUID()
	newConfig.TenantID = "mockTenantID"
	newConfig.CreatedAt = time.Now()

	// Log the configuration
	fmt.Printf("Received SCIM 2.0 configuration: %+v\n", newConfig)

	mutex.Lock()
	scimConfigurations[newConfig.ID] = newConfig
	mutex.Unlock()

	// log response
	responseBytes, _ := json.Marshal(newConfig)
	fmt.Printf("Response body: %s\n", string(responseBytes))

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(newConfig)
}

func deleteSCIMConfiguration(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Path[len("/frontegg/directory/resources/v1/configurations/scim2/"):]

	mutex.Lock()
	if _, exists := scimConfigurations[id]; !exists {
		mutex.Unlock()
		http.Error(w, "Configuration not found", http.StatusNotFound)
		return
	}

	delete(scimConfigurations, id)
	mutex.Unlock()

	w.WriteHeader(http.StatusNoContent)
}

func createSSOConfig(w http.ResponseWriter, r *http.Request) {
	var newConfig SSOConfig
	if err := json.NewDecoder(r.Body).Decode(&newConfig); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Adjusting fields to match production data
	newConfig.Id = generateConfigID()
	newConfig.PublicCertificate = base64.StdEncoding.EncodeToString([]byte(newConfig.PublicCertificate))
	newConfig.CreatedAt = time.Now()
	newConfig.UpdatedAt = newConfig.CreatedAt
	newConfig.GeneratedVerification = generateMockUUID()
	newConfig.ConfigMetadata = nil
	newConfig.OverrideActiveTenant = true
	newConfig.SubAccountAccessLimit = 0
	newConfig.SkipEmailDomainValidation = false

	newConfig.RoleIds = newConfig.DefaultRoles.RoleIds

	for i, group := range newConfig.Groups {
		group.Enabled = true
		newConfig.Groups[i] = group
	}

	mutex.Lock()
	ssoConfigs[newConfig.Id] = newConfig
	mutex.Unlock()

	sendResponse(w, http.StatusCreated, newConfig)
}

func listSSOConfigs(w http.ResponseWriter, r *http.Request) {
	mutex.Lock()
	configs := make([]SSOConfig, 0, len(ssoConfigs))
	for _, config := range ssoConfigs {
		// Ensure that Domains and Groups are not nil
		if config.Domains == nil {
			config.Domains = []Domain{}
		}
		if config.Groups == nil {
			config.Groups = []GroupMapping{}
		}
		// Initialize RoleIds if it's nil
		if config.RoleIds == nil {
			config.RoleIds = []string{}
		}
		configs = append(configs, config)
	}
	mutex.Unlock()

	responseBytes, err := json.Marshal(configs)
	if err != nil {
		fmt.Printf("Error marshaling response: %v\n", err)
		sendResponse(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	fmt.Printf("Response body: %s\n", string(responseBytes))
	sendResponse(w, http.StatusOK, configs)
}

func getSSOConfig(w http.ResponseWriter, r *http.Request, configID string) {
	var updatedConfig SSOConfig
	if err := json.NewDecoder(r.Body).Decode(&updatedConfig); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	updatedConfig.UpdatedAt = time.Now()
	mutex.Lock()
	config, ok := ssoConfigs[configID]
	mutex.Unlock()

	if !ok {
		http.Error(w, "SSO configuration not found", http.StatusNotFound)
		return
	}

	sendResponse(w, http.StatusOK, config)
}

func updateSSOConfig(w http.ResponseWriter, r *http.Request, configID string) {
	var updatedConfig SSOConfig
	if err := json.NewDecoder(r.Body).Decode(&updatedConfig); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	mutex.Lock()
	if _, ok := ssoConfigs[configID]; !ok {
		mutex.Unlock()
		http.Error(w, "SSO configuration not found", http.StatusNotFound)
		return
	}

	updatedConfig.Id = configID
	updatedConfig.PublicCertificate = base64.StdEncoding.EncodeToString([]byte(updatedConfig.PublicCertificate))
	ssoConfigs[configID] = updatedConfig
	mutex.Unlock()

	sendResponse(w, http.StatusOK, updatedConfig)
}

func deleteSSOConfig(w http.ResponseWriter, r *http.Request, configID string) {
	mutex.Lock()
	if _, ok := ssoConfigs[configID]; !ok {
		mutex.Unlock()
		http.Error(w, "SSO configuration not found", http.StatusNotFound)
		return
	}

	delete(ssoConfigs, configID)
	mutex.Unlock()

	w.WriteHeader(http.StatusOK)
}

func createDomain(w http.ResponseWriter, r *http.Request, ssoConfigID string) {
	var newDomain Domain
	if err := json.NewDecoder(r.Body).Decode(&newDomain); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	mutex.Lock()
	defer mutex.Unlock()

	config, exists := ssoConfigs[ssoConfigID]
	if !exists {
		http.Error(w, "SSO configuration not found", http.StatusNotFound)
		return
	}

	newDomain.ID = generateDomainID()
	newDomain.SsoConfigId = ssoConfigID
	config.Domains = append(config.Domains, newDomain)
	ssoConfigs[ssoConfigID] = config

	sendResponse(w, http.StatusCreated, newDomain)
}

func generateDomainID() string {
	return fmt.Sprintf("domain-%d", time.Now().UnixNano())
}

func listDomains(w http.ResponseWriter, ssoConfigID string) {
	mutex.Lock()
	defer mutex.Unlock()

	config, exists := ssoConfigs[ssoConfigID]
	if !exists {
		http.Error(w, "SSO configuration not found", http.StatusNotFound)
		return
	}

	sendResponse(w, http.StatusOK, config.Domains)
}

func getDomain(w http.ResponseWriter, ssoConfigID string, domainID string) {
	mutex.Lock()
	defer mutex.Unlock()

	config, exists := ssoConfigs[ssoConfigID]
	if !exists {
		http.Error(w, "SSO configuration not found", http.StatusNotFound)
		return
	}

	for _, domain := range config.Domains {
		if domain.ID == domainID {
			sendResponse(w, http.StatusOK, domain)
			return
		}
	}

	http.Error(w, "Domain not found", http.StatusNotFound)
}

func updateDomain(w http.ResponseWriter, r *http.Request, ssoConfigID string, domainID string) {
	var updatedDomain Domain
	if err := json.NewDecoder(r.Body).Decode(&updatedDomain); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	mutex.Lock()
	defer mutex.Unlock()

	config, exists := ssoConfigs[ssoConfigID]
	if !exists {
		http.Error(w, "SSO configuration not found", http.StatusNotFound)
		return
	}

	for i, domain := range config.Domains {
		if domain.ID == domainID {
			updatedDomain.ID = domainID
			config.Domains[i] = updatedDomain
			ssoConfigs[ssoConfigID] = config
			sendResponse(w, http.StatusOK, updatedDomain)
			return
		}
	}

	http.Error(w, "Domain not found", http.StatusNotFound)
}

func deleteDomain(w http.ResponseWriter, ssoConfigID string, domainID string) {
	mutex.Lock()
	defer mutex.Unlock()

	config, exists := ssoConfigs[ssoConfigID]
	if !exists {
		http.Error(w, "SSO configuration not found", http.StatusNotFound)
		return
	}

	for i, domain := range config.Domains {
		if domain.ID == domainID {
			config.Domains = append(config.Domains[:i], config.Domains[i+1:]...)
			ssoConfigs[ssoConfigID] = config
			w.WriteHeader(http.StatusOK)
			return
		}
	}

	http.Error(w, "Domain not found", http.StatusNotFound)
}

func createGroupMapping(w http.ResponseWriter, r *http.Request, ssoConfigID string) {
	var newGroupMapping GroupMapping
	if err := json.NewDecoder(r.Body).Decode(&newGroupMapping); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	mutex.Lock()
	defer mutex.Unlock()

	config, exists := ssoConfigs[ssoConfigID]
	if !exists {
		http.Error(w, "SSO configuration not found", http.StatusNotFound)
		return
	}

	newGroupMapping.ID = generateGroupMappingID()
	newGroupMapping.SsoConfigId = ssoConfigID
	config.Groups = append(config.Groups, newGroupMapping)
	ssoConfigs[ssoConfigID] = config

	sendResponse(w, http.StatusCreated, newGroupMapping)
}

func generateGroupMappingID() string {
	return fmt.Sprintf("groupmap-%d", time.Now().UnixNano())
}

func listGroupMappings(w http.ResponseWriter, ssoConfigID string) {
	mutex.Lock()
	defer mutex.Unlock()

	config, exists := ssoConfigs[ssoConfigID]
	if !exists {
		http.Error(w, "SSO configuration not found", http.StatusNotFound)
		return
	}

	sendResponse(w, http.StatusOK, config.Groups)
}

func getGroupMapping(w http.ResponseWriter, ssoConfigID string, groupMappingID string) {
	mutex.Lock()
	defer mutex.Unlock()

	config, exists := ssoConfigs[ssoConfigID]
	if !exists {
		http.Error(w, "SSO configuration not found", http.StatusNotFound)
		return
	}

	for _, mapping := range config.Groups {
		if mapping.ID == groupMappingID {
			sendResponse(w, http.StatusOK, mapping)
			return
		}
	}

	http.Error(w, "Group mapping not found", http.StatusNotFound)
}

func updateGroupMapping(w http.ResponseWriter, r *http.Request, ssoConfigID string, groupMappingID string) {
	var updatedMapping GroupMapping
	if err := json.NewDecoder(r.Body).Decode(&updatedMapping); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	mutex.Lock()
	defer mutex.Unlock()

	config, exists := ssoConfigs[ssoConfigID]
	if !exists {
		http.Error(w, "SSO configuration not found", http.StatusNotFound)
		return
	}

	for i, mapping := range config.Groups {
		if mapping.ID == groupMappingID {
			updatedMapping.ID = groupMappingID
			config.Groups[i] = updatedMapping
			ssoConfigs[ssoConfigID] = config
			sendResponse(w, http.StatusOK, updatedMapping)
			return
		}
	}

	http.Error(w, "Group mapping not found", http.StatusNotFound)
}

func deleteGroupMapping(w http.ResponseWriter, ssoConfigID string, groupMappingID string) {
	mutex.Lock()
	defer mutex.Unlock()

	config, exists := ssoConfigs[ssoConfigID]
	if !exists {
		http.Error(w, "SSO configuration not found", http.StatusNotFound)
		return
	}

	for i, mapping := range config.Groups {
		if mapping.ID == groupMappingID {
			config.Groups = append(config.Groups[:i], config.Groups[i+1:]...)
			ssoConfigs[ssoConfigID] = config
			w.WriteHeader(http.StatusOK)
			return
		}
	}

	http.Error(w, "Group mapping not found", http.StatusNotFound)
}

func handleDefaultRolesRequests(w http.ResponseWriter, r *http.Request, ssoConfigID string) {
	switch r.Method {
	case http.MethodPut:
		setDefaultRoles(w, r, ssoConfigID)
	case http.MethodGet:
		getDefaultRoles(w, ssoConfigID)
	case http.MethodDelete:
		clearDefaultRoles(w, ssoConfigID)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func setDefaultRoles(w http.ResponseWriter, r *http.Request, ssoConfigID string) {
	var roles DefaultRoles
	if err := json.NewDecoder(r.Body).Decode(&roles); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	mutex.Lock()
	defer mutex.Unlock()

	config, exists := ssoConfigs[ssoConfigID]
	if !exists {
		http.Error(w, "SSO configuration not found", http.StatusNotFound)
		return
	}

	config.DefaultRoles = roles
	config.RoleIds = roles.RoleIds

	ssoConfigs[ssoConfigID] = config

	sendResponse(w, http.StatusCreated, roles)
}

func getDefaultRoles(w http.ResponseWriter, ssoConfigID string) {
	mutex.Lock()
	config, exists := ssoConfigs[ssoConfigID]
	mutex.Unlock()

	if !exists {
		http.Error(w, "SSO configuration not found", http.StatusNotFound)
		return
	}

	sendResponse(w, http.StatusOK, config.DefaultRoles)
}

func clearDefaultRoles(w http.ResponseWriter, ssoConfigID string) {
	mutex.Lock()
	defer mutex.Unlock()

	config, exists := ssoConfigs[ssoConfigID]
	if !exists {
		http.Error(w, "SSO configuration not found", http.StatusNotFound)
		return
	}

	config.DefaultRoles = DefaultRoles{RoleIds: []string{}}
	ssoConfigs[ssoConfigID] = config

	w.WriteHeader(http.StatusOK)
}

func generateMockUUID() string {
	return uuid.New().String()
}
