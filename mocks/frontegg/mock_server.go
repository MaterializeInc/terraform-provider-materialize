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

var (
	appPasswords = make(map[string]AppPassword)
	users        = make(map[string]User)
	ssoConfigs   = make(map[string]SSOConfig)
	mutex        = &sync.Mutex{}
)

func main() {
	http.HandleFunc("/identity/resources/auth/v1/api-token", handleTokenRequest)
	http.HandleFunc("/identity/resources/users/api-tokens/v1", handleAppPasswords)
	http.HandleFunc("/identity/resources/users/v1/", handleUserRequest)
	http.HandleFunc("/identity/resources/users/v2", handleUserRequest)
	http.HandleFunc("/identity/resources/roles/v2", handleRolesRequest)
	http.HandleFunc("/frontegg/team/resources/sso/v1/configurations", handleSSOConfigRequest)
	http.HandleFunc("/frontegg/team/resources/sso/v1/configurations/", handleSSOConfigAndDomainRequest)
	http.HandleFunc("/frontegg/identity/resources/groups/v1", handleSCIMGroupsRequest)

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

func handleSCIMGroupsRequest(w http.ResponseWriter, r *http.Request) {
	logRequest(r)
	switch r.Method {
	case http.MethodGet:
		listSCIMGroups(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func listSCIMGroups(w http.ResponseWriter, r *http.Request) {
	// TODO: update this to return the groups that are created by the user
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"groups":[]}`))
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
		configs = append(configs, config)
	}
	mutex.Unlock()

	responseBytes, _ := json.Marshal(configs)
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
