package resources

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/MaterializeInc/terraform-provider-materialize/pkg/clients"
	"github.com/MaterializeInc/terraform-provider-materialize/pkg/frontegg"
	"github.com/MaterializeInc/terraform-provider-materialize/pkg/utils"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

var ScimGroupRoleSchema = map[string]*schema.Schema{
	"group_id": {
		Type:        schema.TypeString,
		Required:    true,
		Description: "The ID of the SCIM group.",
	},
	"roles": {
		Type:        schema.TypeSet,
		Elem:        &schema.Schema{Type: schema.TypeString},
		Optional:    true,
		Description: "The set of role names to assign to the SCIM group.",
	},
}

func SCIM2GroupRoles() *schema.Resource {
	return &schema.Resource{
		CreateContext: scimGroupRoleCreate,
		ReadContext:   scimGroupRoleRead,
		UpdateContext: scimGroupRoleUpdate,
		DeleteContext: scimGroupRoleDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: ScimGroupRoleSchema,

		Description: "The materialize_scim_group_role resource allows managing roles within a SCIM group in Frontegg.",
	}
}

func scimGroupRoleCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	groupID := d.Get("group_id").(string)
	roleNames := expandStringSet(d.Get("roles").(*schema.Set))

	providerMeta, err := utils.GetProviderMeta(meta)
	if err != nil {
		return diag.FromErr(err)
	}
	client := providerMeta.Frontegg

	roleIDs, err := getRoleIDsByName(ctx, client, roleNames)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error getting role IDs: %s", err))
	}

	err = frontegg.AddRolesToGroup(ctx, client, groupID, roleIDs)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error adding roles to SCIM group: %s", err))
	}

	d.SetId(groupID)
	return scimGroupRoleRead(ctx, d, meta)
}

func scimGroupRoleRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	groupID := d.Id()
	if groupID == "" {
		return diag.Errorf("group ID is not set")
	}

	providerMeta, err := utils.GetProviderMeta(meta)
	if err != nil {
		return diag.FromErr(err)
	}
	client := providerMeta.Frontegg

	group, err := frontegg.GetSCIMGroupByID(ctx, client, groupID)
	if err != nil {
		d.SetId("")
		return diag.FromErr(fmt.Errorf("error fetching SCIM group: %s", err))
	}

	var roleNames []interface{}
	for _, role := range group.Roles {
		roleName := strings.TrimPrefix(role.Name, "Organization ")
		roleNames = append(roleNames, roleName)
	}

	d.Set("group_id", groupID)
	d.Set("roles", roleNames)

	return nil
}

func scimGroupRoleUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	groupID := d.Get("group_id").(string)
	oldRoleNames := expandStringSet(d.Get("roles").(*schema.Set))

	providerMeta, err := utils.GetProviderMeta(meta)
	if err != nil {
		return diag.FromErr(err)
	}
	client := providerMeta.Frontegg

	// Get the current roles assigned to the group
	group, err := frontegg.GetSCIMGroupByID(ctx, client, groupID)
	if err != nil {
		d.SetId("")
		return diag.FromErr(fmt.Errorf("error fetching SCIM group: %s", err))
	}

	// Determine the role IDs that need to be removed
	var removedRoleIDs []string
	for _, role := range group.Roles {
		roleRemoved := true
		for _, roleName := range oldRoleNames {
			if role.Name == roleName {
				roleRemoved = false
				break
			}
		}
		if roleRemoved {
			removedRoleIDs = append(removedRoleIDs, role.ID)
		}
	}

	// Check if removedRoleIDs is empty
	if len(removedRoleIDs) > 0 {
		// Remove the roles that were removed from the group only if removedRoleIDs is not empty
		err = frontegg.RemoveRolesFromGroup(ctx, client, groupID, removedRoleIDs)
		if err != nil {
			return diag.FromErr(fmt.Errorf("error removing roles from SCIM group: %s", err))
		}
	}

	// Add the new roles to the group
	newRoleNames := expandStringSet(d.Get("roles").(*schema.Set))
	newRoleIDs, err := getRoleIDsByName(ctx, client, newRoleNames)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error getting role IDs: %s", err))
	}

	log.Printf("[DEBUG] Adding roles to SCIM group: %v", newRoleIDs)
	// Check if newRoleIDs is empty
	if len(newRoleIDs) > 0 {
		// Add the new roles to the group only if newRoleIDs is not empty
		err = frontegg.AddRolesToGroup(ctx, client, groupID, newRoleIDs)
		if err != nil {
			return diag.FromErr(fmt.Errorf("error adding roles to SCIM group: %s", err))
		}
	}

	return scimGroupRoleRead(ctx, d, meta)
}

func scimGroupRoleDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	groupID := d.Get("group_id").(string)
	roleNames := expandStringSet(d.Get("roles").(*schema.Set))

	providerMeta, err := utils.GetProviderMeta(meta)
	if err != nil {
		return diag.FromErr(err)
	}
	client := providerMeta.Frontegg

	roleIDs, err := getRoleIDsByName(ctx, client, roleNames)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error getting role IDs: %s", err))
	}

	err = frontegg.RemoveRolesFromGroup(ctx, client, groupID, roleIDs)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error removing roles from SCIM group: %s", err))
	}

	// Forcing deletion by setting an empty ID
	d.SetId("")

	return nil
}

// Helper function to expand a set of role names to a list of role IDs
func getRoleIDsByName(ctx context.Context, client *clients.FronteggClient, roleNames []string) ([]string, error) {
	roleMap, err := frontegg.ListSSORoles(ctx, client)
	if err != nil {
		return nil, err
	}

	var roleIDs []string
	for _, roleName := range roleNames {
		if roleID, ok := roleMap[roleName]; ok {
			roleIDs = append(roleIDs, roleID)
		} else {
			return nil, fmt.Errorf("role not found: %s", roleName)
		}
	}

	return roleIDs, nil
}

// Helper function to convert a schema.Set to a list of strings
func expandStringSet(input *schema.Set) []string {
	var result []string
	for _, v := range input.List() {
		result = append(result, v.(string))
	}
	return result
}
