package resources

import (
	"context"
	"strings"

	"github.com/MaterializeInc/terraform-provider-materialize/pkg/clients"
	"github.com/MaterializeInc/terraform-provider-materialize/pkg/frontegg"
	"github.com/MaterializeInc/terraform-provider-materialize/pkg/utils"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func User() *schema.Resource {
	return &schema.Resource{
		CreateContext: userCreate,
		ReadContext:   userRead,
		UpdateContext: userUpdate,
		DeleteContext: userDelete,

		Description: `The user resource allows you to invite and delete users in your Materialize organization.`,

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
			"password": {
				Type:        schema.TypeString,
				Optional:    true,
				Sensitive:   true,
				Description: "The password for the user. If not provided, an activation email will be sent.",
			},
			"send_activation_email": {
				Type:     schema.TypeBool,
				Default:  true,
				Optional: true,
				Description: "Whether to send an email either inviting the user to activate their account, " +
					"if the user is new, or inviting the user to join the organization, if the user already " +
					"exists in another organization. Changing this property after the resource is created " +
					"has no effect.",
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
	password := d.Get("password").(string)
	sendActivationEmail := d.Get("send_activation_email").(bool)
	roleNames := convertToStringSlice(d.Get("roles").([]interface{}))

	for _, roleName := range roleNames {
		if roleName != "Member" && roleName != "Admin" {
			return diag.Errorf("invalid role: %s. Roles must be either 'Member' or 'Admin'", roleName)
		}
	}

	roleMap := providerMeta.FronteggRoles

	var roleIDs []string
	for _, roleName := range roleNames {
		if roleID, ok := roleMap[roleName]; ok {
			roleIDs = append(roleIDs, roleID)
		} else {
			return diag.Errorf("role not found: %s", roleName)
		}
	}

	userRequest := frontegg.UserRequest{
		Email:           email,
		RoleIDs:         roleIDs,
		SkipInviteEmail: !sendActivationEmail || password != "",
	}

	// If a password is provided, add it to the request
	if password != "" {
		userRequest.Password = password
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
		if clients.IsNotFoundError(err) {
			// User doesn't exist, remove from state
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	d.Set("email", userResponse.Email)
	// NOTE: we intentionally don't (and can't) read `send_activation_email`
	// back here, as the value only applies during initial creation.
	d.Set("auth_provider", userResponse.Provider)
	d.Set("verified", userResponse.Verified)
	d.Set("metadata", userResponse.Metadata)

	roleNames := make([]string, len(userResponse.Roles))
	for i, role := range userResponse.Roles {
		// Directly trim the "Organization " prefix from the role name.
		roleName := strings.TrimPrefix(role.Name, "Organization ")
		roleNames[i] = roleName
	}
	d.Set("roles", roleNames)

	return nil
}

func userUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerMeta, err := utils.GetProviderMeta(meta)
	if err != nil {
		return diag.FromErr(err)
	}

	client := providerMeta.Frontegg
	userID := d.Id()
	email := d.Get("email").(string)

	if d.HasChange("roles") {
		roleNames := convertToStringSlice(d.Get("roles").([]interface{}))
		roleMap := providerMeta.FronteggRoles

		var roleIDs []string
		for _, roleName := range roleNames {
			if roleID, ok := roleMap[roleName]; ok {
				roleIDs = append(roleIDs, roleID)
			} else {
				return diag.Errorf("role not found: %s", roleName)
			}
		}

		err := frontegg.UpdateUserRoles(ctx, client, userID, email, roleIDs)
		if err != nil {
			return diag.FromErr(err)
		}
	}

	return userRead(ctx, d, meta)
}

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
