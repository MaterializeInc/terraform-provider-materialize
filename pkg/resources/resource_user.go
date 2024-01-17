package resources

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/MaterializeInc/terraform-provider-materialize/pkg/utils"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func User() *schema.Resource {
	return &schema.Resource{
		CreateContext: userCreate,
		ReadContext:   userRead,
		// UpdateContext: userUpdate,
		DeleteContext: userDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"email": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The email address of the user. This must be unique across all users in the organization.",
			},
			"auth_provider": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The authentication provider for the user.",
			},
			"roles": {
				Type:        schema.TypeList,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Required:    true,
				ForceNew:    true,
				Description: "The roles to assign to the user. Allowed values are 'Member' and 'Admin'.",
			},
			"verified": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"metadata": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

// CreateUserRequest is used to serialize the request body for creating a new user.
type CreateUserRequest struct {
	Email   string   `json:"email"`
	RoleIDs []string `json:"roleIds"`
}

// CreatedUser represents the expected structure of a user creation response.
type CreatedUser struct {
	ID                string `json:"id"`
	Email             string `json:"email"`
	ProfilePictureURL string `json:"profilePictureUrl"`
	Verified          bool   `json:"verified"`
	Metadata          string `json:"metadata"`
	Provider          string `json:"provider"`
}

// userCreate is the Terraform resource create function for a Frontegg user.
func userCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerMeta, err := utils.GetProviderMeta(meta)
	if err != nil {
		return diag.FromErr(err)
	}

	client := providerMeta.Frontegg
	email := d.Get("email").(string)
	roleNames := convertToStringSlice(d.Get("roles").([]interface{}))

	for _, roleName := range roleNames {
		if roleName != "Member" && roleName != "Admin" {
			return diag.Errorf("invalid role: %s. Roles must be either 'Member' or 'Admin'", roleName)
		}
	}

	// Fetch role IDs based on role names.
	roleMap, err := utils.ListRoles(ctx, client)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error fetching roles: %s", err))
	}

	var roleIDs []string
	for _, roleName := range roleNames {
		if roleID, ok := roleMap[roleName]; ok {
			roleIDs = append(roleIDs, roleID)
		} else {
			// Consider failing the process if the role is not found
			return diag.Errorf("role not found: %s", roleName)
		}
	}

	createUserRequest := CreateUserRequest{
		Email:   email,
		RoleIDs: roleIDs,
	}

	requestBody, err := json.Marshal(createUserRequest)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error marshaling create user request: %s", err))
	}

	req, err := http.NewRequestWithContext(ctx, "POST", fmt.Sprintf("%s/identity/resources/users/v2", client.Endpoint), bytes.NewBuffer(requestBody))
	if err != nil {
		return diag.FromErr(fmt.Errorf("error creating request: %s", err))
	}

	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", "Bearer "+client.Token)

	resp, err := client.HTTPClient.Do(req)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error sending request: %s", err))
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return diag.FromErr(fmt.Errorf("error creating user: status %d", resp.StatusCode))
	}

	var createdUser CreatedUser
	if err := json.NewDecoder(resp.Body).Decode(&createdUser); err != nil {
		return diag.FromErr(fmt.Errorf("error decoding response: %s", err))
	}

	d.Set("verified", createdUser.Verified)
	d.Set("metadata", createdUser.Metadata)
	d.Set("auth_provider", createdUser.Provider)
	d.SetId(createdUser.ID)
	return nil
}

func userRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerMeta, err := utils.GetProviderMeta(meta)
	if err != nil {
		return diag.FromErr(err)
	}

	client := providerMeta.Frontegg
	userID := d.Id()

	// Construct the API request
	req, err := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("%s/identity/resources/users/v1/%s", client.Endpoint, userID), nil)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error creating request: %s", err))
	}
	req.Header.Add("Authorization", "Bearer "+client.Token)

	// Send the request to the Frontegg API
	resp, err := client.HTTPClient.Do(req)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error reading user: %s", err))
	}
	defer resp.Body.Close()

	// Check for a successful response
	if resp.StatusCode != http.StatusOK {
		// If the user is not found, remove it from the Terraform state
		if resp.StatusCode == http.StatusNotFound {
			d.SetId("")
			return nil
		}
		return diag.Errorf("API error: %s", resp.Status)
	}

	// Parse the response body
	var user CreatedUser
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return diag.FromErr(fmt.Errorf("error decoding user response: %s", err))
	}

	// Update the Terraform state with the fetched user data
	d.Set("email", user.Email)
	d.Set("verified", user.Verified)
	d.Set("metadata", user.Metadata)
	d.Set("auth_provider", user.Provider)

	return nil
}

// TODO: Add userUpdate function to change user roles

// userDelete is the Terraform resource delete function for a Frontegg user.
func userDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerMeta, err := utils.GetProviderMeta(meta)
	if err != nil {
		return diag.FromErr(err)
	}

	client := providerMeta.Frontegg
	userID := d.Id()

	// Send the request to the Frontegg API to delete the user.
	req, err := http.NewRequestWithContext(ctx, "DELETE", fmt.Sprintf("%s/identity/resources/users/v1/%s", client.Endpoint, userID), nil)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error creating request to delete user: %s", err))
	}
	req.Header.Add("Authorization", "Bearer "+client.Token)

	// Perform the request
	resp, err := client.HTTPClient.Do(req)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error sending request to delete user: %s", err))
	}
	defer resp.Body.Close()

	// Check for a successful response
	if resp.StatusCode != http.StatusOK {
		return diag.FromErr(fmt.Errorf("error deleting user: status %d", resp.StatusCode))
	}

	// Remove the user from the Terraform state
	d.SetId("")
	return nil
}

// convertToStringSlice is a helper function to convert an interface slice to a string slice.
func convertToStringSlice(input []interface{}) []string {
	result := make([]string, len(input))
	for i, v := range input {
		result[i] = v.(string)
	}
	return result
}
