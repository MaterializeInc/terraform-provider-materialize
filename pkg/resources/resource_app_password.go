package resources

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/MaterializeInc/terraform-provider-materialize/pkg/clients"
	"github.com/MaterializeInc/terraform-provider-materialize/pkg/frontegg"
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

	passwords, err := frontegg.ListAppPasswords(ctx, client)
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
func createAppPassword(ctx context.Context, d *schema.ResourceData, client *clients.FronteggClient) (frontegg.AppPasswordResponse, error) {
	var response frontegg.AppPasswordResponse

	createRequest := appPasswordCreateRequest{
		Description: d.Get("name").(string),
	}
	requestBody, err := json.Marshal(createRequest)
	if err != nil {
		return response, err
	}

	resp, err := clients.FronteggRequest(ctx, client, "POST", frontegg.GetAppPasswordApiEndpoint(client, frontegg.ApiTokenPath), requestBody)
	if err != nil {
		return response, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return response, clients.HandleApiError(resp)
	}

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return response, err
	}

	return response, nil
}

// Helper function to delete an app password.
func deleteAppPassword(ctx context.Context, client *clients.FronteggClient, resourceID string) error {
	resp, err := clients.FronteggRequest(ctx, client, "DELETE", frontegg.GetAppPasswordApiEndpoint(client, frontegg.ApiTokenPath, resourceID), nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return clients.HandleApiError(resp)
	}
	return nil
}

// Helper function to find an app password by ID.
func findAppPasswordById(passwords []frontegg.AppPasswordResponse, id string) *frontegg.AppPasswordResponse {
	for _, password := range passwords {
		if password.ClientID == id {
			return &password
		}
	}
	return nil
}
