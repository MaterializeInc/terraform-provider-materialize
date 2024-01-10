package resources

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/MaterializeInc/terraform-provider-materialize/pkg/clients"
	"github.com/MaterializeInc/terraform-provider-materialize/pkg/utils"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func AppPassword() *schema.Resource {
	return &schema.Resource{
		CreateContext: appPasswordCreate,
		ReadContext:   appPasswordRead,
		DeleteContext: appPasswordDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"owner": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"created_at": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"secret": {
				Type:      schema.TypeString,
				Computed:  true,
				Sensitive: true,
			},
			"password": {
				Type:      schema.TypeString,
				Computed:  true,
				Sensitive: true,
			},
		},
	}
}

type appPasswordCreateRequest struct {
	Description string `json:"description"`
}

type appPasswordResponse struct {
	ClientID    string    `json:"clientId"`
	Description string    `json:"description"`
	Owner       string    `json:"owner"`
	CreatedAt   time.Time `json:"created_at"`
	Secret      string    `json:"secret"`
}

const (
	apiTokenPath = "/identity/resources/users/api-tokens/v1"
)

func appPasswordCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerMeta, err := utils.GetProviderMeta(meta)
	if err != nil {
		return diag.FromErr(err)
	}

	// Create the app password using the helper function.
	response, err := createAppPassword(ctx, d, providerMeta.Frontegg)
	if err != nil {
		return diag.FromErr(err)
	}

	clientId := strings.ReplaceAll(response.ClientID, "-", "")
	secret := strings.ReplaceAll(response.Secret, "-", "")
	appPassword := fmt.Sprintf("mzp_%s%s", clientId, secret)

	// Set the Terraform resource ID and state.
	d.SetId(response.ClientID)
	if err := d.Set("name", response.Description); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("created_at", response.CreatedAt.Format(time.RFC3339)); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("secret", response.Secret); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("password", appPassword); err != nil {
		return diag.FromErr(err)
	}
	// TODO: Get the owner from the API as it's not returned in the response.
	// For now, we can either leave this unset or set a default value.
	// d.Set("owner", "some-default-or-fetched-value")

	return nil
}

// appPasswordRead reads the app password resource from the API.
func appPasswordRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerMeta, err := utils.GetProviderMeta(meta)
	if err != nil {
		return diag.FromErr(err)
	}

	client := providerMeta.Frontegg
	resourceID := d.Id()

	passwords, err := listAppPasswords(ctx, client)
	if err != nil {
		return diag.FromErr(err)
	}

	foundPassword := findAppPasswordById(passwords, resourceID)
	if foundPassword == nil {
		d.SetId("")
		return nil
	}

	appPassword := clients.ConstructAppPassword(foundPassword.ClientID, foundPassword.Secret)

	// Update the Terraform state with the retrieved values.
	d.Set("name", foundPassword.Description)
	d.Set("created_at", foundPassword.CreatedAt.Format(time.RFC3339))
	d.Set("secret", foundPassword.Secret)
	d.Set("password", appPassword)
	// TODO: Get the owner from the API as it's not returned in the response.
	// d.Set("owner", foundPassword.Owner)

	return nil
}

func appPasswordDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerMeta, err := utils.GetProviderMeta(meta)
	if err != nil {
		return diag.FromErr(err)
	}

	client := providerMeta.Frontegg
	resourceID := d.Id()

	err = deleteAppPassword(ctx, client, resourceID)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId("")
	return nil
}

// Helper function to create or make app password
func createAppPassword(ctx context.Context, d *schema.ResourceData, client *clients.FronteggClient) (appPasswordResponse, error) {
	var response appPasswordResponse

	createRequest := appPasswordCreateRequest{
		Description: d.Get("name").(string),
	}
	requestBody, err := json.Marshal(createRequest)
	if err != nil {
		return response, err
	}

	resp, err := doRequest(ctx, client, "POST", getApiEndpoint(client, apiTokenPath), requestBody)
	if err != nil {
		return response, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return response, handleApiError(resp)
	}

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return response, err
	}

	return response, nil
}

// Helper function to construct the full API endpoint for app passwords
func getApiEndpoint(client *clients.FronteggClient, resourcePath string, resourceID ...string) string {
	if len(resourceID) > 0 {
		return fmt.Sprintf("%s%s/%s", client.Endpoint, resourcePath, resourceID[0])
	}
	return fmt.Sprintf("%s%s", client.Endpoint, resourcePath)
}

// Helper function to perform HTTP requests
func doRequest(ctx context.Context, client *clients.FronteggClient, method, url string, body []byte) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, method, url, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	req.Header.Add("Content-Type", "application/json")

	return client.HTTPClient.Do(req)
}

// Helper function to handle API errors
func handleApiError(resp *http.Response) error {
	responseBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode == http.StatusNotFound {
		return nil // Resource not found, equivalent to a successful delete
	}
	return fmt.Errorf("API error: %s - %s", resp.Status, string(responseBody))
}

// Helper function to delete an app password.
func deleteAppPassword(ctx context.Context, client *clients.FronteggClient, resourceID string) error {
	resp, err := doRequest(ctx, client, "DELETE", getApiEndpoint(client, apiTokenPath, resourceID), nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return handleApiError(resp)
	}
	return nil
}

// Helper function to find an app password by ID.
func findAppPasswordById(passwords []appPasswordResponse, id string) *appPasswordResponse {
	for _, password := range passwords {
		if password.ClientID == id {
			return &password
		}
	}
	return nil
}

// listAppPasswords fetches a list of app passwords from the API.
func listAppPasswords(ctx context.Context, client *clients.FronteggClient) ([]appPasswordResponse, error) {
	var passwords []appPasswordResponse

	// Construct the request URL
	url := getApiEndpoint(client, apiTokenPath)

	// Create and send the HTTP request
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request failed: %w", err)
	}
	req.Header.Add("Authorization", "Bearer "+client.Token)

	// Execute the request
	resp, err := client.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("executing request failed: %w", err)
	}
	defer resp.Body.Close()

	// Check the response code
	if resp.StatusCode != http.StatusOK {
		return nil, handleApiError(resp)
	}

	// Decode the response body
	responseBody, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		return nil, fmt.Errorf("reading response body failed: %w", readErr)
	}

	if err := json.Unmarshal(responseBody, &passwords); err != nil {
		return nil, fmt.Errorf("decoding response failed: %w", err)
	}

	return passwords, nil
}
