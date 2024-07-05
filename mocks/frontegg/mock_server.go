package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

// Struct definitions
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

type SCIM2Configuration struct {
	ID                   string    `json:"id"`
	Source               string    `json:"source"`
	TenantID             string    `json:"tenantId"`
	ConnectionName       string    `json:"connectionName"`
	SyncToUserManagement bool      `json:"syncToUserManagement"`
	CreatedAt            time.Time `json:"createdAt"`
	Token                string    `json:"token"`
}

type GroupCreateParams struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Color       string `json:"color,omitempty"`
	Metadata    string `json:"metadata,omitempty"`
}

type GroupUpdateParams struct {
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
	Color       string `json:"color,omitempty"`
	Metadata    string `json:"metadata,omitempty"`
}

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

type ScimRole struct {
	ID          string `json:"id"`
	Key         string `json:"key"`
	Name        string `json:"name"`
	Description string `json:"description"`
	IsDefault   bool   `json:"is_default"`
}

type ScimUser struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

type SCIMGroupsResponse struct {
	Groups []ScimGroup `json:"groups"`
}

type AddRolesToGroupParams struct {
	RoleIds []string `json:"roleIds"`
}

type TenantApiTokenRequest struct {
	Description string            `json:"description"`
	Metadata    map[string]string `json:"metadata"`
	RoleIDs     []string          `json:"roleIds"`
}

type TenantApiTokenResponse struct {
	ClientID        string            `json:"clientId"`
	Description     string            `json:"description"`
	Secret          string            `json:"secret"`
	CreatedByUserId string            `json:"createdByUserId"`
	Metadata        map[string]string `json:"metadata"`
	CreatedAt       time.Time         `json:"createdAt"`
	RoleIDs         []string          `json:"roleIds"`
}

// App struct to hold dependencies
type App struct {
	Router *mux.Router
	Store  *DataStore
	Logger *log.Logger
}

// DataStore holds all the in-memory data
type DataStore struct {
	Mu                 sync.RWMutex
	AppPasswords       map[string]AppPassword
	TenantAppPasswords map[string]TenantApiTokenResponse
	Users              map[string]User
	SSOConfigs         map[string]SSOConfig
	ScimConfigurations map[string]SCIM2Configuration
	Groups             map[string]ScimGroup
}

// Main function to start the server
func main() {
	app := &App{
		Router: mux.NewRouter(),
		Store:  newDataStore(),
		Logger: log.New(os.Stdout, "MOCK-SERVICE: ", log.LstdFlags),
	}

	app.routes()

	app.Logger.Println("Mock Frontegg server is running at http://localhost:3000")
	log.Fatal(http.ListenAndServe(":3000", app.Router))
}

// DataStore constructor to initialize the in-memory data
func newDataStore() *DataStore {
	return &DataStore{
		AppPasswords:       make(map[string]AppPassword),
		TenantAppPasswords: make(map[string]TenantApiTokenResponse),
		Users:              make(map[string]User),
		SSOConfigs:         make(map[string]SSOConfig),
		ScimConfigurations: make(map[string]SCIM2Configuration),
		Groups:             make(map[string]ScimGroup),
	}
}

