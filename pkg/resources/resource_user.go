package resources

import (
	"context"
	"fmt"
	"strings"

	"github.com/MaterializeInc/terraform-provider-materialize/pkg/frontegg"
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
	roleMap, err := frontegg.ListSSORoles(ctx, client)
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

	userRequest := frontegg.UserRequest{
		Email:   email,
		RoleIDs: roleIDs,
	}

	userResponse, err := frontegg.CreateUser(ctx, client, userRequest)
	if err != nil {
		return diag.Errorf("Failed to create user: %s", err)
	}

	// Set the Terraform resource ID and other properties from the response
	d.SetId(userResponse.ID)
	d.Set("auth_provider", userResponse.Provider)
	d.Set("verified", userResponse.Verified)
	d.Set("metadata", userResponse.Metadata)

	return nil
}

func userRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerMeta, err := utils.GetProviderMeta(meta)
	if err != nil {
		return diag.FromErr(err)
	}

	client := providerMeta.Frontegg
	userID := d.Id()

	userResponse, err := frontegg.ReadUser(ctx, client, userID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	d.Set("email", userResponse.Email)
	d.Set("auth_provider", userResponse.Provider)
	d.Set("verified", userResponse.Verified)
	d.Set("metadata", userResponse.Metadata)

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

	err = frontegg.DeleteUser(ctx, client, userID)
	if err != nil {
		return diag.FromErr(err)
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
