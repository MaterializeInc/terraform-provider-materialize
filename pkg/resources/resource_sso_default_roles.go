package resources

import (
	"context"

	"github.com/MaterializeInc/terraform-provider-materialize/pkg/frontegg"
	"github.com/MaterializeInc/terraform-provider-materialize/pkg/utils"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

var SSODefaultRolesSchema = map[string]*schema.Schema{
	"sso_config_id": {
		Type:        schema.TypeString,
		Required:    true,
		Description: "The ID of the associated SSO configuration.",
	},
	"roles": {
		Type:        schema.TypeSet,
		Elem:        &schema.Schema{Type: schema.TypeString},
		Required:    true,
		Description: "Set of default role names for the SSO configuration. These roles will be assigned by default to users who sign up via SSO.",
	},
}

type SSOConfigRolesResponse struct {
	RoleIds []string `json:"roleIds"`
}

func SSODefaultRoles() *schema.Resource {
	return &schema.Resource{
		CreateContext: ssoDefaultRolesCreateOrUpdate,
		ReadContext:   ssoDefaultRolesRead,
		UpdateContext: ssoDefaultRolesCreateOrUpdate,
		DeleteContext: ssoDefaultRolesDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: SSODefaultRolesSchema,

		Description: "The SSO default roles resource allows you to set the default roles for an SSO configuration. These roles will be assigned to users who sign in with SSO.",
	}
}

func ssoDefaultRolesCreateOrUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerMeta, err := utils.GetProviderMeta(meta)
	if err != nil {
		return diag.FromErr(err)
	}
	client := providerMeta.Frontegg

	ssoConfigID := d.Get("sso_config_id").(string)
	roleNames := convertToStringSlice(d.Get("roles").(*schema.Set).List())

	roleMap, err := frontegg.ListSSORoles(ctx, client)
	if err != nil {
		return diag.FromErr(err)
	}

	var roleIDs []string
	for _, roleName := range roleNames {
		if roleID, ok := roleMap[roleName]; ok {
			roleIDs = append(roleIDs, roleID)
		} else {
			return diag.Errorf("role not found: %s", roleName)
		}
	}

	err = frontegg.SetSSODefaultRoles(ctx, client, ssoConfigID, roleIDs)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(ssoConfigID)
	return ssoDefaultRolesRead(ctx, d, meta)
}

func ssoDefaultRolesRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerMeta, err := utils.GetProviderMeta(meta)
	if err != nil {
		return diag.FromErr(err)
	}
	client := providerMeta.Frontegg

	ssoConfigID := d.Id()

	if err := d.Set("sso_config_id", ssoConfigID); err != nil {
		return diag.FromErr(err)
	}

	roleIDs, err := frontegg.GetSSODefaultRoles(ctx, client, ssoConfigID)
	if err != nil {
		return diag.FromErr(err)
	}

	roleMap, err := frontegg.ListSSORoles(ctx, client)
	if err != nil {
		return diag.FromErr(err)
	}

	var roleNames []string
	for _, roleID := range roleIDs {
		for name, id := range roleMap {
			if id == roleID {
				roleNames = append(roleNames, name)
				break
			}
		}
	}

	if err := d.Set("roles", schema.NewSet(schema.HashString, convertToStringInterface(roleNames))); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func ssoDefaultRolesDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerMeta, err := utils.GetProviderMeta(meta)
	if err != nil {
		return diag.FromErr(err)
	}
	client := providerMeta.Frontegg

	ssoConfigID := d.Id()

	err = frontegg.ClearSSODefaultRoles(ctx, client, ssoConfigID)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId("")
	return nil
}

func convertToStringInterface(stringSlice []string) []interface{} {
	var interfaceSlice []interface{}
	for _, str := range stringSlice {
		interfaceSlice = append(interfaceSlice, str)
	}
	return interfaceSlice
}