// Routes setup to handle different endpoints
func (app *App) routes() {
	app.Router.HandleFunc("/identity/resources/auth/v1/api-token", app.handleTokenRequest).Methods("POST")
	app.Router.HandleFunc("/identity/resources/users/api-tokens/v1", app.handleAppPasswords).Methods("GET", "POST")
	app.Router.HandleFunc("/identity/resources/users/api-tokens/v1/{id}", app.handleAppPasswordsDelete).Methods("DELETE")
	app.Router.HandleFunc("/identity/resources/tenants/api-tokens/v1", app.handleTenantAppPasswords).Methods("GET", "POST")
	app.Router.HandleFunc("/identity/resources/tenants/api-tokens/v1/{id}", app.handleTenantAppPasswordsDelete).Methods("DELETE")
	app.Router.HandleFunc("/identity/resources/users/v1/{id}", app.handleUserRequest).Methods("GET", "DELETE")
	app.Router.HandleFunc("/identity/resources/users/v2", app.handleUserRequest).Methods("POST")
	app.Router.HandleFunc("/identity/resources/roles/v2", app.handleRolesRequest).Methods("GET")
	app.Router.HandleFunc("/identity/resources/users/v3", app.handleUserV3Request).Methods("GET")
	app.Router.HandleFunc("/frontegg/team/resources/sso/v1/configurations", app.handleSSOConfigRequest).Methods("GET", "POST")
	app.Router.HandleFunc("/frontegg/team/resources/sso/v1/configurations/{id}", app.handleSSOConfigAndDomainRequest).Methods("GET", "PATCH", "DELETE")
	app.Router.HandleFunc("/frontegg/team/resources/sso/v1/configurations/{id}/domains", app.handleDomainRequests).Methods("GET", "POST")
	app.Router.HandleFunc("/frontegg/team/resources/sso/v1/configurations/{id}/domains/{domainId}", app.handleDomainRequests).Methods("GET", "PATCH", "DELETE")
	app.Router.HandleFunc("/frontegg/team/resources/sso/v1/configurations/{id}/groups", app.handleGroupMappingRequests).Methods("GET", "POST")
	app.Router.HandleFunc("/frontegg/team/resources/sso/v1/configurations/{id}/groups/{groupId}", app.handleGroupMappingRequests).Methods("GET", "PATCH", "DELETE")
	app.Router.HandleFunc("/frontegg/team/resources/sso/v1/configurations/{id}/roles", app.handleDefaultRolesRequests).Methods("GET", "PUT", "DELETE")
	app.Router.HandleFunc("/frontegg/identity/resources/groups/v1", app.handleSCIMGroupsRequest).Methods("GET", "POST")
	app.Router.HandleFunc("/frontegg/identity/resources/groups/v1/{id}", app.handleSCIMGroupsParamRequest).Methods("GET", "PATCH", "DELETE")
	app.Router.HandleFunc("/frontegg/identity/resources/groups/v1/{id}/roles", app.handleAddRolesToGroup).Methods("POST", "DELETE")
	app.Router.HandleFunc("/frontegg/identity/resources/groups/v1/{id}/users", app.handleAddUsersToGroup).Methods("POST", "DELETE")
	app.Router.HandleFunc("/frontegg/directory/resources/v1/configurations/scim2", app.handleSCIM2ConfigurationsRequest).Methods("GET", "POST")
	app.Router.HandleFunc("/frontegg/directory/resources/v1/configurations/scim2/{id}", app.handleSCIMConfigurationByID).Methods("DELETE")
}

// Handler methods
func (app *App) handleTokenRequest(w http.ResponseWriter, r *http.Request) {
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
		sendJSONResponse(w, http.StatusOK, response)
	} else {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
	}
}

