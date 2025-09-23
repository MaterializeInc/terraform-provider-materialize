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

var SSORoleGroupMappingSchema = map[string]*schema.Schema{
	"sso_config_id": {
		Type:        schema.TypeString,
		Required:    true,
		Description: "The ID of the associated SSO configuration.",
	},
	"group": {
		Type:        schema.TypeString,
		Required:    true,
		Description: "The name of the SSO group.",
	},
	"roles": {
		Type:        schema.TypeSet,
		Elem:        &schema.Schema{Type: schema.TypeString},
		Required:    true,
		Description: "List of role names associated with the group.",
	},
	"enabled": {
		Type:        schema.TypeBool,
		Computed:    true,
		Description: "Whether the group mapping is enabled.",
	},
}

func SSORoleGroupMapping() *schema.Resource {
	return &schema.Resource{
		CreateContext: ssoGroupMappingCreate,
		ReadContext:   ssoGroupMappingRead,
		UpdateContext: ssoGroupMappingUpdate,
		DeleteContext: ssoGroupMappingDelete,

		Importer: &schema.ResourceImporter{
			StateContext: ssoRoleGroupMappingImport,
		},

		Schema: SSORoleGroupMappingSchema,

		Description: "The SSO group role mapping resource allows you to set the roles for an SSO group. This allows you to automatically assign additional roles according to your identity provider groups",
	}
}

func ssoGroupMappingCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerMeta, err := utils.GetProviderMeta(meta)
	if err != nil {
		return diag.FromErr(err)
	}

	// Validate that SSO group mappings are only managed in SaaS mode
	if diags := providerMeta.ValidateSaaSOnly("materialize_sso_group_mapping"); diags.HasError() {
		return diags
	}

	client := providerMeta.Frontegg

	ssoConfigID := d.Get("sso_config_id").(string)
	group := d.Get("group").(string)
	roleNames := convertToStringSlice(d.Get("roles").(*schema.Set).List())

	roleMap := providerMeta.FronteggRoles

	var roleIDs []string
	for _, roleName := range roleNames {
		if roleID, ok := roleMap[roleName]; ok {
			roleIDs = append(roleIDs, roleID)
		} else {
			return diag.Errorf("role not found: %s", roleName)
		}
	}

	groupMapping, err := frontegg.CreateSSOGroupMapping(ctx, client, ssoConfigID, group, roleIDs)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(groupMapping.ID)
	return ssoGroupMappingRead(ctx, d, meta)
}

// ssoGroupMappingRead reads the state of the SSO group role mapping from the API and updates the Terraform resource.
func ssoGroupMappingRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerMeta, err := utils.GetProviderMeta(meta)
	if err != nil {
		return diag.FromErr(err)
	}
	client := providerMeta.Frontegg

	var ssoConfigID, groupID string

	ssoConfigID = d.Get("sso_config_id").(string)
	groupID = d.Id()

	// Fetch group mappings from the API.
	groups, err := frontegg.GetSSOGroupMappings(ctx, client, ssoConfigID)
	if err != nil {
		return diag.FromErr(err)
	}

	roleMap := providerMeta.FronteggRoles

	for _, group := range *groups {
		if group.ID == groupID {
			d.Set("group", group.Group)
			d.Set("roles", convertToStringInterface(group.RoleIds))
			d.Set("enabled", group.Enabled)

			// Convert role IDs to role names
			var roleNames []string
			for _, roleID := range group.RoleIds {
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
	}

	d.SetId("")
	return nil
}

func ssoGroupMappingUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerMeta, err := utils.GetProviderMeta(meta)
	if err != nil {
		return diag.FromErr(err)
	}
	client := providerMeta.Frontegg

	ssoConfigID := d.Get("sso_config_id").(string)
	group := d.Get("group").(string)
	groupID := d.Id()
	roleNames := convertToStringSlice(d.Get("roles").(*schema.Set).List())

	roleMap := providerMeta.FronteggRoles

	var roleIDs []string
	for _, roleName := range roleNames {
		if roleID, ok := roleMap[roleName]; ok {
			roleIDs = append(roleIDs, roleID)
		} else {
			return diag.Errorf("role not found: %s", roleName)
		}
	}

	_, err = frontegg.UpdateSSOGroupMapping(ctx, client, ssoConfigID, groupID, group, roleIDs)
	if err != nil {
		return diag.FromErr(err)
	}

	return ssoGroupMappingRead(ctx, d, meta)
}

func ssoGroupMappingDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerMeta, err := utils.GetProviderMeta(meta)
	if err != nil {
		return diag.FromErr(err)
	}
	client := providerMeta.Frontegg

	ssoConfigID := d.Get("sso_config_id").(string)
	groupID := d.Id()

	err = frontegg.DeleteSSOGroupMapping(ctx, client, ssoConfigID, groupID)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId("")
	return nil
}

// Custom import function for SSORoleGroupMapping
func ssoRoleGroupMappingImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	compositeID := d.Id()
	parts := strings.Split(compositeID, ":")

	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid format of ID (%s), expected ssoConfigID:groupID", compositeID)
	}

	d.Set("sso_config_id", parts[0])
	d.SetId(parts[1])

	diags := ssoGroupMappingRead(ctx, d, meta)
	if diags.HasError() {
		var err error
		for _, d := range diags {
			if d.Severity == diag.Error {
				if err == nil {
					err = fmt.Errorf("%s", d.Summary)
				} else {
					err = fmt.Errorf("%v; %s", err, d.Summary)
				}
			}
		}
		return nil, err
	}

	// If the group ID is not set, return an error
	if d.Id() == "" {
		return nil, fmt.Errorf("group not found")
	}

	return []*schema.ResourceData{d}, nil
}
