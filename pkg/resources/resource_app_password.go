package resources

import (
	"context"
	"sort"
	"time"

	"github.com/MaterializeInc/terraform-provider-materialize/pkg/clients"
	"github.com/MaterializeInc/terraform-provider-materialize/pkg/frontegg"
	"github.com/MaterializeInc/terraform-provider-materialize/pkg/utils"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
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
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "A human-readable name for the app password.",
			},
			"type": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				Default:      "personal",
				ValidateFunc: validation.StringInSlice([]string{"personal", "service"}, false),
				Description:  "The type of the app password: personal or service.",
			},
			"user": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				Description:  "The user to associate with the app password. Only valid with service-type app passwords.",
				ValidateFunc: validateServiceUsername,
			},
			"roles": {
				Type:        schema.TypeList,
				Optional:    true,
				ForceNew:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Description: "The roles to assign to the app password. Allowed values are 'Member' and 'Admin'. Only valid with service-type app passwords.",
			},
			"created_at": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The time at which the app password was created.",
			},
			"secret": {
				Type:      schema.TypeString,
				Computed:  true,
				Sensitive: true,
			},
			"password": {
				Type:        schema.TypeString,
				Computed:    true,
				Sensitive:   true,
				Description: "The value of the app password.",
			},
		},
	}
}

func appPasswordCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerMeta, err := utils.GetProviderMeta(meta)
	if err != nil {
		return diag.FromErr(err)
	}
	client := providerMeta.Frontegg

	name := d.Get("name").(string)
	type_ := d.Get("type").(string)
	user := d.Get("user").(string)
	roles := convertToStringSlice(d.Get("roles").([]interface{}))

	if type_ == "personal" {
		if user != "" {
			return diag.Errorf("user cannot be specified for a personal-type app password")
		}
		if len(roles) != 0 {
			return diag.Errorf("roles cannot be specified for a personal-type app password")
		}

		request := frontegg.UserApiTokenRequest{
			Description: name,
		}

		response, err := frontegg.CreateUserApiToken(ctx, client, request)
		if err != nil {
			return diag.FromErr(err)
		}

		appPassword := clients.ConstructAppPassword(response.ClientID, response.Secret)

		d.SetId(response.ClientID)
		if err := d.Set("created_at", response.CreatedAt.Format(time.RFC3339)); err != nil {
			return diag.FromErr(err)
		}
		if err := d.Set("secret", response.Secret); err != nil {
			return diag.FromErr(err)
		}
		if err := d.Set("password", appPassword); err != nil {
			return diag.FromErr(err)
		}
	} else {
		sort.Strings(roles)

		if user == "" {
			return diag.Errorf("user is required for a service-type app password")
		}
		if len(roles) == 0 {
			return diag.Errorf("at least one role is required for a service-type app password")
		}

		roleMap := providerMeta.FronteggRoles

		var roleIDs []string
		for _, role := range roles {
			if roleID, ok := roleMap[role]; ok {
				roleIDs = append(roleIDs, roleID)
			} else {
				return diag.Errorf("role not found: %s", role)
			}
		}

		request := frontegg.TenantApiTokenRequest{
			Description: name,
			RoleIDs:     roleIDs,
			Metadata:    map[string]string{"user": user},
		}

		response, err := frontegg.CreateTenantApiToken(ctx, client, request)
		if err != nil {
			return diag.FromErr(err)
		}

		appPassword := clients.ConstructAppPassword(response.ClientID, response.Secret)

		d.SetId(response.ClientID)
		if err := d.Set("roles", roles); err != nil {
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
	}

	return nil
}

// appPasswordRead reads the app password resource from the API.
func appPasswordRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerMeta, err := utils.GetProviderMeta(meta)
	if err != nil {
		return diag.FromErr(err)
	}

	client := providerMeta.Frontegg
	id := d.Id()

	type_ := d.Get("type").(string)

	if type_ == "personal" {
		tokens, err := frontegg.ListUserApiTokens(ctx, client)
		if err != nil {
			return diag.FromErr(err)
		}

		token := findUserApiTokenById(tokens, id)
		if token == nil {
			d.SetId("")
			return nil
		}

		// Update the Terraform state with the retrieved values.
		d.Set("name", token.Description)
		d.Set("created_at", token.CreatedAt.Format(time.RFC3339))

		// We don't update secret and password because those fields can only be
		// determined at creation time.
	} else {
		roleMap := providerMeta.FronteggRoles

		roleReverseMap := make(map[string]string)
		for roleName, roleId := range roleMap {
			roleReverseMap[roleId] = roleName
		}

		tokens, err := frontegg.ListTenantApiTokens(ctx, client)
		if err != nil {
			return diag.FromErr(err)
		}

		token := findTenantApiTokenById(tokens, id)
		if token == nil {
			d.SetId("")
			return nil
		}

		var roles []string
		for _, roleID := range token.RoleIDs {
			role, ok := roleReverseMap[roleID]
			if !ok {
				return diag.Errorf("unknown role ID: %s", roleID)
			}
			roles = append(roles, role)
		}
		sort.Strings(roles)

		// Update the Terraform state with the retrieved values.
		d.Set("name", token.Description)
		d.Set("user", token.Metadata["user"])
		d.Set("roles", roles)
		d.Set("created_at", token.CreatedAt.Format(time.RFC3339))

		// We don't update secret and password because those fields can only be
		// determined at creation time.
	}

	return nil
}

func appPasswordDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerMeta, err := utils.GetProviderMeta(meta)
	if err != nil {
		return diag.FromErr(err)
	}
	client := providerMeta.Frontegg

	id := d.Id()
	type_ := d.Get("type").(string)

	if type_ == "personal" {
		err := frontegg.DeleteUserApiToken(ctx, client, id)
		if err != nil {
			return diag.FromErr(err)
		}
	} else {
		err := frontegg.DeleteTenantApiToken(ctx, client, id)
		if err != nil {
			return diag.FromErr(err)
		}
	}

	d.SetId("")
	return nil
}

// Helper function to find a user API token by ID.
func findUserApiTokenById(tokens []frontegg.UserApiTokenResponse, id string) *frontegg.UserApiTokenResponse {
	for _, token := range tokens {
		if token.ClientID == id {
			return &token
		}
	}
	return nil
}

// Helper function to find a user API token by ID.
func findTenantApiTokenById(tokens []frontegg.TenantApiTokenResponse, id string) *frontegg.TenantApiTokenResponse {
	for _, token := range tokens {
		if token.ClientID == id {
			return &token
		}
	}
	return nil
}