func (app *App) handleAppPasswords(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		app.createAppPassword(w, r)
	case http.MethodGet:
		app.listAppPasswords(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (app *App) handleAppPasswordsDelete(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	clientID := vars["id"]

	app.Store.Mu.Lock()
	delete(app.Store.AppPasswords, clientID)
	app.Store.Mu.Unlock()

	w.WriteHeader(http.StatusOK)
}

func (app *App) handleTenantAppPasswords(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		app.createTenantAppPassword(w, r)
	case http.MethodGet:
		app.listTenantAppPasswords(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (app *App) handleTenantAppPasswordsDelete(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	clientID := vars["id"]

	app.Store.Mu.Lock()
	delete(app.Store.TenantAppPasswords, clientID)
	app.Store.Mu.Unlock()

	w.WriteHeader(http.StatusOK)
}

func (app *App) handleUserRequest(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		app.getUser(w, r)
	case http.MethodDelete:
		app.deleteUser(w, r)
	case http.MethodPost:
		app.createUser(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (app *App) handleUserV3Request(w http.ResponseWriter, r *http.Request) {
	app.getUsersV3(w, r)
}

func (app *App) handleRolesRequest(w http.ResponseWriter, r *http.Request) {
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

	sendJSONResponse(w, http.StatusOK, response)
}

func (app *App) handleSSOConfigRequest(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		app.createSSOConfig(w, r)
	case http.MethodGet:
		app.listSSOConfigs(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (app *App) handleSSOConfigAndDomainRequest(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	configID := vars["id"]

	switch r.Method {
	case http.MethodGet:
		app.getSSOConfig(w, r, configID)
	case http.MethodPatch:
		app.updateSSOConfig(w, r, configID)
	case http.MethodDelete:
		app.deleteSSOConfig(w, r, configID)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (app *App) handleDomainRequests(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	configID := vars["id"]
	domainID := vars["domainId"]

	switch r.Method {
	case http.MethodPost:
		app.createDomain(w, r, configID)
	case http.MethodGet:
		if domainID == "" {
			app.listDomains(w, configID)
		} else {
			app.getDomain(w, configID, domainID)
		}
	case http.MethodPatch:
		app.updateDomain(w, r, configID, domainID)
	case http.MethodDelete:
		app.deleteDomain(w, configID, domainID)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (app *App) handleGroupMappingRequests(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	configID := vars["id"]
	groupID := vars["groupId"]

	switch r.Method {
	case http.MethodPost:
		app.createGroupMapping(w, r, configID)
	case http.MethodGet:
		if groupID == "" {
			app.listGroupMappings(w, configID)
		} else {
			app.getGroupMapping(w, configID, groupID)
		}
	case http.MethodPatch:
		app.updateGroupMapping(w, r, configID, groupID)
	case http.MethodDelete:
		app.deleteGroupMapping(w, configID, groupID)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (app *App) handleDefaultRolesRequests(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	configID := vars["id"]

	switch r.Method {
	case http.MethodPut:
		app.setDefaultRoles(w, r, configID)
	case http.MethodGet:
		app.getDefaultRoles(w, configID)
	case http.MethodDelete:
		app.clearDefaultRoles(w, configID)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (app *App) handleSCIMGroupsRequest(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		app.listSCIMGroups(w, r)
	case http.MethodPost:
		app.createScimGroup(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (app *App) handleSCIMGroupsParamRequest(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	groupID := vars["id"]

	switch r.Method {
	case http.MethodGet:
		app.getScimGroupByID(w, r, groupID)
	case http.MethodPatch:
		app.updateScimGroup(w, r, groupID)
	case http.MethodDelete:
		app.deleteScimGroup(w, r, groupID)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (app *App) handleAddRolesToGroup(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	groupID := vars["id"]

	switch r.Method {
	case http.MethodPost:
		app.addRolesToGroup(w, r, groupID)
	case http.MethodDelete:
		app.removeRolesFromGroup(w, r, groupID)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (app *App) handleAddUsersToGroup(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	groupID := vars["id"]

	switch r.Method {
	case http.MethodPost:
		app.addUsersToGroup(w, r, groupID)
	case http.MethodDelete:
		app.removeUsersFromGroup(w, r, groupID)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (app *App) handleSCIM2ConfigurationsRequest(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		app.listSCIMConfigurations(w)
	case http.MethodPost:
		app.createSCIMConfiguration(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (app *App) handleSCIMConfigurationByID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	configID := vars["id"]
	app.deleteSCIMConfiguration(w, r, configID)
}

// Handler functions for different routes

func (app *App) createAppPassword(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Description string `json:"description"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	newAppPassword := AppPassword{
		ClientID:    generateID(),
		Secret:      generateID(),
		Description: req.Description,
		Owner:       "mockOwner",
		CreatedAt:   time.Now(),
	}

	app.Store.Mu.Lock()
	app.Store.AppPasswords[newAppPassword.ClientID] = newAppPassword
	app.Store.Mu.Unlock()

	sendJSONResponse(w, http.StatusCreated, newAppPassword)
}

func (app *App) listAppPasswords(w http.ResponseWriter, r *http.Request) {
	app.Store.Mu.RLock()
	passwords := make([]AppPassword, 0, len(app.Store.AppPasswords))
	for _, password := range app.Store.AppPasswords {
		passwords = append(passwords, password)
	}
	app.Store.Mu.RUnlock()

	sendJSONResponse(w, http.StatusOK, passwords)
}

func (app *App) createTenantAppPassword(w http.ResponseWriter, r *http.Request) {
	var req TenantApiTokenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	newToken := TenantApiTokenResponse{
		ClientID:        generateID(),
		Secret:          generateID(),
		Description:     req.Description,
		CreatedByUserId: "mockUser",
		CreatedAt:       time.Now(),
		Metadata:        req.Metadata,
		RoleIDs:         req.RoleIDs,
	}

	app.Store.Mu.Lock()
	app.Store.TenantAppPasswords[newToken.ClientID] = newToken
	app.Store.Mu.Unlock()

	sendJSONResponse(w, http.StatusCreated, newToken)
}

func (app *App) listTenantAppPasswords(w http.ResponseWriter, r *http.Request) {
	app.Store.Mu.RLock()
	passwords := make([]TenantApiTokenResponse, 0, len(app.Store.TenantAppPasswords))
	for _, password := range app.Store.TenantAppPasswords {
		passwords = append(passwords, password)
	}
	app.Store.Mu.RUnlock()

	sendJSONResponse(w, http.StatusOK, passwords)
}

func (app *App) getUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID := vars["id"]

	app.Store.Mu.RLock()
	user, ok := app.Store.Users[userID]
	app.Store.Mu.RUnlock()

	if !ok {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	sendJSONResponse(w, http.StatusOK, user)
}

func (app *App) deleteUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID := vars["id"]

	app.Store.Mu.Lock()
	delete(app.Store.Users, userID)
	app.Store.Mu.Unlock()

	w.WriteHeader(http.StatusOK)
}

func (app *App) createUser(w http.ResponseWriter, r *http.Request) {
	var newUser struct {
		User
		RoleIDs         []string `json:"roleIds"`
		SkipInviteEmail bool     `json:"skipInviteEmail"`
	}

	if err := json.NewDecoder(r.Body).Decode(&newUser); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	newUser.ID = generateID()

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

	app.Store.Mu.Lock()
	app.Store.Users[newUser.ID] = newUser.User
	app.Store.Mu.Unlock()

	sendJSONResponse(w, http.StatusCreated, newUser.User)
}

func (app *App) getUsersV3(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	email := query.Get("_email")
	limit, _ := strconv.Atoi(query.Get("_limit"))
	offset, _ := strconv.Atoi(query.Get("_offset"))
	ids := query.Get("ids")
	sortBy := query.Get("_sortBy")
	order := query.Get("_order")

	app.Store.Mu.RLock()
	var filteredUsers []User
	for _, user := range app.Store.Users {
		if (email == "" || user.Email == email) &&
			(ids == "" || strings.Contains(ids, user.ID)) {
			filteredUsers = append(filteredUsers, user)
		}
	}
	app.Store.Mu.RUnlock()

	if sortBy != "" {
		sort.Slice(filteredUsers, func(i, j int) bool {
			var less bool
			switch sortBy {
			case "email":
				less = filteredUsers[i].Email < filteredUsers[j].Email
			case "id":
				less = filteredUsers[i].ID < filteredUsers[j].ID
			default:
				return false
			}

			if order == "desc" {
				return !less
			}
			return less
		})
	}

	totalItems := len(filteredUsers)
	if offset >= totalItems {
		filteredUsers = []User{}
	} else {
		end := offset + limit
		if end > totalItems {
			end = totalItems
		}
		filteredUsers = filteredUsers[offset:end]
	}

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

	sendJSONResponse(w, http.StatusOK, response)
}

func (app *App) createSSOConfig(w http.ResponseWriter, r *http.Request) {
	var newConfig SSOConfig
	if err := json.NewDecoder(r.Body).Decode(&newConfig); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	newConfig.Id = generateID()
	newConfig.PublicCertificate = base64.StdEncoding.EncodeToString([]byte(newConfig.PublicCertificate))
	newConfig.CreatedAt = time.Now()
	newConfig.UpdatedAt = newConfig.CreatedAt
	newConfig.GeneratedVerification = generateID()
	newConfig.ConfigMetadata = nil
	newConfig.OverrideActiveTenant = true
	newConfig.SubAccountAccessLimit = 0
	newConfig.SkipEmailDomainValidation = false

	newConfig.RoleIds = newConfig.DefaultRoles.RoleIds

	for i := range newConfig.Groups {
		newConfig.Groups[i].Enabled = true
	}

	app.Store.Mu.Lock()
	app.Store.SSOConfigs[newConfig.Id] = newConfig
	app.Store.Mu.Unlock()

	sendJSONResponse(w, http.StatusCreated, newConfig)
}

func (app *App) listSSOConfigs(w http.ResponseWriter, r *http.Request) {
	app.Store.Mu.RLock()
	configs := make([]SSOConfig, 0, len(app.Store.SSOConfigs))
	for _, config := range app.Store.SSOConfigs {
		if config.Domains == nil {
			config.Domains = []Domain{}
		}
		if config.Groups == nil {
			config.Groups = []GroupMapping{}
		}
		if config.RoleIds == nil {
			config.RoleIds = []string{}
		}
		configs = append(configs, config)
	}
	app.Store.Mu.RUnlock()

	sendJSONResponse(w, http.StatusOK, configs)
}

func (app *App) getSSOConfig(w http.ResponseWriter, r *http.Request, configID string) {
	app.Store.Mu.RLock()
	config, ok := app.Store.SSOConfigs[configID]
	app.Store.Mu.RUnlock()

	if !ok {
		http.Error(w, "SSO configuration not found", http.StatusNotFound)
		return
	}

	sendJSONResponse(w, http.StatusOK, config)
}

func (app *App) updateSSOConfig(w http.ResponseWriter, r *http.Request, configID string) {
	var updatedConfig SSOConfig
	if err := json.NewDecoder(r.Body).Decode(&updatedConfig); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	app.Store.Mu.Lock()
	defer app.Store.Mu.Unlock()

	if _, ok := app.Store.SSOConfigs[configID]; !ok {
		http.Error(w, "SSO configuration not found", http.StatusNotFound)
		return
	}

	updatedConfig.Id = configID
	updatedConfig.PublicCertificate = base64.StdEncoding.EncodeToString([]byte(updatedConfig.PublicCertificate))
	updatedConfig.UpdatedAt = time.Now()
	app.Store.SSOConfigs[configID] = updatedConfig

	sendJSONResponse(w, http.StatusOK, updatedConfig)
}

func (app *App) deleteSSOConfig(w http.ResponseWriter, r *http.Request, configID string) {
	app.Store.Mu.Lock()
	defer app.Store.Mu.Unlock()

	if _, ok := app.Store.SSOConfigs[configID]; !ok {
		http.Error(w, "SSO configuration not found", http.StatusNotFound)
		return
	}

	delete(app.Store.SSOConfigs, configID)
	w.WriteHeader(http.StatusOK)
}

func (app *App) createDomain(w http.ResponseWriter, r *http.Request, ssoConfigID string) {
	var newDomain Domain
	if err := json.NewDecoder(r.Body).Decode(&newDomain); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	app.Store.Mu.Lock()
	defer app.Store.Mu.Unlock()

	config, exists := app.Store.SSOConfigs[ssoConfigID]
	if !exists {
		http.Error(w, "SSO configuration not found", http.StatusNotFound)
		return
	}

	newDomain.ID = generateID()
	newDomain.SsoConfigId = ssoConfigID
	config.Domains = append(config.Domains, newDomain)
	app.Store.SSOConfigs[ssoConfigID] = config

	sendJSONResponse(w, http.StatusCreated, newDomain)
}

func (app *App) listDomains(w http.ResponseWriter, ssoConfigID string) {
	app.Store.Mu.RLock()
	config, exists := app.Store.SSOConfigs[ssoConfigID]
	app.Store.Mu.RUnlock()

	if !exists {
		http.Error(w, "SSO configuration not found", http.StatusNotFound)
		return
	}

	sendJSONResponse(w, http.StatusOK, config.Domains)
}

func (app *App) getDomain(w http.ResponseWriter, ssoConfigID string, domainID string) {
	app.Store.Mu.RLock()
	config, exists := app.Store.SSOConfigs[ssoConfigID]
	app.Store.Mu.RUnlock()

	if !exists {
		http.Error(w, "SSO configuration not found", http.StatusNotFound)
		return
	}

	for _, domain := range config.Domains {
		if domain.ID == domainID {
			sendJSONResponse(w, http.StatusOK, domain)
			return
		}
	}

	http.Error(w, "Domain not found", http.StatusNotFound)
}

func (app *App) updateDomain(w http.ResponseWriter, r *http.Request, ssoConfigID string, domainID string) {
	var updatedDomain Domain
	if err := json.NewDecoder(r.Body).Decode(&updatedDomain); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	app.Store.Mu.Lock()
	defer app.Store.Mu.Unlock()

	config, exists := app.Store.SSOConfigs[ssoConfigID]
	if !exists {
		http.Error(w, "SSO configuration not found", http.StatusNotFound)
		return
	}

	for i, domain := range config.Domains {
		if domain.ID == domainID {
			updatedDomain.ID = domainID
			config.Domains[i] = updatedDomain
			app.Store.SSOConfigs[ssoConfigID] = config
			sendJSONResponse(w, http.StatusOK, updatedDomain)
			return
		}
	}

	http.Error(w, "Domain not found", http.StatusNotFound)
}

func (app *App) deleteDomain(w http.ResponseWriter, ssoConfigID string, domainID string) {
	app.Store.Mu.Lock()
	defer app.Store.Mu.Unlock()

	config, exists := app.Store.SSOConfigs[ssoConfigID]
	if !exists {
		http.Error(w, "SSO configuration not found", http.StatusNotFound)
		return
	}

	for i, domain := range config.Domains {
		if domain.ID == domainID {
			config.Domains = append(config.Domains[:i], config.Domains[i+1:]...)
			app.Store.SSOConfigs[ssoConfigID] = config
			w.WriteHeader(http.StatusOK)
			return
		}
	}

	http.Error(w, "Domain not found", http.StatusNotFound)
}

func (app *App) createGroupMapping(w http.ResponseWriter, r *http.Request, ssoConfigID string) {
	var newGroupMapping GroupMapping
	if err := json.NewDecoder(r.Body).Decode(&newGroupMapping); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	app.Store.Mu.Lock()
	defer app.Store.Mu.Unlock()

	config, exists := app.Store.SSOConfigs[ssoConfigID]
	if !exists {
		http.Error(w, "SSO configuration not found", http.StatusNotFound)
		return
	}

	newGroupMapping.ID = generateID()
	newGroupMapping.SsoConfigId = ssoConfigID
	config.Groups = append(config.Groups, newGroupMapping)
	app.Store.SSOConfigs[ssoConfigID] = config

	sendJSONResponse(w, http.StatusCreated, newGroupMapping)
}

func (app *App) listGroupMappings(w http.ResponseWriter, ssoConfigID string) {
	app.Store.Mu.RLock()
	config, exists := app.Store.SSOConfigs[ssoConfigID]
	app.Store.Mu.RUnlock()

	if !exists {
		http.Error(w, "SSO configuration not found", http.StatusNotFound)
		return
	}

	sendJSONResponse(w, http.StatusOK, config.Groups)
}

func (app *App) getGroupMapping(w http.ResponseWriter, ssoConfigID string, groupMappingID string) {
	app.Store.Mu.RLock()
	config, exists := app.Store.SSOConfigs[ssoConfigID]
	app.Store.Mu.RUnlock()

	if !exists {
		http.Error(w, "SSO configuration not found", http.StatusNotFound)
		return
	}

	for _, mapping := range config.Groups {
		if mapping.ID == groupMappingID {
			sendJSONResponse(w, http.StatusOK, mapping)
			return
		}
	}

	http.Error(w, "Group mapping not found", http.StatusNotFound)
}

func (app *App) updateGroupMapping(w http.ResponseWriter, r *http.Request, ssoConfigID string, groupMappingID string) {
	var updatedMapping GroupMapping
	if err := json.NewDecoder(r.Body).Decode(&updatedMapping); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	app.Store.Mu.Lock()
	defer app.Store.Mu.Unlock()

	config, exists := app.Store.SSOConfigs[ssoConfigID]
	if !exists {
		http.Error(w, "SSO configuration not found", http.StatusNotFound)
		return
	}

	for i, mapping := range config.Groups {
		if mapping.ID == groupMappingID {
			updatedMapping.ID = groupMappingID
			config.Groups[i] = updatedMapping
			app.Store.SSOConfigs[ssoConfigID] = config
			sendJSONResponse(w, http.StatusOK, updatedMapping)
			return
		}
	}

	http.Error(w, "Group mapping not found", http.StatusNotFound)
}

func (app *App) deleteGroupMapping(w http.ResponseWriter, ssoConfigID string, groupMappingID string) {
	app.Store.Mu.Lock()
	defer app.Store.Mu.Unlock()

	config, exists := app.Store.SSOConfigs[ssoConfigID]
	if !exists {
		http.Error(w, "SSO configuration not found", http.StatusNotFound)
		return
	}

	for i, mapping := range config.Groups {
		if mapping.ID == groupMappingID {
			config.Groups = append(config.Groups[:i], config.Groups[i+1:]...)
			app.Store.SSOConfigs[ssoConfigID] = config
			w.WriteHeader(http.StatusOK)
			return
		}
	}

	http.Error(w, "Group mapping not found", http.StatusNotFound)
}

func (app *App) setDefaultRoles(w http.ResponseWriter, r *http.Request, ssoConfigID string) {
	var roles DefaultRoles
	if err := json.NewDecoder(r.Body).Decode(&roles); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	app.Store.Mu.Lock()
	defer app.Store.Mu.Unlock()

	config, exists := app.Store.SSOConfigs[ssoConfigID]
	if !exists {
		http.Error(w, "SSO configuration not found", http.StatusNotFound)
		return
	}

	config.DefaultRoles = roles
	config.RoleIds = roles.RoleIds

	app.Store.SSOConfigs[ssoConfigID] = config

	sendJSONResponse(w, http.StatusCreated, roles)
}

func (app *App) getDefaultRoles(w http.ResponseWriter, ssoConfigID string) {
	app.Store.Mu.RLock()
	config, exists := app.Store.SSOConfigs[ssoConfigID]
	app.Store.Mu.RUnlock()

	if !exists {
		http.Error(w, "SSO configuration not found", http.StatusNotFound)
		return
	}

	sendJSONResponse(w, http.StatusOK, config.DefaultRoles)
}

func (app *App) clearDefaultRoles(w http.ResponseWriter, ssoConfigID string) {
	app.Store.Mu.Lock()
	defer app.Store.Mu.Unlock()

	config, exists := app.Store.SSOConfigs[ssoConfigID]
	if !exists {
		http.Error(w, "SSO configuration not found", http.StatusNotFound)
		return
	}

	config.DefaultRoles = DefaultRoles{RoleIds: []string{}}
	app.Store.SSOConfigs[ssoConfigID] = config

	w.WriteHeader(http.StatusOK)
}

func (app *App) listSCIMGroups(w http.ResponseWriter, r *http.Request) {
	app.Store.Mu.RLock()
	allGroups := make([]ScimGroup, 0, len(app.Store.Groups))
	for _, group := range app.Store.Groups {
		allGroups = append(allGroups, group)
	}
	app.Store.Mu.RUnlock()

	sendJSONResponse(w, http.StatusOK, SCIMGroupsResponse{Groups: allGroups})
}

func (app *App) createScimGroup(w http.ResponseWriter, r *http.Request) {
	var params GroupCreateParams
	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	newGroup := ScimGroup{
		ID:          generateID(),
		Name:        params.Name,
		Description: params.Description,
		Metadata:    params.Metadata,
		Roles:       []ScimRole{},
		Users:       []ScimUser{},
	}

	app.Store.Mu.Lock()
	app.Store.Groups[newGroup.ID] = newGroup
	app.Store.Mu.Unlock()

	sendJSONResponse(w, http.StatusCreated, newGroup)
}

func (app *App) getScimGroupByID(w http.ResponseWriter, r *http.Request, groupID string) {
	app.Store.Mu.RLock()
	group, exists := app.Store.Groups[groupID]
	app.Store.Mu.RUnlock()

	if !exists {
		http.Error(w, "Group not found", http.StatusNotFound)
		return
	}

	sendJSONResponse(w, http.StatusOK, group)
}

func (app *App) updateScimGroup(w http.ResponseWriter, r *http.Request, groupID string) {
	var params GroupUpdateParams
	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	app.Store.Mu.Lock()
	group, exists := app.Store.Groups[groupID]
	if !exists {
		app.Store.Mu.Unlock()
		http.Error(w, "Group not found", http.StatusNotFound)
		return
	}

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

	app.Store.Groups[groupID] = group
	app.Store.Mu.Unlock()

	sendJSONResponse(w, http.StatusOK, group)
}

func (app *App) deleteScimGroup(w http.ResponseWriter, r *http.Request, groupID string) {
	app.Store.Mu.Lock()
	_, exists := app.Store.Groups[groupID]
	if !exists {
		app.Store.Mu.Unlock()
		http.Error(w, "Group not found", http.StatusNotFound)
		return
	}

	delete(app.Store.Groups, groupID)
	app.Store.Mu.Unlock()

	w.WriteHeader(http.StatusOK)
}

func (app *App) addRolesToGroup(w http.ResponseWriter, r *http.Request, groupID string) {
	var params AddRolesToGroupParams
	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	app.Store.Mu.Lock()
	group, exists := app.Store.Groups[groupID]
	if !exists {
		app.Store.Mu.Unlock()
		http.Error(w, "Group not found", http.StatusNotFound)
		return
	}

	for _, roleID := range params.RoleIds {
		found := false
		for _, role := range group.Roles {
			if role.ID == roleID {
				found = true
				break
			}
		}

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

	app.Store.Groups[groupID] = group
	app.Store.Mu.Unlock()

	w.WriteHeader(http.StatusCreated)
}

func (app *App) addUsersToGroup(w http.ResponseWriter, r *http.Request, groupID string) {
	var params struct {
		UserIds []string `json:"userIds"`
	}
	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	app.Store.Mu.Lock()
	group, exists := app.Store.Groups[groupID]
	if !exists {
		app.Store.Mu.Unlock()
		http.Error(w, "Group not found", http.StatusNotFound)
		return
	}

	for _, userID := range params.UserIds {
		found := false
		for _, user := range group.Users {
			if user.ID == userID {
				found = true
				break
			}
		}

		if !found {
			group.Users = append(group.Users, ScimUser{ID: userID})
		}
	}

	app.Store.Groups[groupID] = group
	app.Store.Mu.Unlock()

	w.WriteHeader(http.StatusCreated)
}

func (app *App) removeRolesFromGroup(w http.ResponseWriter, r *http.Request, groupID string) {
	var params AddRolesToGroupParams
	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	app.Store.Mu.Lock()
	group, exists := app.Store.Groups[groupID]
	if !exists {
		app.Store.Mu.Unlock()
		http.Error(w, "Group not found", http.StatusNotFound)
		return
	}

	for _, roleID := range params.RoleIds {
		for i, role := range group.Roles {
			if role.ID == roleID {
				group.Roles = append(group.Roles[:i], group.Roles[i+1:]...)
				break
			}
		}
	}

	app.Store.Groups[groupID] = group
	app.Store.Mu.Unlock()

	w.WriteHeader(http.StatusOK)
}

func (app *App) removeUsersFromGroup(w http.ResponseWriter, r *http.Request, groupID string) {
	var params struct {
		UserIds []string `json:"userIds"`
	}
	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	app.Store.Mu.Lock()
	group, exists := app.Store.Groups[groupID]
	if !exists {
		app.Store.Mu.Unlock()
		http.Error(w, "Group not found", http.StatusNotFound)
		return
	}

	for _, userID := range params.UserIds {
		for i, user := range group.Users {
			if user.ID == userID {
				group.Users = append(group.Users[:i], group.Users[i+1:]...)
				break
			}
		}
	}

	app.Store.Groups[groupID] = group
	app.Store.Mu.Unlock()

	w.WriteHeader(http.StatusOK)
}

func (app *App) listSCIMConfigurations(w http.ResponseWriter) {
	app.Store.Mu.RLock()
	configs := make([]SCIM2Configuration, 0, len(app.Store.ScimConfigurations))
	for _, config := range app.Store.ScimConfigurations {
		configs = append(configs, config)
	}
	app.Store.Mu.RUnlock()

	sendJSONResponse(w, http.StatusOK, configs)
}

func (app *App) createSCIMConfiguration(w http.ResponseWriter, r *http.Request) {
	var newConfig SCIM2Configuration
	if err := json.NewDecoder(r.Body).Decode(&newConfig); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	newConfig.ID = generateID()
	newConfig.Token = generateID()
	newConfig.TenantID = "mockTenantID"
	newConfig.CreatedAt = time.Now()

	app.Store.Mu.Lock()
	app.Store.ScimConfigurations[newConfig.ID] = newConfig
	app.Store.Mu.Unlock()

	sendJSONResponse(w, http.StatusCreated, newConfig)
}

func (app *App) deleteSCIMConfiguration(w http.ResponseWriter, r *http.Request, configID string) {
	app.Store.Mu.Lock()
	_, exists := app.Store.ScimConfigurations[configID]
	if !exists {
		app.Store.Mu.Unlock()
		http.Error(w, "Configuration not found", http.StatusNotFound)
		return
	}

	delete(app.Store.ScimConfigurations, configID)
	app.Store.Mu.Unlock()

	w.WriteHeader(http.StatusNoContent)
}

// Helper functions

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

func generateID() string {
	return uuid.New().String()
}

func sendJSONResponse(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}
